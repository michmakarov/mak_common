// loUtils
package kpglo

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"time"

	"kot_common/kerr"

	"github.com/go-pg/pg"
)

const (
//ProgName = "kpglo"
//Version  = "190610"
//Version  = "190521"
//Version  = "190515"
//Version  = "190502"
//var Version = "180919"
//var Version = "180921"

//VersionState = "developing" // "closed190521" // "closed190515" //"closed190315" //"developing" //"closed190205" //or “developing”
/* The last commit of kot_common:
commit de2adb2caa234c4f57e133636058aae0b07d5b43 (HEAD -> master, origin/master, origin/HEAD)
Author: MichMakarov <michmakarov@gmail.com>
Date:   Tue May 21 19:27:25 2019 +0300
*/
)

var (
	writingChunkSize = 10 * 1000 * 1024
)

func SetWritingChunkSize(size int) {
	if size < 10*1024 {
		writingChunkSize = 10 * 1024
		return
	}
	writingChunkSize = size
}

//func GetVesionInfo() string {
//	return ProgName + "_" + Version + " : " + VersionState
//}

type PerformingReport struct {
	Oper       string
	Err        error
	BufferSize int
	Counter    int
	TotalRead  int64
	Beg        time.Time
	Dur        time.Duration
	Loid       int64
	Id, Id2    int // The performation identifier; Ins_Del_metadata //191102 I do not remember for what is it!
}

//var Version = "180919"
//var Version = "180921"

var MaxEmptyReading = 10

func Info() string {
	return "The packet kpglo here!!!"
}

func (pr PerformingReport) HTML() string {
	return fmt.Sprintf("<p>PerformingReport<br>Oper:%v<br>Beg:%v<br>Dur:%v<br>BuffSize:%v<br>Counter:%v<br>Totel:%v<br>Error:%v</p>",
		pr.Oper, pr.Beg, pr.Dur, pr.BufferSize, pr.Counter, pr.TotalRead,
		pr.Err)
}

func (pr PerformingReport) HTML2() string {
	return fmt.Sprintf("<p>PerformingReport Oper:%v; Beg:%v; Dur:%v; BuffSize:%v; Counter:%v; TotalRead:%v; Error:%v</p>",
		pr.Oper, pr.Beg, pr.Dur, pr.BufferSize, pr.Counter, pr.TotalRead,
		pr.Err)
}

func (pr PerformingReport) InsDelRep(numberInQueue int) string {
	return fmt.Sprintf(" (%v) Id=:%v; Beg=:%v; Dur=:%v; Error:%v\n",
		numberInQueue, pr.Id, pr.Beg, pr.Dur, pr.Err)
}

/*
//for Ins_Del_metadata
func (pr PerformingReport) HTML3() string {
	return fmt.Sprintf("<p>PerformingReport Oper:%v; Beg:%v; Dur:%v; BuffSize:%v; Counter:%v; TotalRead:%v; Error:%v</p>",
		pr.Oper, pr.Beg, pr.Dur, pr.BufferSize, pr.Counter, pr.TotalRead,
		pr.Err)
}
*/

func LoToW(w io.Writer, ctx context.Context, tx *pg.Tx, loid, w_buffer_size int) (pr PerformingReport) {
	var (
		buff    = make([]byte, 0)
		loDescr int
		begin   time.Time
	)
	begin = time.Now()
	defer func() {
		if rec := recover(); rec != nil {
			//fmt.Printf("--M-- loToF recover - 19 rec = %v\n", rec)
			pr.Err = fmt.Errorf("LoToW: loid=%v; rec=%v", loid, rec)
		}
	}()

	pr.Oper = "LoToW"
	pr.Beg = begin
	pr.BufferSize = w_buffer_size

	//fmt.Printf("--M-- LoToW here - 24\n")
	if loDescr, pr.Err = openLo(tx, int64(loid), 0x00040000); pr.Err != nil {
		return
	}

loop:
	for {
		select {
		case <-ctx.Done():
			pr.Err = fmt.Errorf("Was cancelled;loid=%v\n", loid)
			return
		default:
			if buff, pr.Err = readFromLo(tx, loDescr, w_buffer_size); pr.Err != nil {
				return
			}

			if len(buff) == 0 {
				break loop
			}
			if _, pr.Err = w.Write(buff); pr.Err != nil {
				return
			}
			pr.Counter++
			pr.TotalRead = pr.TotalRead + int64(len(buff))
		} //select
	} //for
	pr.Dur = time.Now().Sub(begin)
	//fmt.Printf("--M-- loToF: counter=%v, totalRead=%v, dur=%v\n", counter, totalRead, dur)
	pr.Err = closeLo(tx, loDescr)

	return
}

