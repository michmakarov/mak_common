// commandHandlers
package ksess

import (
	"fmt"
	//"io"

	//"io/ioutil"
	//"html/template"
	"mak_common/kutils"
	"net/http"

	//"os"

	//"path/filepath"
	//"strings"
	"mak_common/kerr"
	"sync"
	"time"
)

const maxAgents = 100
const agentPasswordParName = "a_p_p_n"

type Agent struct {
	RegTime time.Time //201221 12:25 The moment of registration; The next data has been at this moment

	RemoteAddress string //r.RemoteAddr
	UserAgent     string //r.UserAgent()
	UserId        int    //cookData.UserID see getSession
}

type Agents map[string]*Agent

//is registered
var agents Agents = make(map[string]*Agent)

//var registerMutex sync.Mutex
//var registeredMutex sync.Mutex
var agentsMutex sync.Mutex

func (a *Agent) String() string {
	return fmt.Sprintf("Time=%v;RA=%v;UA=%v", a.RegTime.Format("20060102_150405"), a.RemoteAddress, a.UserAgent)
}

//The register registers the request it if it has not registered yet. see Registered
//It is helper function for checkAgent
//Errors:
//err = fmt.Errorf("Agents.register: too many agents; allowed not more %v", maxAgents)
//201221 07:08 201222 07:06
func (agents Agents) register(r *http.Request) (signature string, err error) {
	var sessionData SessionData
	var a Agent

	if len(agents) >= maxAgents {
		err = fmt.Errorf("Agents.Register: too many agents; allowed not more %v", maxAgents)
		return
	}

	signature = agents.Registered(r)
	if signature != "" {
		return
	}

	sessionData = GetSession(r)
	signature = kutils.TrueRandInt()
	a.RegTime = time.Now()
	a.RemoteAddress = r.RemoteAddr
	a.UserAgent = r.UserAgent()
	a.UserId = sessionData.UserID
	agentsMutex.Lock()
	agents[signature] = &a
	agentsMutex.Unlock()
	return signature, nil
}

//201221 13:53 see ksess.rules.--AGENT--
//it returns "" if r do not belong any registered agent.
func (a Agents) Registered(r *http.Request) (signature string) {
	agentsMutex.Lock()
	for k, v := range agents {
		if v.RemoteAddress == r.RemoteAddr && v.UserAgent == r.UserAgent() {
			signature = k
			agentsMutex.Unlock()
			return
		}
	}
	agentsMutex.Unlock()
	return
}

func (a Agents) String(lb string) (res string) {
	agentsMutex.Lock()
	for _, value := range a {
		res = res + value.String() + lb
	}
	agentsMutex.Unlock()
	return
}

func GetAgents() Agents {
	return agents
}

// checkAgent does nothin  and returns nil if isOutSess==true or isHijacked(r)==true
//If an error occurs the checkAgent sends to client all necessary messages.
//The returned result indicates whether or not to perform further on the incoming request: if error then not
//201222 06:25; 201223 06:09
func checkAgent(w http.ResponseWriter, r *http.Request, isOutSess bool) (err error) {
	if sessCP.AgentPassword == "" {
		return
	}
	//if isOutSess { //201223 06:09//210222 21:13 All queries must be from a trusted agent!
	//	return // There is "triffles" - from there all may be obtained without any restriction
	//}
	if isHijacked(r) { //201223 20:10
		return
	}
	kerr.PrintDebugMsg(false, "DFLAG210218", fmt.Sprintf(" checkAgent before checking pwd:query=%v; err=%v", r.URL.Query(), err))
	kerr.PrintDebugMsg(false, "DFLAG210218", fmt.Sprintf("VAL=%v;sessCP.AgentPassword=%v", r.FormValue(agentPasswordParName), sessCP.AgentPassword))
	if r.FormValue(agentPasswordParName) != sessCP.AgentPassword {
		err = fmt.Errorf("Not valid agent password")
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(403)
		w.Write([]byte(fmt.Sprintf("%v", err.Error())))
		return
	}
	kerr.PrintDebugMsg(false, "DFLAG210218", fmt.Sprintf(" checkAgent before register:Path=%v; err=%v", r.URL.Path, err))
	if _, err = agents.register(r); err != nil {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(403)
		w.Write([]byte(fmt.Sprintf("%v", err.Error())))
		return
	}
	kerr.PrintDebugMsg(false, "DFLAG210218", fmt.Sprintf(" checkAgent before return:Path=%v; err=%v", r.URL.Path, err))
	return
}
