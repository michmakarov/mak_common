/*
It is the superstructure of sql package.
*/
package kfb

import (
	"container/list"
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"kot_common/kerr"
	"kot_common/ksess"
	"kot_common/kutils"
)

//commonFirebirdDBs is a colecction of Firebird connections pools
//string key - a connection string; If the corresponding value == nil then the pool is absent and will be created
var commonFirebirdDBs map[string]*sql.DB
var commonFirebirdDBsMtx sync.Mutex

var liveTxs = list.New()
var liveTxsMtx sync.Mutex

//A TxRecord is a wrapper for sql.Tx and may be obtained by call of GetKtx
type TxRecord struct {
	//The connStr is source and receiver of data.
	//That is it is a string for connectiong to a RDBMS as it is specified by sql.Open function.
	//In other hand, it is designation of a sql.DB object (a pool of connections)
	connStr string

	//The tx is the thansaction in which all input/output operations will be made
	tx *sql.Tx

	//MBD
	opened time.Time

	//The ctx is the context in which all input/output operations will be made
	ctx context.Context

	//MBD
	procName string

	//MBD
	commited string

	//queries is the list of io operations (queries) which was appointed for execution under management of this object
	queries *list.List
}

var ktxLogChan = make(chan *ksess.TxLogRecord, 255)
var ktxLogChanCorrapted bool
var notKtxLogging bool

func init() {
	go ktxLogger()
}
func ktxLogger() {
	var tlr *ksess.TxLogRecord
	var err error
	for {
		if len(ktxLogChan) > 250 {
			kerr.SysErrPrintf("The queue of *ksess.TxLogRecord overflowed")
			ktxLogChanCorrapted = true
			return
		} else {
			tlr = <-ktxLogChan
			if err = ksess.InsertTxLogRecord(tlr); err != nil {
				kerr.SysErrPrintf("ksess.InsertTxLogRecord err=%v", err.Error())
			}
		}
	}

}

func insToLog(ktx *TxRecord, commited string) {
	var logRec *ksess.TxLogRecord
	var (
		action string
		userId string
		reqNum string
		ok     bool
	)
	if notKtxLogging || ktxLogChanCorrapted {
		return
	}

	if action, ok = ksess.GetCtxStrPar(ktx.ctx, string(ksess.URLCtxKey)); !ok {
		action = "Unknow action"
	}
	if userId, ok = ksess.GetCtxStrPar(ktx.ctx, string(ksess.UserIdCtxKey)); !ok {
		userId = "Unknow user"
	}
	if reqNum, ok = ksess.GetCtxStrPar(ktx.ctx, string(ksess.NumberCtxKey)); !ok {
		reqNum = "Unknow request number"
	}

	logRec = &ksess.TxLogRecord{
		ktx.connStr,
		ktx.opened.Format(ksess.GetStartFormat()),
		strconv.Itoa(int(time.Since(ktx.opened) / 1000000)),
		ktx.procName,
		action,
		reqNum,
		userId,
		commited,
	}
	if !notKtxLogging && !ktxLogChanCorrapted {
		ktxLogChan <- logRec
	}
}

//GetKtx returns *TxRecord or error if any errors have been occured
//This creates a connection pool (*sql.DB) for given connection string ("connString") if
//it yet not be created early.
//Next this creates a new *TxRecord, insert it the list and return it for using instead *sql.Tx
//It is a responsibility of the programmer to call Rollback or Commit of this object
func GetKtx(connString string, ctx context.Context, procName string) (ktx *TxRecord, err error) {
	var db *sql.DB
	var tx *sql.Tx
	var newRec *TxRecord

	if ctx == nil {
		ctx = context.TODO()
	}

	if db, err = getFirebirdDB(connString); err != nil {
		err = fmt.Errorf("kfb.GetKtx: getFirebirdDB:%v", err.Error())
		return
	}

	if tx, err = db.Begin(); err != nil {
		err = fmt.Errorf("kfb.GetKtx: db.Begin error:%v", err.Error())
		//ksess.SendToGenLog()
		return
	}

	if strings.TrimSpace(procName) == "" {
		procName = "no proc"
	}
	newRec = &TxRecord{connString, tx, time.Now(), ctx, procName, "", nil}

	liveTxsMtx.Lock()
	defer liveTxsMtx.Unlock()
	liveTxs.PushBack(newRec)

	return newRec, nil
}