//BuffToLo делает тоже (то есть, пишет в Postgresql large object), что и MfToLo, но данные берет из памяти (buff []byte)
//Функция появилась в версии 190515.
//Как и аналог (MfToLo) в принципе кривая - четвертого параметра не должно быть (как-то не выдумывается ситуация, где-бы требовалось дописывать в lo)
//Но коль аналог работает уже, то и эта имеет право жить.
//Очедедной commit (после Wed Apr 17 18:05:07 2019 +0300) всей библиотеки kot_common (то есть, сейчасный)
//будет сделан ради этой функции без тестирования, которого, возможно, вовсе не будет, если хрень заработает без вопросов
/* before 190521
func BuffToLo(buff []byte, ctx context.Context, tx *pg.Tx, loid int64) (pr PerformingReport) {
	var (
		//ioerr               error
		//countOfEmptyReading int
		loDescr int
		//n, totalN int
	)
	pr.Beg = time.Now()
	pr.Oper = "kpglo.BuffToLo"
	pr.BufferSize = 10 * 1000 * 1024
	defer func() {
		if rec := recover(); rec != nil {
			pr.Err = fmt.Errorf("LoToW: loid=%v; rec=%v", loid, rec)
		}
	}()

	if loDescr, pr.Err = openLo(tx, loid, 0x00020000); pr.Err != nil {
		return
	}

	if pr.Err = writeToLo(tx, loDescr, buff); pr.Err != nil {
		return
	}

	pr.Err = closeLo(tx, loDescr)
	pr.Dur = time.Since(pr.Beg)

	return
}
*/

func BuffToLo(buff []byte, ctx context.Context, tx *pg.Tx, loid int64) (pr PerformingReport) {
	var (
		loDescr     int
		chunk       []byte
		bytesBuffer *bytes.Buffer
		nibbled     int
		nibbledErr  error
	)
	pr.Beg = time.Now()
	pr.Oper = "kpglo.BuffToLo"
	pr.BufferSize = writingChunkSize

	if len(buff) == 0 {
		pr.Err = fmt.Errorf("kpglo.BuffToLo error; empty buffer")
		return
	}

	defer func() {
		if rec := recover(); rec != nil {
			pr.Err = fmt.Errorf("LoToW: loid=%v; rec=%v", loid, rec)
		}
	}()

	if loDescr, pr.Err = openLo(tx, loid, 0x00020000); pr.Err != nil {
		return
	}

	bytesBuffer = bytes.NewBuffer(buff)
	chunk = make([]byte, writingChunkSize)

	for {
		if nibbled, nibbledErr = bytesBuffer.Read(chunk); nibbledErr != nil {
			if nibbledErr == io.EOF {
				break
			} else {
				pr.Err = nibbledErr
				return
			}
		}

		if pr.Err = writeToLo(tx, loDescr, chunk); pr.Err != nil {
			return
		}
		pr.Counter++
		pr.TotalRead = pr.TotalRead + int64(nibbled)
	}

	pr.Err = closeLo(tx, loDescr)
	pr.Dur = time.Since(pr.Beg)

	return
}

