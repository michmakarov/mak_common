package msess

import (
	"fmt"
	"net/http"

	//"mak_common/kerr"
	//"strconv"

	"github.com/gorilla/securecookie"
)

const AgentCookieName = "agent_cookie_210304"

//it is a session cookie value in its natural state before encoding
//It is subset of Agent type's fields
type SessCookieData struct {
	Tag string //a unique tag that identifies the agent
	//UserId string // "" means that no user currently enters the system
}

//var cookieHandler *securecookie.SecureCookie

var cookieHandler = securecookie.New(securecookie.GenerateRandomKey(64), securecookie.GenerateRandomKey(32))

//Returns error if encoding of the session data is failed
//!!!It does not actually send the cookie to the http clients
//210305 07:57
//210316 13:40 Yes, it does not send. So it is a function that may be called only in a defined place.
//And the name of setAgentSession is not fire. Let it be setCookieDate
func setCookieDate(cookieData SessCookieData, w http.ResponseWriter) (err error) {
	var (
		encoded string
		cookie  = http.Cookie{
			Name:  AgentCookieName,
			Value: "",
			Path:  "/",
		}
	)

	if !MsessRuns() {
		panic("setAgentSession err: The msess mechanics is not run yet")
	}

	if encoded, err = cookieHandler.Encode(AgentCookieName, cookieData); err != nil {
		err = fmt.Errorf("setAgentSession: Encode err = %v", err.Error())
		return
	}
	cookie.Value = encoded
	http.SetCookie(w, &cookie)
	return
}

//It will be call for all requst!!!
//returns err!=error if decoding the cookie was not succeded or there is not AgentCookieName cookie
func getCookieData(r *http.Request) (cookieData *SessCookieData, err error) {
	var cookie *http.Cookie
	var cd SessCookieData

	if !MsessRuns() {
		panic("getSession err: The msess system is not run")
	}
	if cookie, err = r.Cookie(AgentCookieName); err != nil {
		err = fmt.Errorf("getSession: r.Cookie(AgentCookieName) err=%v", err.Error())
		return
	}
	cd = SessCookieData{}
	if err = cookieHandler.Decode(AgentCookieName, cookie.Value, cd); err != nil { //Decoding was not successful - an user agent has not a right cookie
		err = fmt.Errorf("getSession: Decode err=%v", err.Error())
		cookieData = nil
	}
	cookieData = &cd
	return
}

//

func clearCookieData(response http.ResponseWriter) {
	cookie := &http.Cookie{
		Name:   AgentCookieName,
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	}
	http.SetCookie(response, cookie)

}
