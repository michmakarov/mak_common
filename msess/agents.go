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
	//"sync"
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
	return fmt.Sprintf("Time=%v;RA=%v;UA=%v", a.RegTime.Format(timeFormat), a.RemoteAddress, a.UserAgent)
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

var sessCP *SessConfigParams

//var agentsMutex sync.Mutex
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
		case "is_registered":
			mR = is_registered(mQ.Data)
		default:
			mR.Err = fmt.Errorf("agentsMonitor: illegal action (%v) of a query", mQ.Action)
		}
		mQ.ResultChan <- mR
	}
} //agentMonitor

func unregAgent(a *Agent) (err error) {
	var mQ MonitorQuery = MonitorQuery{"unregister", a, make(chan MonitorResult)}
	var mR MonitorResult

	if !MsessRuns() {
		panic("Agent unregister: MSess does not run")
	}
	mqChan <- mQ
	mR = <-mQ.ResultChan

	return mR.Err
}

func regAgent(a *Agent) (err error) {
	var mQ MonitorQuery = MonitorQuery{"register", a, make(chan MonitorResult)}
	var mR MonitorResult
	if !MsessRuns() {
		panic("Agent unregister: MSess does not run")
	}
	mqChan <- mQ
	mR = <-mQ.ResultChan
	err = mR.Err
	return
}

//210316 16:36
//It returns err!=nil if a cd.Tag is not registered
//or the agent's (a) data is not corresponded the data of request (r)
//If err==nil the a is a copy of a agents[registered *Agent]
func agentRegistered(cd *SessCookieData, r *http.Request) (a *Agent, err error) {
	var mQ MonitorQuery = MonitorQuery{"is_registered", cd, make(chan MonitorResult)}
	var mR MonitorResult
	var ok bool
	var forgedMess string

	if !MsessRuns() {
		panic("Agent unregister: MSess does not run")
	}
	mqChan <- mQ
	mR = <-mQ.ResultChan
	if mR.Err != nil {
		err = mR.Err
		return
	}
	if a, ok = mR.Data.(*Agent); !ok {
		panic("agentRegistered: data returned from is_registered is not converted to *Agent")
	}
	if a.UserId != cd.UserId {
		forgedMess = fmt.Sprintf("not equal a.UserId==%v;cd.UserId==%v", a.UserId, cd.UserId)
		goto forgedAgent
	}
	if a.RemoteAddress != r.RemoteAddr {
		forgedMess = fmt.Sprintf("not equal a.RemoteAddress==%v;r.RemoteAddr==%v", a.RemoteAddress, r.RemoteAddr)
		goto forgedAgent
	}
	if a.UserAgent != r.UserAgent() {
		forgedMess = fmt.Sprintf("not equal a.UserAgent==%v;r.UserAgent()==%v", a.UserAgent, r.UserAgent())
		goto forgedAgent
	}

forgedAgent:
	err = fmt.Errorf("agentRegistered: forgeded agent: %v", forgedMess)
	unregAgent(a)
	a = nil
	return
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

	//210310 16:26 Path to index file
	//IndexFIle string
	//--------------------- 210310 16:26

	HurryForbidden bool
}

func MsessRuns() bool {
	if server != nil {
		return true
	}
	return false
}
