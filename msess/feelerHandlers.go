// feelerHandlers
package msess

import (
	"fmt"
	"io/ioutil"
	"mak_common/kerr"
	"mak_common/kutils"
	"net/http"
	"os"
	"time"
)

func FeelerHandlers() {
	fmt.Println("feelerHandlers: Hello World!")
}

//210609 03:33 See sessCP.AgentForceReg and func CreateHub (fifth parameter)
//210607 16:12 This hunc generates answers of (have intercepted) requests with the Path="/"
//210419 10:40 it registers a new agent!
//That is it is helper function that may be called only in above pointed place.
func indexHandler(w http.ResponseWriter, r *http.Request) {
	var err error
	var currAgent *Agent
	var newAgent *Agent
	var indF *os.File
	var cookieData SessCookieData
	var b []byte
	var errCode = 500
	defer func() {
		if rec := recover(); rec != nil {
			//panicMessage := fmt.Sprintf("(Addr=%v;N=%v) panic:%v", r.RemoteAddr, rc, rec)
			panicMessage := fmt.Sprintf("(Addr=%v) indexHandler panic:%v", r.RemoteAddr, rec)
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.WriteHeader(errCode)
			w.Write([]byte(panicMessage))
			kerr.SysErrPrintf("msess.indHandler panics: = %v", panicMessage)
		}
	}()

	if currAgent = agentRegistered(r); currAgent != nil {
		if !sessCP.AgentForceReg {
			errCode = 403
			panic(fmt.Sprintf("agent already exist:%v", currAgent))
		} else {
			if err = unregAgent(currAgent); err != nil {
				errCode = 400
				panic(fmt.Sprintf("Error of unregistration existing agent:%v", err.Error()))
			}
		}
	}

	if indF, err = os.Open("ind.html"); err != nil {
		panic(fmt.Sprintf("ind.html err=%v", err.Error()))
	}
	if b, err = ioutil.ReadAll(indF); err != nil {
		panic(fmt.Sprintf("Reading from ind.html err=%v", err.Error()))
	}
	//if tag = kutils.TrueRandInt(); tag=""

	newAgent = &Agent{
		RegTime: time.Now(),

		RemoteAddress: r.RemoteAddr,
		UserAgent:     r.UserAgent(),
		Tag:           kutils.TrueRandInt(),
	}

	if err = regAgent(newAgent); err != nil {
		panic(fmt.Sprintf("registration new agent err=%v", err.Error()))
	}
	cookieData.Tag = newAgent.Tag
	if err = setCookieData(cookieData, w); err != nil {
		panic(fmt.Sprintf("setting cookie data err=%v", err.Error()))
	}

	w.Header().Add("Content-Type", "text/html; charset=utf-8")
	w.Write(b)

}

func agentHandler(w http.ResponseWriter, r *http.Request) {
	kerr.PrintDebugMsg(false, "DFLAG210610", "Next serving agent.js")
	http.ServeFile(w, r, "agent.js")
}
