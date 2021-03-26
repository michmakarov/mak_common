package msess

import (
	//"errors"
	"fmt"

	"mak_common/kerr"
	//"mak_common/kutils"

	//"math/rand"
	"net/http"
	"strconv"
	"strings"
	//"time"
)

//

func login(w http.ResponseWriter, r *http.Request, a *Agent) {
	var (
		err     error
		errMess string

		loginFormValue string
		passFormValue  string
		user_id        int
		userId         string
		account        string
		existAgent     *Agent
	)
	var sendResult = func(code int, mess string) {
		mess = "Authorisation: " + mess
		w.Header().Add("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(code)
		w.Write([]byte(mess))
	}

	defer func() {
		if rec := recover(); rec != nil {
			mess := kerr.GetRecoverErrorText(rec)
			mess = "Authorisation (login function problem): " + mess
			kerr.SysErrPrintln(mess)

			w.Header().Add("Content-Type", "text/plain; charset=utf-8")
			w.WriteHeader(500)
			w.Write([]byte(mess))
		}
	}()

	if !MsessRuns() {
		panic("The msess framework has not been run")
	}

	if checkUserCredential == nil {
		panic("no checkUserCredential function")
	}
	switch strings.ToUpper(r.Method) {
	case "POST", "GET":
		//kerr.PrintDebugMsg(false, "DFLAG201223_14:45", fmt.Sprintf("loginpost:M=%v contType=%v", r.Method, r.Header.Values("Content-Type")))
		if err = r.ParseForm(); err != nil {
			panic(fmt.Sprintf("Error of r.ParseForm(): %v", err.Error()))
		}

		if a.UserId != "" {
			sendResult(400, fmt.Sprintf("The /logout must be sent; agent %v already has user %v", a.Tag, a.UserId))
			return
		}

		loginFormValue = r.FormValue("login")
		if loginFormValue == "" {
			sendResult(400, fmt.Sprint("not \"login\" field "))
			return
		}
		passFormValue = r.FormValue("password")
		if passFormValue == "" {
			sendResult(400, fmt.Sprint("not \"password\" field "))
			return
		}

		user_id, account, errMess = checkUserCredentailsEnv(loginFormValue, passFormValue)

		if errMess == "" { //Success of checking credentials
			if user_id < 0 {
				sendResult(400, fmt.Sprint("user_id = %v < 0 (login = %v)", user_id, loginFormValue))
				return
			} else {
				userId = strconv.Itoa(user_id)
			}

			if existAgent = whereUser(userId); existAgent != nil {
				if err = unregAgent(existAgent); err != nil {
					mess := fmt.Sprintf("err of unregAgent=%v; existAgent=%v", err.Error(), existAgent)
					kerr.SysErrPrintf(mess)
					sendResult(500, mess)
					return
				}
			}

			assignUser(a, userId)

		} else { //errMess != ""
			sendResult(400, fmt.Sprintf("checking credentials of user %v is fault with message %v", loginFormValue, errMess))
			return
		}
	default:
		sendResult(400, fmt.Sprintf("Only POST or GET methods are allowed, not %v", r.Method))
	} //switch r.Method

} //login

//210326 17:19
func logout(w http.ResponseWriter, r *http.Request, a *Agent) {
	var sendResult = func(code int, mess string) { //see 201209 06:48 note
		mess = "Exit from session: " + mess
		w.Header().Add("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(code)
		w.Write([]byte(mess))
	}

	defer func() { //see 201209 06:48 note
		if rec := recover(); rec != nil {
			mess := kerr.GetRecoverErrorText(rec)
			mess = "Exit from session (logout function problem): " + mess
			kerr.SysErrPrintln(mess)

			w.Header().Add("Content-Type", "text/plain; charset=utf-8")
			w.WriteHeader(500)
			w.Write([]byte(mess))
		}
	}()

	if !MsessRuns() {
		panic("The msess system is not run")
	}

	if a.UserId == "" { //the session does not exist
		sendResult(400, fmt.Sprintf("A user was not assignes to; tag=%v", a.Tag))
	}

	switch r.Method {
	case "GET":
		count := GetCountOfPerformingChoresOfUser(strconv.Itoa(cookData.UserID))
		hub.unregisterSess(cookData.UserID)
		clearSession(w)
		clearLogErrCookie(w)
		if count > 0 {
			mess = fmt.Sprintf("The session of %v has been ended\n", cookData.UserID)
			mess = mess + fmt.Sprintf("%v requests have been remaining in pergorming", count)
			sendResult(200, mess)
		} else {
			mess = fmt.Sprintf("The session of %v has been ended\n", cookData.UserID)
			mess = mess + fmt.Sprintf("no requests have been remaining in pergorming")
			sendResult(200, mess)
		}
	default:
		mess = fmt.Sprintf("allowed only GET method")
		sendResult(400, mess)
	}

}