func MfToLo(f multipart.File, ctx context.Context, tx *pg.Tx, r_buffer_size int, loid int64) (pr PerformingReport) {
	var (
		buff                = make([]byte, r_buffer_size)
		ioerr               error
		countOfEmptyReading int
		loDescr             int
		//loid                int64
		n, totalN int
	)
	pr.Beg = time.Now()
	pr.Oper = "kpglo.MfToLo"
	pr.BufferSize = r_buffer_size
	defer func() {
		if rec := recover(); rec != nil {
			//fmt.Printf("--M-- loToF recover - 19 rec = %v\n", rec)
			pr.Err = fmt.Errorf("MfToLo: loid=%v; rec=%v", loid, rec)
		}
	}()

	if loDescr, pr.Err = openLo(tx, loid, 0x00020000); pr.Err != nil {
		return
	}

loop:
	for {
		n = 0
		select {
		case <-ctx.Done():
			pr.Err = fmt.Errorf("Was cancelled")
			fmt.Printf("--M-- MfToLo cancelled pr.Counter=%v, loid=%v;countOfEmptyReading=%v;totalN=%v\n", pr.Counter, loid, countOfEmptyReading, totalN)
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
	//fmt.Printf("--M-- loToF: counter=%v, totalRead=%v, dur=%v\n", counter, totalRead, dur)
	pr.Err = closeLo(tx, loDescr)
	pr.Dur = time.Since(pr.Beg)

	return
}

func DeleteLo(tx *pg.Tx, loid int64) (err error) {
	var (
		query string
		res   int
	)
	query = "SELECT lo_unlink(?) AS res"
	_, err = tx.QueryOne(pg.Scan(&res), query, loid)
	if err != nil {
		return
	}
	if res < 1 {
		err = fmt.Errorf("For loid=%v  SELECT lo_unlink(?) AS res return < 1", loid)
	}
	return
}

//The createLo is a right alternative to the CreateLo.
//The rightness consists mainly in that that this kind of functionality does not be public.
func createLo(tx *pg.Tx) (loid int64, err error) {
	var (
		query string
	)
	query = "SELECT lo_creat(-1) AS loid"
	_, err = tx.QueryOne(pg.Scan(&loid), query)
	return
}

func CreateLo(tx *pg.Tx) (loid int64, err error) {
	var (
		query string
	)
	query = "SELECT lo_creat(-1) AS loid"
	_, err = tx.QueryOne(pg.Scan(&loid), query)
	return
}

func openLo(tx *pg.Tx, loid int64, mode int) (loDescr int, err error) {
	var (
		query string
	)
	defer func() {
		if rec := recover(); rec != nil {
			err = fmt.Errorf("openLo: loid=%v;mode=%v; rec=%v", loid, mode, rec)
		}
	}()

	query = "SELECT lo_open(?,?) AS loDescr"
	_, err = tx.QueryOne(pg.Scan(&loDescr), query, loid, mode)
	if err != nil {
		return
	}
	if loDescr < 0 { //190610_1
		err = fmt.Errorf("openLo:for loid=%v;mode=%v  loDescr (%v) < 0", loid, mode, loDescr)
	}
	return
}

func OpenLo(tx *pg.Tx, loid int64, mode int) (loDescr int, err error) {
	var (
		query string
	)
	defer func() {
		if rec := recover(); rec != nil {
			err = fmt.Errorf("openLo: loid=%v;mode=%v; rec=%v", loid, mode, rec)
		}
	}()

	query = "SELECT lo_open(?,?) AS loDescr"
	_, err = tx.QueryOne(pg.Scan(&loDescr), query, loid, mode)
	if err != nil {
		return
	}
	if loDescr < 0 { //190610_1
		err = fmt.Errorf("openLo:for loid=%v;mode=%v  loDescr (%v) < 0", loid, mode, loDescr)
	}
	return
}

func readFromLo(tx *pg.Tx, loDescr int, buffLen int) (buff []byte, err error) {
	var (
		query string
	)
	query = "SELECT loread(?,?) AS buff"
	_, err = tx.QueryOne(pg.Scan(&buff), query, loDescr, buffLen)
	if err != nil {
		err = fmt.Errorf("readFromLo: tx.QueryOne err = %v ", err.Error())
		return
	}
	return
}

func ReadFromLo(tx *pg.Tx, loDescr int, buffLen int) (buff []byte, err error) {
	var (
		query string
	)
	query = "SELECT loread(?,?) AS buff"
	_, err = tx.QueryOne(pg.Scan(&buff), query, loDescr, buffLen)
	if err != nil {
		err = fmt.Errorf("readFromLo: tx.QueryOne err = %v ", err.Error())
		return
	}
	return
}

func writeToLo(tx *pg.Tx, loDescr int, buff []byte) (err error) {
	var (
		query string
		n     int64
	)
	//fmt.Printf("--M--  writeToLo here loDescr=%v; len(buff)=%v\n", loDescr, len(buff))
	query = "SELECT lowrite(?,?) AS loid"
	_, err = tx.QueryOne(pg.Scan(&n), query, loDescr, buff)
	if err != nil {
		err = fmt.Errorf("writeToLo: tx.QueryOne err = %v ", err.Error())
		return
	}
	if n < 0 {
		err = fmt.Errorf("writeToLo: tx.QueryOne returned %v ", n)
		return
	}
	return
}

func closeLo(tx *pg.Tx, loDescr int) (err error) {
	var (
		query string
		res   int
	)
	query = "SELECT lo_close(?) AS res"
	if _, err = tx.QueryOne(pg.Scan(&res), query, loDescr); err != nil {
		err = fmt.Errorf(" closeLo err=%v", err.Error())
		return
	}
	if res < 0 {
		err = fmt.Errorf(" closeLo not zero result=%v", 0)
		return
	}
	return
}

func GetLoLength(tx *pg.Tx, loid string) (loLen int64, err error) {
	var (
		query string
	)

	kerr.PrintDebugMsg(false, "mainProcessesDownload", fmt.Sprintf("kpglo.GetLoLength: loid=%v", loid))

	query = "select length(lo_get(lmt.oid)) from pg_catalog.pg_largeobject_metadata lmt where lmt.oid=?"
	_, err = tx.QueryOne(pg.Scan(&loLen), query, loid)

	return
}

//191027 Someone somewhen rewrote the GetLoLength function. Why (for what) is not clear at all
//For KSODD Ver=ksodd(pgf_2)191026 : developing; commit_date=b3c43814_27.10.2019
///api/processes/download?file_id=13958&file_name=MOV_2276.mp4&type_card=request&p_id=2&ta_id=0
// the function gives
//"kpglo.GetLoLength:ERROR #54000 large object read request is too large"
//So I restore the function in my old version
//Now I have
//kpglo.GetLoLength:ERROR #42501 permission denied for relation pg_largeobject
func GetLoLength_Mak(tx *pg.Tx, loid string) (loLen int64, err error) {
	var (
		query string
	)
	//query = "SELECT lo_creat(-1) AS loid"
	query = "select sum(length(lo.data)) from pg_largeobject lo where lo.loid=?"
	_, err = tx.QueryOne(pg.Scan(&loLen), query, loid)
	return
}

//20190610
type LoidHashe struct {
	Loid int64
	Hash []byte
}
type GetDescr func(loid int64) string

func CheckLoids(begin, ent time.Time, loids []LoidHashe, getDescr GetDescr, deleteBad bool) {

}
