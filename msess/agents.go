package msess

import (
	"context"
	"fmt"
	"net/http"

	//"sync"
	"time"

	"github.com/gorilla/websocket"
	//"os"
	//"path/filepath"
	//"strings"
	//"mak_common/kerr"
	//"mak_common/kutils"
)

const maxAgents = 100

type Agent struct {
	RegTime time.Time //210303 16:06 The moment of registration;

	RemoteAddress string //r.RemoteAddr
	UserAgent     string //r.UserAgent()

	//The next two fields gives content of the agent's coocie
	Tag    string //a unique tag that identifies the agent
	UserId string // "" means that no user currently enters the system

	conn  *websocket.Conn
	WsOut chan WsMess
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

var mqChan chan MonitorQuery

var server *http.Server

//var calcHTTPResponseMtx sync.Mutex

func (a *Agent) shortDescr(sd string) {
	var user string
	if a == nil {
		sd = "no agent"
		return
	}
	if a.UserId == "" {
		user = "no"
	} else {
		user = a.UserId
	}
	sd = fmt.Sprintf("User:%v(%v)", user, a.Tag)
	return
}

func startAgentMonitor() {
	mqChan = make(chan MonitorQuery)
	go agentsMonitor()
}

func agentsMonitor() {
	var mQ MonitorQuery
	var mR MonitorResult
	var wsInMess WsMess
	for true {
		select {
		case mQ = <-mqChan:
			switch mQ.Action {
			case "register":
				mR = register(mQ.Data)
			case "unregister":
				mR = unregister(mQ.Data)
			case "is_registered":
				mR = is_registered(mQ.Data)
			case "send_to_ws":
				mR = sendToWs(mQ.Data)
			case "where_user":
				mR = where_user(mQ.Data)
			case "assign_user":
				mR = assign_user(mQ.Data)
			default:
				mR.Err = fmt.Errorf("agentsMonitor: illegal action (%v) of a query", mQ.Action)
			} //switch
			mQ.ResultChan <- mR
		case wsInMess = <-inWsMessChan:
			doInWsMess(wsInMess)
		} //select
	} //for
} //agentMonitor

func unregAgent(a *Agent) (err error) {
	var mQ MonitorQuery = MonitorQuery{"unregister", a, make(chan MonitorResult)}
	var mR MonitorResult

	if !MsessRuns() {
		panic("Agent unregister: MSess does not run")
	}
	mqChan <- mQ
	mR = <-mQ.ResultChan
	if mR.Err == nil {
		sendNoteAboutUnregister(a)
	}
	return mR.Err
}

func regAgent(a *Agent) (err error) {
	var mQ MonitorQuery = MonitorQuery{"register", a, make(chan MonitorResult)}
	var mR MonitorResult
	if !MsessRuns() {
		panic("regAgent: MSess does not run")
	}
	mqChan <- mQ
	mR = <-mQ.ResultChan
	err = mR.Err
	return
}

//210316 16:36
//It returns err!=nil
//if no agent registered with agent.Tag=cd.Tag
//or if there is such agent but the obtained current cookie is forged
//If err==nil the a is a copy of a registered *Agent
func agentRegistered(cd *SessCookieData, r *http.Request) (a *Agent, err error) {
	var mQ MonitorQuery = MonitorQuery{"is_registered", cd, make(chan MonitorResult)}
	var mR MonitorResult
	var ok bool
	var forgedMess string

	if !MsessRuns() {
		panic("agentRegistered: MSess does not run")
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

	if a.RemoteAddress != r.RemoteAddr {
		forgedMess = fmt.Sprintf("not equal a.RemoteAddress==%v;r.RemoteAddr==%v", a.RemoteAddress, r.RemoteAddr)
		goto forgedAgent
	} else {
		return
	}
	if a.UserAgent != r.UserAgent() {
		forgedMess = fmt.Sprintf("not equal a.UserAgent==%v;r.UserAgent()==%v", a.UserAgent, r.UserAgent())
		goto forgedAgent
	} else {
		return
	}

forgedAgent:
	err = fmt.Errorf("agentRegistered: forgeded agent: %v", forgedMess)
	unregAgent(a)
	sendNoteAboutUnregister(a)
	a = nil
	return
}

//210330 11:48
//The parameter "userId" is apparently excessive. So:
//The "a" must not point to some item of agents but there must be the item with item.Tag==a.Tag; item.UserId==""
//The function does not return any errors but it panics if something is not as it expects.
//See func assign_user(data interface{}) (res MonitorResult)
//The question on 210330 12:43: Must it check for forging of the a?
//210326 04:17
//Let's agent is some item into agents
//Then this function assigns agent.UserId==userId where agent.Tag=a.Tag
//It causes panics through func assign_user:
//if initially agent.UserId != "" or there is not such agent into registry.
//func assignUser(a *Agent, userId string) {
func assignUser(a *Agent) {
	var mQ MonitorQuery = MonitorQuery{"assign_user", a, make(chan MonitorResult)}
	var mR MonitorResult
	var ok bool
	//var forgedMess string

	if !MsessRuns() {
		panic("assignUser: MSess does not run")
	}
	//a.UserId = userId //!!!
	mqChan <- mQ
	mR = <-mQ.ResultChan
	if mR.Err != nil {
		err = mR.Err
		return
	}
	if a, ok = mR.Data.(*Agent); !ok {
		panic("userRegistered: data returned from is_registered is not converted to *Agent")
	}
	return
}

//210325 06:27 This returns copy of the agent where agent.UserId==userId or nil
func whereUser(userId string) (a *Agent) {
	var mQ MonitorQuery = MonitorQuery{"where_user", userId, make(chan MonitorResult)}
	var mR MonitorResult
	var ok bool
	//var forgedMess string

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
		panic("userRegistered: data returned from is_registered is not converted to *Agent")
	}
	return
}

//210319 13:50
//SendMessToAgent makes copy of mess and send it to monitor
func sendMessToAgent(mess WsMess) (err error) {
	var messCopy WsMess
	if messCopy, err = makeCopyAndCheck(mess); err != nil {
		return
	}
	var mQ MonitorQuery = MonitorQuery{"send_to_ws", messCopy, make(chan MonitorResult)}
	var mR MonitorResult
	if !MsessRuns() {
		panic("SendMessToAgent: MSess does not run")
	}
	mqChan <- mQ
	mR = <-mQ.ResultChan
	err = mR.Err
	return
}

//210304 11:02 an agent session's configuration parameters
type SessConfigParams struct {

	//-------------------------

	//210101 This is set of bit flags
	//201222 08:17 At the moment there is only application: debug.PrintStack() when feeler catchs panic
	Debug byte //210104 13:31 it affects through feelerLogger.mode (see func createFlrLog) on
	//210604 06:29
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

	//Admins []string //Administrators, default Admins={"0"} //excluded 210323 15:47

	//210304 11:16
	WithoutActivity int // minutes - How many minutes some session may exist without activity
	//--------------------- 210304 11:16

	//210322 16:37
	ServerReadTimeout int //second >= 1; Server.ReadTimeout = time.Second*time.Duration(ServerReadTimeout)
	//--------------------- ServerReadTimeout

	//210323 15:52
	WithoutHTTPActivity int //minutes; not less 15
	//--------------------- WithoutHTTPActivity

	//--------------------- 210324 05:42 //181228_2
	CleanUpNotDoneRequestStorage int //the period of cleaning up the global storage of not done requsts in millisecond
	//If it less than 100 it will be set in 100 (the default value)
	//---------------------

	//--------------------- 210324 20:56
	CallBakTimeout int //miliseconds; the period of waiting retuning of callback function
	//If it less than 100 it will be set in 100 (the default value)
	//---------------------

	//--------------------- 210603 06:22
	Loggers string // See LOGGING definition
	//---------------------

	HurryForbidden bool
}

func sendNoteAboutUnregister(a *Agent) {
	panic("The sendNoteAboutUnregister has not been realized yet.")
}

//
func calcHTTPResponse(reqNum int64, a *Agent, w http.ResponseWriter, r *http.Request, cancel context.CancelFunc) {
	var (
		ulr       *userLogRecord
		start     string
		begin     = time.Now()
		doneChan2 chan *userLogRecord
		//err          error
	)

	//kerr.PrintDebugMsg(false, "ServeHTTP_201203_1129", fmt.Sprintf("calcHTTPResponse:very start; c=%v", c))

	start = begin.Format(timeFormat)

	//1        					2      3        4    5     6    7    8     9
	//reqNum, 					start, user_id, tag, addr, url, dur, code, extraInfo
	ulr = newUserLogRecord(fmt.Sprintf("%v", reqNum), start, a.UserId, a.Tag, r.RemoteAddr, r.RequestURI, "", "", "")

	doneChan2, err = globalNotDone.addHTTPChore(ulr, w, r, cancel)

	ulr = <-doneChan2 //here we can and most likely will receive a pointer not that that was sent to the addHTTPChore

	//ulr.dur = fmt.Sprintf("%v", time.Now().Sub(begin))

	insertUserLogRecord(ulr)

}
