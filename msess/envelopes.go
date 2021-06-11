// envelopes
//210604 17:26 It is for envelope functions that are for isolating (or sandboxing) other functions
package msess

import (
	"fmt"
	"mak_common/kerr"

	"time"
)

//checkUserCredentailsEnv and other functions with suffix "Env" are
//envelopes for a programer's callback functions
func checkUserCredentailsEnv(userLogName, userPassword string) (account, errMess string) {
	type Result struct {
		user_id int
		account string
		errMess string
	}
	var res Result
	var resChan = make(chan Result)
	var exec = func() {
		var user_id int
		var errMess string

		defer func() {
			var rec interface{}
			if rec = recover(); rec != nil {
				errMess = kerr.GetRecoverErrorText(rec)
			}
			resChan <- Result{user_id, account, errMess}
		}()

		//kerr.PrintDebugMsg(false, "DFLAG201224_07:09", fmt.Sprintf("checkUserCredentailsEnv: before calling"))
		//user_id, account, errMess = checkUserCredential(userLogName, userPassword)
		account, errMess = checkUserCredential(userLogName, userPassword)

		//resChan <- Result{user_id, account, errMess}
	}

	go exec()

	select {
	case res = <-resChan:
		//user_id = res.user_id
		account = res.account
		errMess = res.errMess
		//kerr.PrintDebugMsg(false, "DFLAG201224_07:09", fmt.Sprintf("checkUserCredentailsEnv: case res (%v--%v)", user_id, errMess))
		return
	case <-time.After(time.Duration(sessCP.CallBakTimeout) * time.Millisecond):
		errMess = fmt.Sprintf("checkUserCredentails was interrupted by timeout (CallBakTimeout=%v)", sessCP.CallBakTimeout)
		return
	}
}

func checkURLPathEnv(path string) bool {
	var res bool
	var resChan = make(chan bool)
	var errMess string
	var exec = func() {

		defer func() {
			var rec interface{}
			if rec = recover(); rec != nil {
				errMess = kerr.GetRecoverErrorText(rec)
				kerr.SysErrPrintf("checkURLPath panicked with message=%v", errMess)
			}
			resChan <- false
		}()

		resChan <- checkURLPath(path)

	}

	if path == "/" {
		return false
	} //!!!! false for all INTECEPT_REQUESTS

	go exec()

	select {
	case res = <-resChan:
		return res
	case <-time.After(time.Duration(sessCP.CallBakTimeout) * time.Millisecond):
		errMess = fmt.Sprintf("checkURLPath was interrupted by timeout (CallBakTimeout=%v)", sessCP.CallBakTimeout)
		kerr.SysErrPrintln(errMess)
		return false
	}
}
