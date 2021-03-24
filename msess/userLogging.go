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
	start     string // a moment of doing start in timeFormat (see feeler constant)
	user_id   string //
	tag       string //
	addr      string //remote address
	url       string //"<action name>:..." or "/..."
	dur       string // duration of doing
	code      string //http return code
	extraInfo string
}

var usersLog *os.File

//user_id string,	tag string, tp string, ip string, port string, url string, start string, dur int64, bytes_in int64, bytes_out int64
//user_id , tag, tp, ip, port, url, start, dur, bytes_in, bytes_out

func createUsersLog() (err error) {
	var usersLogFileName = "Ulog" + time.Now().Format("20060102_150405") + ".log"

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

	var s = fmt.Sprintf("%v| |%v| |%v| |%v| |%v| |%v| |%v |%v |%v\n",
		ulr.reqNum, ulr.start, ulr.user_id, ulr.tag, ulr.addr, ulr.url, ulr.dur, ulr.code, ulr.extraInfo)

	n, err := usersLog.WriteString(s)
	if err != nil {
		kerr.SysErrPrintf("Writing to user log error=%v", err.Error())
	}
	if n != len(s) {
		kerr.SysErrPrintf("Writing to user log; n=%v; len=%v", n, len(s))
	}
	return
}

//1        2      3        4    5     6    7    8     9
//reqNum, start, user_id, tag, addr, url, dur, code, extraInfo

func newUserLogRecord(reqNum, start, user_id, tag, addr, url, dur, code, extraInfo string) (rec *userLogRecord) {
	*rec = userLogRecord{
		reqNum,
		start,
		user_id,
		tag,
		addr,
		url,
		dur,
		code,
		extraInfo,
	}
	return
}
