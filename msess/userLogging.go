// userLogging histiry 201203 07:30
// 210315 19:10 To begin with, /home/mich412/go/src/mak_common/msess/history.txt begins from 210302 12:44
// So, the issue is, as I remember, in that, that in this file must be logged only those requests
//which have passed the feeler and have done by a programmer's handler.
//That is, all those, that have been accepted.
package msess

import (
	"fmt"
	"mak_common/kerr"
	"os"
	"time"
)

type userLogRecord struct {
	reqNum    string //feelerCount
	user_id   string //
	tag       string //
	addr      string //remote address
	url       string //"<action name>:..." or "/..."
	start     string // a moment of doing start in timeFormat (see feeler constant)
	dur       string // duration of doing
	code      string //http return code
	extraInfo string
}

//user_id string,	tag string, tp string, ip string, port string, url string, start string, dur int64, bytes_in int64, bytes_out int64
//user_id , tag, tp, ip, port, url, start, dur, bytes_in, bytes_out

func (ulr userLogRecord) String() string {
	if ulr.done == 0 {
		return fmt.Sprintf("%v(%v) c %v в обработке: %v( началась %v)",
			ulr.user_id, ulr.tp, ulr.ip, ulr.url, ulr.start)
	}
	mks := ulr.dur / 1000
	return fmt.Sprintf("%v(%v) c %v выполнено : %v( %v мкс; начало %v)",
		ulr.user_id, ulr.tp, ulr.ip, ulr.url, mks, ulr.start)
}

func createUsersLog() (err error) {
	var usersLogFileName = "Ulog" + time.Now().Format("20060102_150405") + ".log"
	if err = checkLogsDir(sessCP.LogsDir); err != nil {
		return
	}
	if usersLog, err = os.Create("logs/" + usersLogFileName); err != nil {
		err = fmt.Errorf("createUsersLog err=%v", err.Error())
		return
	}
	//kerr.PrintDebugMsg(false, "DFLAG201204_0638", fmt.Sprintf("createUsersLog: success"))
	return
}

func insertUserLogRecord(ulr *userLogRecord) {
	if usersLog == nil {
		return
	}
	var s = fmt.Sprintf("%v| |%v| |%v| |%v| |%v| |%v| |%v\n", ulr.user_id, ulr.tp, ulr.ip, ulr.port, ulr.url, ulr.start, ulr.dur)
	n, err := usersLog.WriteString(s)
	if err != nil {
		kerr.SysErrPrintf("Writing to user log error=%v", err.Error())
	}
	if n != len(s) {
		kerr.SysErrPrintf("Writing to user log; n=%v; len=%v", n, len(s))
	}
	return
}
