package msess

import (
	//"errors"
	"fmt"

	"mak_common/kerr"
	//"mak_common/kutils"

	//"math/rand"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	//"time"
)

//210330 05:16
func login(w http.ResponseWriter, r *http.Request, a *Agent) {
	var (
		err     error
		ok      bool
		errMess string

		loginFormValue string
		passFormValue  string
		user_id        int
		userId         string
		accountMap     map[string]string
		account        string //210330 05:39 json text that will be sent if no err
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
		panic("msess.login:no checkUserCredential function")
	}
	switch strings.ToUpper(r.Method) {
	case "POST", "GET":
		//kerr.PrintDebugMsg(false, "DFLAG201223_14:45", fmt.Sprintf("loginpost:M=%v contType=%v", r.Method, r.Header.Values("Content-Type")))
		if err = r.ParseForm(); err != nil {
			panic(fmt.Sprintf("msess.login:Error of r.ParseForm(): %v", err.Error()))
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

		account, errMess = checkUserCredentailsEnv(loginFormValue, passFormValue)

		if errMess == "" { //Success of checking credentials
			if err = json.Unmarshal([]byte(account), accountMap); err != nil {
				sendResult(500, fmt.Sprint("But account (login = %v):%v", loginFormValue, err.Error()))
				return
			}
			if userId, ok = accountMap["iserId"]; !ok {
				sendResult(500, fmt.Sprint("But account (login = %v): no key of \"userId\"", loginFormValue))
				return
			}
			if user_id, err = strconv.Atoi(userId); err != nil {
				sendResult(500, fmt.Sprint("But account (login = %v): err of converting \"userId\" to int=%v", loginFormValue, err.Error()))
				return
			}
			if user_id < 0 {
				sendResult(500, fmt.Sprint("But account (login = %v): userId(%v)<0", loginFormValue, user_id))
				return
			}

			if existAgent = whereUser(userId); existAgent != nil {
				if err = unregAgent(existAgent); err != nil {
					mess := fmt.Sprintf("err of unregAgent=%v; existAgent=%v", err.Error(), existAgent)
					kerr.SysErrPrintf(mess)
					sendResult(500, mess)
					return
				}
			}

			a.UserId = userId //210330 12:47
			assignUser(a)
			sendResult(200, account)
			return

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
	var err error
	var mess string
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
		return
	}

	switch r.Method {
	case "GET":
		//count := GetCountOfPerformingChoresOfUser(strconv.Itoa(cookData.UserID))
		if err = unregAgent(a); err != nil { //the session does not exist
			sendResult(400, fmt.Sprintf("error=%v", err.Error()))
			return
		}
	default:
		mess = fmt.Sprintf("allowed only GET method")
		sendResult(400, mess)
		return
	}

	mess = fmt.Sprintf("Agent successfully unregistered; tag=%v; user=%v", a.Tag, a.UserId)
	sendResult(200, mess)
	return
}
