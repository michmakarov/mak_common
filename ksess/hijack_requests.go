//181228
//Control request (see control_requests.go) is hijacked too, but it is intercepted before finding out a session
//Requests here are worked after finding out a session. That is their being worked depends on  a session is or not.
package ksess

import (
	"encoding/json"
	"fmt"
	"mak_common/kerr"
	"net/http"
	"os"
	"strings"
	"time"
)

//201208 16:26 But how to be with the big principle?
//Let's say that it is a help function for isolating a part of code
//and that in that case disturbing the big principle is alowed.
//________________
//return true if the request was wrought (processed)
//it look up the list of requests distincly for user id<0 and user id>=0
//the list contains "/ws", "/login", "/logout"
//That means that a request may be or may not be hijaked in dependce of its session cookie
//Only "/ws" is hijacked at any conditions
//181228_1
func doHijackedRequest(w http.ResponseWriter, r *http.Request, cookData sessCookieData, c *sessClient) (yes bool) {
	yes = true
	kerr.PrintDebugMsg(false, "ws", fmt.Sprintf("doHijackedRequest HERE; cookData%v, c=%v", cookData, c))
	//if cookData.UserID < 0 {
	switch r.URL.Path {
	case "/ws":
		//serveWs(hub, w, r)
		w.WriteHeader(400)
		w.Write([]byte(fmt.Sprintf("the /ws is not now supported")))
		return true
	case "/login":
		if !sessCP.RegistrationThrouLogin {
			return false
		} else {
			loginpost(w, r)
			return true
		}
		//loginpost(w, r)
		//return true
	case "/logout":
		if !sessCP.RegistrationThrouLogin {
			return false
		} else {
			logout(w, r)
			//w.WriteHeader(400)
			//w.Write([]byte(fmt.Sprintf("(RegistrationThrouLogin==true)/logout may be requested by  registered user only")))
			return true
		}
	case "/ping":
		{
			pingHandler(w, r)
			return true //At any outcome this request will be performed.
		}
	} //switch

	return false //it is for all requests not listed in the switch
}

//201222 16:53
//if agents are not supported the request is not wrought (is leaved to the programmer's handler)
//r.URL.Path == "/ping"
func pingHandler(w http.ResponseWriter, r *http.Request) (performed bool) {
	type PA struct {
		PingTag      string
		From         string
		RequestCount int64
		AnswerTime   string
	}
	var err error
	var pa PA
	var pingTag string
	var answer []byte
	if sessCP.AgentPassword == "" { //agents are not supported and the request is not wrought (is leaved to the programmer\'s handler)
		performed = false
		return
	}

	pingTag = strings.TrimSpace(r.FormValue("PING_TAG"))
	if pingTag == "" {
		pingTag = "no ping tag"
	}
	pa.PingTag = pingTag
	if pa.From, err = os.Hostname(); err != nil {
		pa.From = err.Error()
	}
	pa.AnswerTime = time.Now().Format(startFormat)
	pa.RequestCount = flr.feelerCount
	answer, _ = json.Marshal(pa)
	w.Header().Set("Contenr-Type", "application/json; charset=utf-8")
	w.Write(answer)
	return

}
