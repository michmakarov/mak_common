package ksess

import (
	"errors"
	"fmt"
	"net/http"

	"mak_common/kerr"
	"strconv"

	"github.com/gorilla/securecookie"
)

const LogErrorCookieName = "log_error"
const SesssionCookieName = "com180417_session"

//it is a public data that is assosiated with some HTTP request
//if the request is belonging some session then UserID>=0
type SessionData struct {
	UserID int
	Tag    int

	RemoteAddr string //The address the registration has been tied if UserID>=0
	//Otherwise it is remote address from the request
	Host string //The host the registration has been tied if UserID>=0
	//Otherwise it is host from the request

	//== nil if UserID<0
	InitData interface{}
	Data     interface{}
}

func (sd SessionData) SessAsString() string {
	return fmt.Sprintf("ID=%v;RA=%v;Host=%v;", sd.UserID, sd.RemoteAddr, sd.Host)
}

//it is a private data that is a session cookie's value in its natural state before encoding
type sessCookieData struct {
	UserID int
	Tag    int
}

func (cd sessCookieData) UserIDAsString() string {
	return strconv.Itoa(cd.UserID)
}

//var cookieHandler = securecookie.New(
//	securecookie.GenerateRandomKey(64),
//	securecookie.GenerateRandomKey(32),
//)
var cookieHandler *securecookie.SecureCookie

//Returns error if encoding of the session data is failed
func setSession(cookieData sessCookieData, initData interface{}, response http.ResponseWriter, remoteHTTP string, host string) (err error) {
	var (
		encoded string
		cookie  = http.Cookie{
			Name:  SesssionCookieName,
			Value: encoded,
			Path:  "/",
		}
		cln *sessClient
	)

	if !KsessRuns() {
		panic("SetSession err: The ksess system is not run")
	}

	encoded, err = cookieHandler.Encode(SesssionCookieName, cookieData)
	if err == nil {
		cookie.Value = encoded
		cln = newClient(cookieData.UserID, cookieData.Tag, remoteHTTP, host)
		cln.InitData = initData
		hub.registerSess(cln) //<- cln //?? Where is a guarantee that will the session be registered?
		http.SetCookie(response, &cookie)

	}
	return err
}

//if the request is not beloning some secssion the sessionData.UserID<0
func GetSession(request *http.Request) (sessionData SessionData) {

	var (
		err      error
		c        *sessClient
		cookData sessCookieData
	)
	c, cookData, _ = getSession(request)
	sessionData.UserID = cookData.UserID
	sessionData.Tag = cookData.Tag
	sessionData.RemoteAddr = request.RemoteAddr
	sessionData.Host = request.Host
	if sessionData.UserID < 0 {
		return
	}
	sessionData.RemoteAddr = c.RemoteHTTP //181003
	sessionData.Host = c.Host             //181005
	if sessionData.InitData, err = GetSessInitData(sessionData.UserID); err != nil {
		kerr.SysErrPrintf("ksess.GetSession: %v", err.Error())
	}

	if sessionData.Data, err = GetSessData(sessionData.UserID); err != nil {
		kerr.SysErrPrintf("GetSession: %v", err.Error())
	}

	return
}

//It will be call for all requst!!!
//returns error if decoding the cookie was succeded but UserID<0
//that is an error reflects a strange situation when  the server has set user id less than zero
//You, programmer with head, think why do not you declare user id as unsigned integer?
//cookData.UserID>=0 - the session exists, otherwise
// == -11 - no session cookie
// == -1 - a session cookie is decoded with error
// == -2 - decoding was sucsessful but no such seccion
// == -21 - decoding was sucsessful but cookData.UserID<0 ??? in this only case an err is not nil
func getSession(request *http.Request) (c *sessClient, cookData sessCookieData, err error) {
	var (
		cookie *http.Cookie
		res    int
	)

	if !KsessRuns() {
		panic("GetSession err: The ksess system is not run")
	}
	cookData = sessCookieData{}
	cookie, err = request.Cookie(SesssionCookieName)
	if err == nil { // Cookie of SesssionCookieName has been found
		err = cookieHandler.Decode(SesssionCookieName, cookie.Value, &cookData)
		if err != nil { //Decoding was not successful - an user agent has not a right cookie
			cookData.UserID = -1
			cookData.Tag = -1
			c = nil
			err = nil
		} else { //Decoding was successful
			if cookData.UserID < 0 { //???
				err = errors.New("GetSession error: the cookie exists and decoded, but user_id<0")
				cookData.UserID = -21
				cookData.Tag = -21
				c = nil
				return
			}
			res, c = hub.clnRegistered(cookData.UserID, cookData.Tag)
			switch res {
			case 0:
				//Decoding was successful, but such session does not exist
				cookData.UserID = -2
				cookData.Tag = cookData.UserID //181003 - fo saving UserID
				c = nil
			case 1, 2:
				//Decoding was successful, and session does exist
				//
			}
		}
	} else { // Cookie of SesssionCookieName has not been found
		cookData.UserID = -11
		cookData.Tag = -11
		err = nil
		c = nil
	}

	return c, cookData, err
}

//

func clearSession(response http.ResponseWriter) {
	cookie := &http.Cookie{
		Name:   SesssionCookieName,
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	}
	http.SetCookie(response, cookie)

}

func ClearSession(w http.ResponseWriter, r *http.Request) (err error) {
	_, cookData, err1 := getSession(r)
	if err1 != nil {
		err = err1
		return
	}
	if cookData.UserID >= 0 {
		err = errors.New(fmt.Sprintf("ClearSession error:  session for user_id=%v yet exists", cookData.UserID))
		return
	}
	clearSession(w)
	return
}

func setLogErrCookie(response http.ResponseWriter, mess string) {
	cookie := &http.Cookie{
		Name:  LogErrorCookieName,
		Value: mess,
		Path:  "/",
	}
	http.SetCookie(response, cookie)
}

func clearLogErrCookie(response http.ResponseWriter) {
	cookie := &http.Cookie{
		Name:   LogErrorCookieName,
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	}
	http.SetCookie(response, cookie)

}

func getLogErrCookie(request *http.Request) (logErr string) {
	cookie, err := request.Cookie(LogErrorCookieName)
	if err == nil {
		logErr = cookie.Value
	}
	return logErr
}
