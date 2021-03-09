package msess

import (
	"fmt"
	//"mak_common/kutils"
	"net/http"

	"github.com/gorilla/websocket"

	//"os"

	//"path/filepath"
	//"strings"
	//"mak_common/kerr"
	"sync"
	"time"
)

const maxAgents = 100

type Agent struct {
	RegTime time.Time //210303 16:06 The moment of registration;

	RemoteAddress string //r.RemoteAddr
	UserAgent     string //r.UserAgent()

	//The next two fields gives content of the agent's coocie
	Tag    string //a unique tag that identifies the agent
	UserId string // "" means that no user currently enters the system

	conn *websocket.Conn
}

func (a *Agent) String() string {
	return fmt.Sprintf("Time=%v;RA=%v;UA=%v", a.RegTime.Format("20060102_150405"), a.RemoteAddress, a.UserAgent)
}

type Agents []*Agent
type MonitorResult struct {
	Err  error
	Data interface{}
}
type MonitorQuery struct {
	Action     string
	Data       interface{}
	ResultChan chan MonitorResult
}

var agents Agents
var agentsMutex sync.Mutex
var mqChan chan MonitorQuery

var server *http.Server

func startAgentMonitor() {
	mqChan = make(chan MonitorQuery)
	go agentsMonitor()
}

func agentsMonitor() {
	var mQ MonitorQuery
	var mR MonitorResult
	for true {
		mQ = <-mqChan
		switch mQ.Action {
		case "register":
			mR = register(mQ.Data)
		case "unregister":
			mR = unregister(mQ.Data)
		default:
			mR.Err = fmt.Errorf("agentsMonitor: illegal action (%v) of a query", mQ.Action)
		}
		mQ.ResultChan <- mR
	}
} //agentMonitor

func (agents Agents) unregAgent(a *Agent) (err error) {
	if !MsessRuns() {
		panic("Agent unregister: MSess does not run")
	}
	unRegChan <- a
	return
}

func (agents Agents) regAgent(a *Agent) (err error) {
	if !MsessRuns() {
		panic("Agent unregister: MSess does not run")
	}
	regChan <- a
	return
}

//
func (a Agents) getCurrAgent(r *http.Request) (curA *Agent, err error) {
	var cookieData *sessCookieData
	if cookieData, err = getCookieData(r); err != nil {
		err = fmt.Errorf("getCurrAgent: getCookieData err = %v ", err.Error())
		return
	}
	agentsMutex.Lock()

	for ind, item := range agents {
		if item.Tag == cookieData.Tag {
			curA = agents[ind]
			break
		}
	}
	agentsMutex.Unlock()

	if curA.RemoteAddress != r.RemoteAddr {
		err = fmt.Errorf("getCurrAgent: there is discrepancy registered remote addresses (%v) and actual one (%v) for agent with tag = %v. The registration record will be removed (unregistered). ",
			curA.RemoteAddress, r.RemoteAddr, curA.Tag)
		return
	}

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

//210304 11:02 an agent session's configuration parameters
type SessConfigParams struct {

	//-------------------------

	//210101 This is set of bit flags
	//201222 08:17 At the moment there is only application: debug.PrintStack() when feeler catchs panic
	Debug byte //210104 13:31 it affects through feelerLogger.mode (see func createFlrLog) on
	//(1) (f *feeler) ServeHTTP defer func() - if 00000000 then not printing the stack when there is a panic
	//(2) behavior of (fl *feelerLogger) getFlrlogMess - see this metod
	//(3) doubling a requestRecord to StdOut (see func (fl *feelerLogger) Run)
	//-------------

	//TLS params; if (CertFile!="") then ListenAndServeTLS is run
	CertFile, KeyFile string
	//---------------

	//The Listening address; if "" then ":8080"
	Listening_address string
	//----------------

	Admins []string //Administrators, default Admins={"0"}

	//210304 11:16
	WithoutActivity int // minutes - How many minutes some session may exist without activity
	//--------------------- 210304 11:16

	HurryForbidden bool
}

func MsessRuns() bool {
	if server != nil {
		return true
	}
	return false
}
