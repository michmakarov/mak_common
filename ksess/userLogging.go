// userLogging histiry 201203 07:30
package ksess

import (
	"fmt"
	"mak_common/kerr"
	"os"
	"time"
)

type userLogRecord struct {
	recId     string
	user_id   string //
	tag       string //
	tp        string //"ws" or "http"
	ip        string //IP address
	port      string //TCP port
	url       string //"<action name>:..." or "/..."
	start     string // a moment of time in const startFormat
	dur       int64
	bytes_in  int64
	bytes_out int64
	done      int64 // 0 - not done;1 -normar; 2 - error
	errMess   string
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
	if usersLog, err = os.Create(sessCP.LogsDir + usersLogFileName); err != nil {
		err = fmt.Errorf("createUsersLog err=%v", err.Error())
		return
	}
	kerr.PrintDebugMsg(false, "DFLAG201204_0638", fmt.Sprintf("createUsersLog: success"))
	return
}

func newUserLogRecord(recId string, user_id string,
	tag string, tp string, ip string, port string, url string, start string) (ld *userLogRecord) {
	ld = &userLogRecord{
		recId:     recId,
		user_id:   user_id,
		tag:       tag,
		tp:        tp,
		ip:        ip,
		port:      port,
		url:       url,
		start:     start,
		dur:       0,
		bytes_in:  0,
		bytes_out: 0,
		done:      0,
	}
	return ld
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
