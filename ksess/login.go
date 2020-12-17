package ksess

import (
	"errors"
	"fmt"

	//"html/template"
	"mak_common/kerr"
	"mak_common/kutils"

	//"math/rand"
	"net/http"
	"strconv"
	//"time"
)

//201209 06:48 The loginpost itself finishes (completes) a request.
//In other words, it SHOULD be the last chain link in working the request.
//What is here the matter? It is an old question : Is it good to give a result through a panic?
//Let's answer NO. It is better to keep flyes and cutlets apart.
//As the consequence: if there is a panic then a returning code is 500 and no else
//If there is not a panic then the sendResult workes.
//_____
/* removed since 180813 until 181229
 */
//181128_developing What does this function presume?
//(1) user_id <0; That is this function is allowed to call only if the request is come from not registered user.
// But what will be if it is not so?
//Some strange behaviour may be. For example, the server may say "bad password" although a registration was already done
//181128_developing (181231 ) What does this function admit
//(2) For Request (r). It must be "POST" and have Content-Type of application/x-www-form-urlencoded
// The last, as it seems, really is necessary if you want to pass an initial data for otherwise thiis data is be regarded absent in the request
func loginpost(w http.ResponseWriter, r *http.Request) {
	var (
		cookData          sessCookieData
		err               error
		errMess           string
		loginFormValue    string
		passFormValue     string
		initDataFormValue string
		user_id           int
		initData          interface{}
		//panicCode         int = 500 //181228 400 or 500 For what cause is the panic?
	)
	var sendResult = func(code int, mess string) { //see 201209 06:48 note
		mess = "Authorisation: " + mess
		w.Header().Add("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(code)
		w.Write([]byte(mess))
	}

	defer func() { //see 201209 06:48 note
		if rec := recover(); rec != nil {
			mess := kerr.GetRecoverErrorText(rec)
			mess = "Authorisation (loginpost function problem): " + mess
			kerr.SysErrPrintln(mess)

			w.Header().Add("Content-Type", "text/plain; charset=utf-8")
			w.WriteHeader(500)
			w.Write([]byte(mess))
		}
	}()

	if !KsessRuns() {
		panic("The ksess framework has not been run")
	}

	if checkUserCredentails == nil {
		panic(errors.New(fmt.Sprintf("no checkUserCredentails function")))
	}

	switch r.Method {
	case "POST":
		if err = r.ParseForm(); err != nil {
			panic(errors.New(fmt.Sprintf("Error of r.ParseForm(): %v", err.Error())))
		}
		loginFormValue = r.FormValue("login")
		if loginFormValue == "" {
			sendResult(400, fmt.Sprint("not \"login\" field "))
			return
		}
		passFormValue = r.FormValue("pass")
		if passFormValue == "" {
			sendResult(400, fmt.Sprint("not \"password\" field "))
			return
		}

		initDataFormValue = r.FormValue("initData")

		user_id, errMess = checkUserCredentails(loginFormValue, passFormValue)

		if user_id > -1 { //Success of checking credentials
			if hub.userRegistered(user_id) {
				sendResult(400, fmt.Sprint("user %v (id = %v) has already registered", loginFormValue, user_id))
				return
			}

			cookData.UserID = user_id
			cookData.Tag, err = kutils.TrueRandIntAsInt()
			if err != nil {
				kerr.SysErrPrintf("loginpost: error of getting TrueRandIntAsInt = %s", err.Error())
				cookData.Tag = 123456789
			}

			//Gettting initData
			//Here the priority is done to the data from the "initData" parameter of the request
			if initDataFormValue != "" {
				initData = initDataFormValue
			} else {
				if getInitData != nil {
					initData, err = getInitData(user_id)
					if err != nil {
						initData = nil
						kerr.SysErrPrintf("Error of calculating the init data of %v: %v", user_id, err.Error())
					}
				}
			}

			if err = setSession(cookData, initData, w, r.RemoteAddr, r.Host); err != nil {
				panic(errors.New(fmt.Sprintf("error of establishing a session (of SetSession): %v", err.Error())))
			}
			clearLogErrCookie(w)

			http.Redirect(w, r, sessCP.IndURL, 302)

		} else { //user_id<0
			if errMess == "" {
				errMess = "checkUserCredentails: Во дела! user_id<0, а сообщение об ошибке пусто"
				kerr.SysErrPrintln(errMess)
			}
			if sessCP.OnFaultRegictrationRedirectTo != "" {
				setLogErrCookie(w, errMess)
				//http.Redirect(w, r, sessCP.LoginURL, 302)
				http.Redirect(w, r, sessCP.OnFaultRegictrationRedirectTo, 302)
			} else {
				sendResult(400, fmt.Sprintf("user %v is not exist (checking credentials is fault)", loginFormValue))
				return
			}
		}
	default:
		kerr.SysErrPrintln("loginpost: Not POST methods")
		//panicCode = 400
		panic(fmt.Sprintf("Only POST methods are allowed, not %v", r.Method))
	}

}

//see note 201209 _______14:16 (Like loginpost)
//
//201209 15:26 The big principle: Not doubling info! (it is about if _, cookData, err = getSession(r);)
//func logout(w http.ResponseWriter, r *http.Request, cookData sessCookieData, cln *sessClient) {
func logout(w http.ResponseWriter, r *http.Request) {
	var (
		//anicCode int = 500 //181228 400 or 500 For what cause is the panic?
		cookData sessCookieData
		mess     string
	)
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

	if !KsessRuns() {
		panic("The ksess system is not run")
	}

	if _, cookData, err = getSession(r); err != nil {
		panic("getting session parameters: getSession(r) returns err!=nil")
	}

	if cookData.UserID < 0 { //the session does not exist
		sendResult(400, fmt.Sprintf("the session DOES NOT EXIST (user_id=%v; tag=%v)", cookData.UserID, cookData.Tag))
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