//Commits ktx.tx and remove ktx from the list
//If there is any error it does nothing
func (ktx *TxRecord) Commit() (err error) {
	var count int
	var commitedEl *list.Element
	var commited string //"yes" or "already done"
	kutils.Printf("(ktx *TxRecord) Commit() HERE; ktx.procName=%v\n", ktx.procName)
	liveTxsMtx.Lock()
	defer liveTxsMtx.Unlock()
	kutils.Printf("(ktx *TxRecord) Commit() before FOR\n")
	for e := liveTxs.Front(); e != nil; e = e.Next() {
		count++
		if e.Value == ktx {
			commited = "yes"
			if err = e.Value.(*TxRecord).tx.Commit(); err != nil {
				if err != sql.ErrTxDone {
					err = fmt.Errorf("ktx.Commit: %v", err.Error())
					//return !!!??? 190411
				} else {
					commited = "already done"
				}
			}
			commitedEl = e
		}
	}
	kutils.Printf("(ktx *TxRecord) Commit() after FOR; count=%v\n", count)
	if commitedEl != nil {
		liveTxs.Remove(commitedEl)
		insToLog(commitedEl.Value.(*TxRecord), commited)
	} else {
		err = fmt.Errorf("ktx.Commit: no such transaction -  %v", ktx)
	}
	kutils.Printf("(ktx *TxRecord) Commit() before return err=%v\n", err)
	return
}

//Rollbacks ktx.tx and remove ktx from the list
//If there is any error it does nothing
func (ktx *TxRecord) Rollback() (err error) {
	var rollbackedEl *list.Element
	var commited string //"yes" or "already done"
	//kutils.Printf("ktx.Rollback() before for; rollbackedEl =%v\n", rollbackedEl)
	liveTxsMtx.Lock()
	defer liveTxsMtx.Unlock()
	for e := liveTxs.Front(); e != nil; e = e.Next() {
		if e.Value == ktx {
			commited = "not"
			//kutils.Printf("ktx.Rollback(): for: e.Value == ktx; e.Value.(*TxRecord).tx =%v\n", e.Value.(*TxRecord).tx)
			if err = e.Value.(*TxRecord).tx.Rollback(); err != nil {
				if err != sql.ErrTxDone {
					err = fmt.Errorf("ktx.Rollback: %v", err.Error())
					//return
				} else {
					commited = "RB already done"
				}
			}
			rollbackedEl = e
		}
	}
	//kutils.Printf("ktx.Rollback() before liveTxs.Remove; rollbackedEl =%v\n", rollbackedEl)
	if rollbackedEl != nil {
		liveTxs.Remove(rollbackedEl)
		insToLog(rollbackedEl.Value.(*TxRecord), commited)
	} else {
		err = fmt.Errorf("ktx.Rollback: no such transaction -  %v", ktx)
	}
	//kutils.Printf("ktx.Rollback() before return; err  =%v\n", err)
	return
}

//Returns true if ktx is in the list
func (ktx *TxRecord) IsOpen() bool {
	liveTxsMtx.Lock()
	defer liveTxsMtx.Unlock()
	for e := liveTxs.Front(); e != nil; e = e.Next() {
		if e.Value == ktx {
			return true
		}
	}
	return false
}

//It is the wrap of *sql.Tx.QueryRow
func (ktx *TxRecord) QueryRow(queryStr string, args ...interface{}) (row *sql.Row, err error) {
	if !ktx.IsOpen() {
		err = fmt.Errorf("ktx is not open; ktx=%v", ktx)
		return
	}
	liveTxsMtx.Lock()
	defer liveTxsMtx.Unlock()
	row = ktx.tx.QueryRow(queryStr, args...)
	return
}

//It is the wrap of *sql.Tx.Query
func (ktx *TxRecord) Query(queryStr string, args ...interface{}) (rows *sql.Rows, err error) {
	if !ktx.IsOpen() {
		err = fmt.Errorf("ktx is not open; ktx=%v", ktx)
		return
	}

	liveTxsMtx.Lock()
	defer liveTxsMtx.Unlock()
	rows, err = ktx.tx.Query(queryStr, args...)
	if err != nil {
		err = fmt.Errorf("ktx.Query err=%v", err.Error())
		return
	}
	return
}

func (ktx *TxRecord) Query_(sync bool, queryStr string, args ...interface{}) (res *QueryResult, err error) {
	var kfbTask = KfbTask{}
	liveTxsMtx.Lock()
	defer liveTxsMtx.Unlock()
	kfbTask.Query(ktx.tx, ktx.ctx, queryStr, args)

	return
}

//It is the wrap of *sql.Tx.QueryRow
func (ktx *TxRecord) Exec(queryStr string, args ...interface{}) (err error) {
	if !ktx.IsOpen() {
		err = fmt.Errorf("ktx is not open; ktx=%v", ktx)
		return
	}
	liveTxsMtx.Lock()
	defer liveTxsMtx.Unlock()
	_, err = ktx.tx.Exec(queryStr, args...)
	return
}

//func setPoolSize(connString string, size int) ( err error)

