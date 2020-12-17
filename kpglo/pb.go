//pb.go
package kpglo

import (
	"bytes"
	"context"
	"crypto/sha1"
	"math/rand"

	//"crypto/sha1"
	"fmt"
	"io"
	"strings"

	//"mime/multipart"
	"time"

	"kot_common/kerr"

	pg "github.com/go-pg/pg"
)

//pgDB is nil or it points to an active connection pool established by function SetPGConPool(cD pgloConData) (err error)
//Here it is presumption that if the pool was established there is not circumstances that lead to necessary to reestablish it!
//var pgDB *pg.DB
var pgDB *pg.DB

var activeConnData *ConnData

type ConnData struct {
	// TCP host:port
	Addr string

	User     string
	Password string
	Database string

	// ApplicationName is the application name. Used in logs on Pg side.
	// Only available from pg-9.0.
	ApplicationName string

	// Maximum number of retries before giving up.
	// If it is less than 3 it is assigned with 3
	MaxRetries int
}

//GetActivePoolConnData returs a string representation of a connection data (type ConnData struct) that was used by the SetPGConPool.
func GetActivePoolConnData() string {
	if pgDB == nil {
		return fmt.Sprintf("The pool was not established yet")
	}
	return fmt.Sprintf("ApplicationName=%v, Addr=%v, Database=%v, User=%v, MaxRetries=%v", activeConnData.ApplicationName, activeConnData.Addr, activeConnData.Database, activeConnData.User, activeConnData.MaxRetries)
}

//SetPGConPool creates a new connection pool (pg.DB) and if through it a successful query has not been fulfilled the function
// returns the private global variable (pgDB *pg.DB) to nil and returns an error.
//An error also is returned if initially pgDB != nil
//Bised this it saves the connection data (parameter cD) that can be obtained by GetActivePoolConnData()
func SetPGConPool(cD ConnData) (err error) {
	var conData *pg.Options
	var n int
	if pgDB != nil {
		err = fmt.Errorf("kpglo.SetPGConData: connection options already have been set -(%v)", *conData)
		return
	}

	if cD.MaxRetries < 3 {
		cD.MaxRetries = 3
	}

	conData = new(pg.Options)
	conData.Addr = cD.Addr
	conData.Database = cD.Database
	conData.Password = cD.Password
	conData.User = cD.User
	conData.MaxRetries = 3 //cD.MaxRetries

	pgDB = pg.Connect(conData)

	if _, err = pgDB.QueryOne(pg.Scan(&n), "SELECT 110+1"); err != nil {
		err = fmt.Errorf("kpglo.SetPGConData: SELECT 110+1 Error = %v", err.Error())
		pgDB = nil
		activeConnData = nil
		return
	}

	activeConnData = &cD

	//kerr.PrintDebugMsg(false, "pglo", fmt.Sprintf("kpglo.SetPGConPool successful, pgDB=%+v", pgDB))

	return
}

func GetPGPool() (db *pg.DB) {
	return pgDB
}

//GetPGTx returns the result of pgDB.Begin(), where the pgDB is a global variable established by SetPGConPool
func GetPGTx() (pgTx *pg.Tx, err error) {

	defer func() {
		if rec := recover(); rec != nil {
			err = kerr.GetRecoverError(rec)
			err = fmt.Errorf("GetPGTx recover err: %v", err.Error())
			//pgTx = nil
		}
	}()
	//kerr.PrintDebugMsg(false, "pglo", fmt.Sprintf("kpglo.GetPGTx() HERE pgDB=%v", pgDB))

	if pgTx, err = pgDB.Begin(); err != nil {
		err = fmt.Errorf("kpglo.GetPGTx: pgDB.Begin() Error = %v", err.Error())
		//kerr.PrintDebugMsg(false, "pglo", fmt.Sprintf("kpglo.GetPGTx() pgDB.Begin() err=%v", err.Error()))
		//pgTx = nil
		return
	}

	//kerr.PrintDebugMsg(false, "pglo", "kpglo.GetPGTx() successful")

	return
}

//LoExists answers is paticular large object is exist or is not
//It attempts to open the object for reading. If the atempt is not successful and there is not any error it returns true
//190611 it was written. But the question remains not answered: why *pg.Tx
func LoExists(tx *pg.Tx, loid int64) (existence bool, err error) {
	//from libpq-fs.h (C:\Program Files\PostgreSQL\9.3\include\libpq)
	//#define INV_WRITE		0x00020000
	//#define INV_READ		0x00040000
	var (
		query string
		//loDescr int
		//mode    = 0x00040000
		count int
	)
	defer func() {
		if rec := recover(); rec != nil {
			existence = false
			err = fmt.Errorf("kpglo.ExistsLo:unexpected (panic) error, loid=%v; rec=%v", loid, rec)
		}
	}()

	//query = "SELECT lo_open(?,?) AS loDescr"
	//_, err = tx.QueryOne(pg.Scan(&loDescr), query, loid, mode)
	query = "SELECT count(oid) from pg_largeobject_metadata where oid = ?"
	_, err = tx.QueryOne(pg.Scan(&count), query, loid)
	if err != nil {
		err = fmt.Errorf("ExistsLo:tx.QueryOne error=%v", err.Error())
		existence = false
		return
	}
	if count != 0 {
		existence = true
	}
	//kerr.PrintDebugMsg(false, "pglo", fmt.Sprintf("LoExists: before return loDescr=%v", loDescr))
	kerr.PrintDebugMsg(false, "pglo", fmt.Sprintf("LoExists: before return count=%v", count))
	return
}

