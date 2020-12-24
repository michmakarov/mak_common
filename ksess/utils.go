// utils
package ksess

import (
	"fmt"
	"mak_common/kerr"
	"os"
	"strings"
	"time"
)

//201203 16:29
//A logsDir paramrter must
func checkLogsDir(logsDir string) (err error) {
	if logsDir == "" {
		return
	}
	switch string(os.PathSeparator) {
	case "/":
		if !strings.HasSuffix(logsDir, "/") {
			err = fmt.Errorf("Checking a log dir error: the dir not ended by /")
			return
		}
	case "\\":
		if !strings.HasSuffix(logsDir, "\\") {
			err = fmt.Errorf("Checking a log dir error: the dir not ended by \\")
			return
		}
	default:
		err = fmt.Errorf("Checking a log dir error: Illegal value of the path separator= %v", os.PathSeparator)
		return
	}
	return
}

func checkUserCredentailsEnv(action, userLogName, userPassword string) (user_id int, errMess string) {
	type Result struct {
		user_id int
		errMess string
	}
	var res Result
	var resChan = make(chan Result)
	var exec = func() {
		var user_id int
		var errMess string

		kerr.PrintDebugMsg(false, "DFLAG201224_07:09", fmt.Sprintf("checkUserCredentailsEnv: before calling"))
		user_id, errMess = checkUserCredentails(action, userLogName, userPassword)
		kerr.PrintDebugMsg(false, "DFLAG201224_07:09", fmt.Sprintf("checkUserCredentailsEnv: after calling (%v--%v)", user_id, errMess))

		resChan <- Result{user_id, errMess}
	}

	go exec()

	select {
	case res = <-resChan:
		user_id = res.user_id
		errMess = res.errMess
		//kerr.PrintDebugMsg(false, "DFLAG201224_07:09", fmt.Sprintf("checkUserCredentailsEnv: case res (%v--%v)", user_id, errMess))
		return
	case <-time.After(time.Duration(sessCP.CallBakTimeout) * time.Millisecond):
		errMess = fmt.Sprintf("checkUserCredentails was interrupted by timeout (CallBakTimeout=%v)", sessCP.CallBakTimeout)
		kerr.SysErrPrintln("checkUserCredentailsEnv: interrupted by timeout")
		return
	}
}

//201224 12:29 for loginpost
func isInInts(i int, ints []int) bool {
	for _, val := range ints {
		if i == val {
			return true
		}
	}
	return false
}