/*before 190529
//This returns a db (of *sql.DB) that once have been obtained and was kept in a global varisble
// Before the global variable is assigned preceding checking  (db.Ping()) is done
//Futher it does not attempts to check if the pool is not closed and to take actions to assign a new pool
*/
//This returns a db (of *sql.DB) that once have been obtained and kept in a global varisble
// Before the returning a check  ( by db.Ping()) is made. If error have been occured the db is removed from the global variable and
//a message is being written by kerr.SysErrPrintf
func getFirebirdDB(connString string) (db *sql.DB, err error) {
	//kutils.Printf("getFirebirdDB: connString=%v\n", connString)
	var newPool *sql.DB

	commonFirebirdDBsMtx.Lock()
	defer commonFirebirdDBsMtx.Unlock()

	if commonFirebirdDBs == nil {
		commonFirebirdDBs = make(map[string]*sql.DB)
	}
	if commonFirebirdDBs[connString] == nil {
		//kutils.Printf("getFirebirdDB: before newPool; commonFirebirdDBs=%v\n", commonFirebirdDBs)
		if newPool, err = sql.Open("firebirdsql", connString); err != nil {
			err = fmt.Errorf("getFirebirdDB: sql.Open  error=%v", err.Error())
			return
		}
		kerr.PrintDebugMsg(false, "kfb", fmt.Sprintf("newPool.Ping=%v;connString=%v", newPool.Ping(), connString))
		commonFirebirdDBs[connString] = newPool
	}

	db = commonFirebirdDBs[connString]
	//191013 For what is it being done? Now I see the only explanation: to cut silly errors as a bad format connection string
	if err = db.Ping(); err != nil { //190529//191009
		commonFirebirdDBs[connString] = nil
		err = fmt.Errorf("getFirebirdDB: Ping(%v)  error=%v", connString, err.Error())
		kerr.SysErrPrintf("getFirebirdDB (%v): db.Ping()  error=%v", connString, err.Error())
		return
	}

	//kutils.Printf("getFirebirdDB: before return;db=%v\n", db)
	return
}

//ShowFbConns returns content of a system table of  "MON$ATTACHMENTS" of a database identified by first parameter "connString"
//For example: "SYSDBA:1qaz@WSX@77.108.87.134:7070/var/lib/firebird/3.0/data/Z_TEST.FDB"
//The second parameter is new line marker, for example "\n" or "<br>". That is content of the table is outputed line by line with "nl" as line delimiter
func ShowFbConns(connString string, nl string) (res string) {
	var db *sql.DB
	var err error
	if db, err = getFirebirdDB(connString); err != nil {
		res = err.Error()
		return
	}
	if res, err = showFbConns(db, nl); err != nil {
		res = err.Error()
	}

	return
}

func showFbConns(db *sql.DB, nl string) (res string, err error) {
	var tx *sql.Tx
	var queryStr string
	var rows *sql.Rows
	var rowCount int
	var (
		MON_ATTACHMENT_ID   int
		MON_STATE           int
		MON_USER            string
		MON_REMOTE_PROTOCOL interface{}
		MON_REMOTE_ADDRESS  interface{}
		MON_REMOTE_PID      interface{}
		MON_REMOTE_PROCESS  interface{}
		MON_CLIENT_VERSION  interface{}
		MON_REMOTE_VERSION  interface{}
		MON_REMOTE_HOST     interface{}
		MON_REMOTE_OS_USER  interface{}
	)
	/*
		&MON$STATE,
		&MON$USER,
		&MON$REMOTE_PROTOCOL,
		&MON$REMOTE_ADDRESS,
		&MON$REMOTE_PID,
		&MON$REMOTE_PROCESS,
		&MON$CLIENT_VERSION,
		&MON$REMOTE_VERSION,
		&MON$REMOTE_HOST,
		&MON$REMOTE_OS_USER
	*/

	//if err = db.Ping(); err != nil {
	//	err = fmt.Errorf("showFbConns: db.Ping err = %s", err.Error())
	//	return
	//}

	//kutils.Printf("showFbConns: after  db.Ping();dbFileName=<%v>\n", dbFileName)

	if tx, err = db.Begin(); err != nil {
		err = fmt.Errorf("showConns: db.Begin() err = %s", err.Error())
		return
	}
	defer tx.Commit()

	queryStr = "SELECT MON$ATTACHMENT_ID," +
		"MON$STATE, MON$USER, MON$REMOTE_PROTOCOL, MON$REMOTE_ADDRESS," +
		" MON$REMOTE_PID, MON$REMOTE_PROCESS, MON$CLIENT_VERSION, MON$REMOTE_VERSION," +
		"MON$REMOTE_HOST, MON$REMOTE_OS_USER" +
		" FROM MON$ATTACHMENTS ORDER BY MON$ATTACHMENT_ID, MON$STATE, MON$REMOTE_ADDRESS"
	if rows, err = tx.Query(queryStr); err != nil {
		err = fmt.Errorf("showConns: tx.Query err = %s", err.Error())
		return
	}

	for rows.Next() {
		rowCount++
		if err = rows.Scan(&MON_ATTACHMENT_ID, &MON_STATE,
			&MON_USER,
			&MON_REMOTE_PROTOCOL,
			&MON_REMOTE_ADDRESS,
			&MON_REMOTE_PID,
			&MON_REMOTE_PROCESS,
			&MON_CLIENT_VERSION,
			&MON_REMOTE_VERSION,
			&MON_REMOTE_HOST,
			&MON_REMOTE_OS_USER); err != nil {
			err = fmt.Errorf("showConns: rows.Scan err = %s", err.Error())
			return
		}
		res = res + fmt.Sprintf("id==%v; %v; %v; %v; %v; %v%v", MON_ATTACHMENT_ID, MON_STATE,
			MON_USER,
			//MON_REMOTE_PROTOCOL,
			MON_REMOTE_ADDRESS,
			MON_REMOTE_PID,
			//MON_REMOTE_PROCESS,
			//MON_CLIENT_VERSION,
			//MON_REMOTE_VERSION,
			//MON_REMOTE_HOST,
			MON_REMOTE_OS_USER,
			nl)
	}
	if err = rows.Err(); err != nil {
		err = fmt.Errorf("showConns: rows.Scan (after for) err = %s", err.Error())
		return
	}
	res = fmt.Sprintf("total ==%v--------------------%v", rowCount, nl) + res
	return
}