//SaveAsLo saves a content of (f io.Reader) as a new postgresql large object.
//That is it creates a new object and fills it from f (f io.Reader)
//If (checkHash bool)=true it reads newly created object and checks it against hash (hash []byte)
//If the check is not successful the transaction is rollbacked
func SaveAsLo(f io.Reader, ctx context.Context, tx *pg.Tx, r_buffer_size int, hash []byte, checkHash bool) (pr PerformingReport) {
	var (
		err                 error
		loid                int64
		buff                = make([]byte, r_buffer_size)
		ioerr               error
		countOfEmptyReading int
		loDescr             int
		n, totalN           int
		//buffMBD             bytes.Buffer
	)
	pr.Beg = time.Now()
	pr.Oper = "kpglo.WriteToLo"
	pr.BufferSize = r_buffer_size
	defer func() {
		pr.Dur = time.Since(pr.Beg)
		if loDescr == 0 {
			_ = closeLo(tx, loDescr)
		}
		if rec := recover(); rec != nil {
			pr.Err = fmt.Errorf("SaveAsLo: panic error (rec) = %v", kerr.GetRecoverErrorText(rec))
		}
	}()

	if loid, err = createLo(tx); err != nil {
		pr.Err = fmt.Errorf("SaveAsLo: createLo error = %v", err.Error())
		return
	}
	pr.Loid = loid
	if loDescr, err = openLo(tx, loid, 0x00020000); err != nil {
		pr.Err = fmt.Errorf("SaveAsLo: openLo error = %v", err.Error())
		return
	}

loop:
	for {
		n = 0
		select {
		case <-ctx.Done():
			pr.Err = fmt.Errorf("Was cancelled")
			return
		default:
			countOfEmptyReading = 0
			for (ioerr == nil) && (n == 0) {
				countOfEmptyReading++
				n, ioerr = f.Read(buff)
				if countOfEmptyReading > MaxEmptyReading {
					pr.Err = fmt.Errorf("Too many empty reading:%v", countOfEmptyReading)
					return
				}
				if n == 0 {
					time.Sleep(time.Millisecond * 10)
				}
			}
			if n == 0 {
				break loop
			}
			buff = buff[:n]
			totalN = totalN + n
			if pr.Err = writeToLo(tx, loDescr, buff); pr.Err != nil {
				return
			}
			pr.Counter++
		} //select
	} //for

	if checkHash {
		if hash == nil || len(hash) == 0 {
			pr.Err = fmt.Errorf("No hash summ")
			return
		}
		var s = sha1.New()
		_ = LoToW(s, context.TODO(), tx, int(loid), r_buffer_size)

		//_ = LoToW(&buffMBD, context.TODO(), tx, int(loid), r_buffer_size)
		//kerr.PrintDebugMsg(false, "pglo", fmt.Sprintf("kpglo.SaveAsLo buffMBD=%v;", string(buffMBD.Bytes())))

		kerr.PrintDebugMsg(false, "pglo", fmt.Sprintf("kpglo.SaveAsLo hash=%v; s.Sum(nil)=%v", hash, s.Sum(nil)))

		if !bytes.Equal(s.Sum(nil), hash) {
			pr.Err = fmt.Errorf("Hashs are not equal")
			return
		}

	}
	return
}

