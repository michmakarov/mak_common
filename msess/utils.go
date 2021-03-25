// utils
package msess

import (
	"fmt"
	//"mak_common/kerr"

	//"os"
	//"strings"
	"time"
)

//checkUserCredentailsEnv and other functions with suffix "Env" are
//envelopes for a programer's callback functions
func checkUserCredentailsEnv(userLogName, userPassword string) (user_id int, account, errMess string) {
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

		//kerr.PrintDebugMsg(false, "DFLAG201224_07:09", fmt.Sprintf("checkUserCredentailsEnv: before calling"))
		user_id, account, errMess = checkUserCredential(userLogName, userPassword)

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
		return
	}
}

//210101 for func (fl *feelerLogger) getFlrlogMess
func byteSet(value byte, byteNum int) bool {
	var mask byte
	if byteNum < 1 || byteNum > 7 {
		panic(fmt.Sprintf("byteSet: illegal byte number=%v", byteNum))
	}
	switch byteNum {
	case 1:
		mask = 0b00000001
	case 2:
		mask = 0b00000010
	case 3:
		mask = 0b00000100
	case 4:
		mask = 0b00001000
	case 5:
		mask = 0b00010000
	case 6:
		mask = 0b00100000
	case 7:
		mask = 0b01000000
	case 8:
		mask = 0b10000000
	}
	return (value & mask) != 0
}