// func showPoolStats(db *sql.DB, nl string) (res string) {
// 	var stats = db.Stats()
// 	res = fmt.Sprintf("MaxOpenConnections=%v, %v, Idles=%v, %v, InUses=%v, %v,OpenConnectionss=%v, %v, WaitCounts=%v, %v, WaitDurations=%v, %v, MaxIdleCloseds=%v, %v,MaxLifetimeCloseds=%v",
// 		stats.MaxOpenConnections, nl, stats.Idle, nl, stats.InUse, nl, stats.OpenConnections, nl, stats.WaitCount, nl, stats.WaitDuration, nl, stats.MaxIdleClosed, nl, stats.MaxLifetimeClosed)
//
// 	return
// }

//NotDoneTx returns a list of not done at very moment transactions.
//Strings represented transaction are divided from each other by parameter "nl' (new line)
func NotDoneTx(nl string) (s string) {
	liveTxsMtx.Lock()
	defer liveTxsMtx.Unlock()
	s = fmt.Sprintf("Not done transaction(%v)-----", liveTxs.Len()) + nl
	for e := liveTxs.Front(); e != nil; e = e.Next() {
		s = s + fmt.Sprintf("opend==%v;", e.Value.(*TxRecord).opened.Format("20060102_150405"))
		s = s + fmt.Sprintf("proc==%v;", e.Value.(*TxRecord).procName)
		s = s + fmt.Sprintf("Request number==%v", e.Value.(*TxRecord).ctx.Value(ksess.NumberCtxKey))
		s = s + fmt.Sprintf("Request URL==%v", e.Value.(*TxRecord).ctx.Value(ksess.URLCtxKey))
		s = s + fmt.Sprintf("UserId==%v", e.Value.(*TxRecord).ctx.Value(ksess.UserIdCtxKey)) + nl
	}
	s = s + "-------------------------------" + nl
	return

}

//190731 prepositions for Ilnur

//Предполагается, что драйвер firebirdsql успешно подсоединен где-то, то есть sql.Open("firebirdsql", connString) срабатывает без ошибки
//Строка подключения (connString) далее является также идентификатором удаленной СУБД, из которой извлекается информация

//Этот кэш содержит результаты успешных запросов sql.Query
//Первый ключ - идентификатор базы (строка подключения)
//Второй ключ получается  из cacheKey применением json.Marshal
//То есть queryCache содержит указатель на значение sql.Rows для SQL запроса и его параметров
var queryCache map[string]map[string]*sql.Rows

//Значение этого типа преобразуется через json.Marshal в ключ доступа к sql.Rows
type cacheKey struct {
	query  string        // собственно запрос
	params []interface{} //список параметров запроса
}

//Если в queryCache есть *sql.Rows, то он возвращается
//Иначе, делается попытка получить sql.Rows запросом к удаленной СУБД
//Максимальное число попыток и время между ними - см. type CacheOptions
func Query(
	connStr string, //строка подключения - то есть идентификатор базы из которой извлекается информация
	query string, // собственно запрос
	params []interface{}, //список параметров запроса
) (rows *sql.Rows, err error) {
	err = fmt.Errorf("kfb.Query is not realized yet.")
	return
}

//Задает число попыток (и ожидание между ними) получения успешного результата от обращения к внешней СУБД
type CacheOptions struct {
	TryingNum int
	DelayTime time.Duration
}

func SetCacheOptions(co CacheOptions) {

}