func LoToB(loid int64, hash []byte) (lo []byte, err error) {
	var tx *pg.Tx
	var loExist bool
	var buff bytes.Buffer
	//var buff2 bytes.Buffer
	var hash2 []byte
	var pr PerformingReport

	kerr.PrintDebugMsg(false, "pglo", fmt.Sprintf("kpglo.LoToB HERE  loid=%v;hash=%v", loid, hash))

	if tx, err = GetPGTx(); err != nil {
		err = fmt.Errorf("kpglo.ReadLoToSlice: err (of kpglo.GetPGTx())= %v", err.Error())
		return
	}
	if loExist, err = LoExists(tx, loid); err != nil {
		err = fmt.Errorf("kpglo.ReadLoToSlice: err (of kpglo.LoExists())= %v", err.Error())
		return
	}
	if !loExist {
		err = fmt.Errorf("kpglo.ReadLoToSlice: lo(%v) is not exist", loid)
		return
	}

	if pr = LoToW(&buff, context.TODO(), tx, int(loid), writingChunkSize); pr.Err != nil {
		err = fmt.Errorf("kpglo.ReadLoToSlice: err (of kpglo.LoToW())= %v", pr.Err.Error())
		return
	}
	lo = buff.Bytes()
	//buff2.ReadFrom(&buff)
	//_ = LoToW(&buff2, context.TODO(), tx, int(loid), writingChunkSize)

	//lo = buff2.Bytes()

	if hash == nil {
		pr.Oper = "LoToB without chesking hash;" + pr.Oper
	} else {
		pr.Oper = "LoToB with checking hash:" + pr.Oper
	}

	if hash != nil {
		//var s = sha1.New()
		var s2 = sha1.New()

		//io.Copy(s, &buff)
		io.Copy(s2, &buff)

		//hash = s.Sum(nil)
		hash2 = s2.Sum(nil)
		if !bytes.Equal(hash, hash2) {
			err = fmt.Errorf("LoToB: Hashs are not equal: %v---%v", hash, hash2)
			lo = nil
			return
		}
		kerr.PrintDebugMsg(false, "pglo", fmt.Sprintf("kpglo.LoToB   hash2=%v;hash=%v", hash2, hash))
	}
	return
}

/*
CREATE TABLE lo_metadata(
loid bigint PRIMARY KEY, --'Identifier of the large object. Primary key but not foreign key.'
file_name text,
sha1 bytea,
comment text
);
*/
//
type LoMetadata struct {
	Loid      int64
	File_name string
	Sha1      []byte
	Comment   string
}

func (lm LoMetadata) GetLoFileExt() (ext string) {
	var sfn []string
	sfn = strings.Split(lm.File_name, ".")
	if len(sfn) < 2 {
		return
	}
	ext = sfn[len(sfn)-1]
	return
}

//This is a device for testing "broken pool" problem. See kpglo_190610_4 and ksess_pgf_190418_?, where it is used for testing (func pgloInsDelTest(w http.ResponseWriter, r *http.Request))
//It is surely that this function should not be here among publics functions but it is as it is
func Ins_Del_metadata(tx *pg.Tx, id int) (pr PerformingReport) {
	var (
		queryIns, comment string
		queryDel          string
		liod              int
	)
	pr.Id = id
	pr.Beg = time.Now()
	pr.Oper = "kpglo.Ins_Del_metadata"
	defer func() {
		pr.Dur = time.Since(pr.Beg)
		if rec := recover(); rec != nil {
			pr.Err = fmt.Errorf("kpglo.Ins_Del_metadata: panic error (rec) = %v", kerr.GetRecoverErrorText(rec))
		}
	}()

	queryIns = "INSERT INTO lo_metadata (loid, comment) VALUES (?, ?)"
	queryDel = "DELETE FROM lo_metadata WHERE loid=?"
	liod = rand.Int()
	comment = fmt.Sprintf("Inserted with liod=%v", liod)

	_, pr.Err = tx.Exec(queryIns, liod, comment)
	if pr.Err == nil {
		if _, pr.Err = tx.Exec(queryDel, liod); pr.Err != nil {
			pr.Err = fmt.Errorf("Ins_Del_metadata: DEL err=%v", pr.Err.Error())
			return
		}
	} else {
		pr.Err = fmt.Errorf("Ins_Del_metadata: INS err=%v", pr.Err.Error())
	}
	return

}

func GetLoMetadata(loid int64) (lomd LoMetadata, err error) {

	var tx *pg.Tx
	var loExist bool
	var query string

	if tx, err = GetPGTx(); err != nil {
		err = fmt.Errorf("kpglo.GetLoMetadata: err (of kpglo.GetPGTx())= %v", err.Error())
		return
	}
	if loExist, err = LoExists(tx, loid); err != nil {
		err = fmt.Errorf("kpglo.GetLoMetadata: err (of kpglo.LoExists())= %v", err.Error())
		return
	}
	if !loExist {
		err = fmt.Errorf("kpglo.GetLoMetadata: LO (%v) not exists", loid)
		return
	}

	query = "SELECT loid, file_name, sha1, comment  from lo_metadata WHERE loid=?"
	if _, err = tx.QueryOne(pg.Scan(&lomd.Loid, &lomd.File_name, &lomd.Sha1, &lomd.Comment), query, loid); err != nil {
		if err != pg.ErrNoRows {
			err = fmt.Errorf("kpglo.GetLoMetadata: err (of tx.QueryOne(pg.Scan(&lomd)...)= %v", err.Error())
		} else {
			err = fmt.Errorf("kpglo.GetLoMetadata: for %v there is not a metadata)= %v", loid)
		}

		return
	}

	kerr.PrintDebugMsg(false, "pglo", fmt.Sprintf("kpglo.GetLoMetadata =%v", lomd))

	return

}
