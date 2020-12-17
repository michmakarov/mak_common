//181228
//Control request (see control_requests.go) is hijacked too, but it is intercepted before finding out a session
//Requests here are worked after finding out a session. That is their being worked depends on  a session is or not.
package ksess

import (
	"fmt"
	"mak_common/kerr"
	"net/http"
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
	} //switch

	return false //it is for all requests not listed in the switch
}
