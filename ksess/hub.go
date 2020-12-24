/*
-------VERSUION NOTES--------
190425 - First surprise - third parameter of main function (CreateHub; cuc CheckUserCredentails) may be nill
This version is reasoned by screwing KSESS to KSODD (pgf_190418)
*/
package ksess

import (
	"errors"
	"fmt"

	//	"database/sql"
	"encoding/json"
	"net/http"
	"reflect"

	//"runtime"
	//"log"
	"strconv"
	"sync"
	"time"

	"os"

	"context"

	"mak_common/kerr"
	"mak_common/kutils"
	//"github.com/gorilla/mux"
	//_ "github.com/mattn/go-sqlite3"
)

//The type sessHub and its pointer methods are privite

//The question - may a user have registered the session from a one host but create WS connection trom an other host
//The question - for what is the runSleepTime needed?
//I think that it is the big beastliness to run infinite loops withot delaying into iteration

const userSendChanLen = 255

/* 181024
type HttpPingAnswer struct {
	PingTag      string
	From         string
	Answertime   string
	RuquestCount int64
}
*/

//type toClient
type toClients struct { //СC (сообщение сервера) - see KSCEX
	users   []int //if len(Clients)==0 then massage must be sent to all clients
	message []byte
}

type sessHub struct { //see global var hub . They are for onnly privite usage
	clients map[*sessClient]bool

	//through it sessHub receive a message for sending to a client or clients
	outChan chan toClients
}

//201203 14:45
//181230 The question: Are parameters dependent of order in which their are entered? So far it seems no
type SessConfigParams struct {
	NotFeelerLogging bool //for increasing feeler's speed if no interest who forced his way to the server
	NotUserLogging   bool //? for what

	//201203 14:45
	//UsersLogDir      string // "" (default) or with the file separator at the end, e.g. "/home ... logs/"
	//UsersLogName     string // default "usersLog.sqlite"
	LogsDir string // "" (default, the working directory) or with the file separator at the end, e.g. "/home ... logs/"

	//--------201208 20:13 !!!
	IndURL string //default "/"; The URL to where the redirection will be performed when a registration is succsessful
	//--------------------

	HubRunSleepTime int //setting to not less than 10 - the sleep time in millisecond in iteration of infinite loop of the run method

	// Time allowed to read the next pong message from the peer, seconds
	PongWait int // setting to not less than 60 seconds

	// Send pings to peer with this period. Must be less than pongWait.
	PingPeriod time.Duration //setting to  (PongWait * 9) / 10

	// Maximum message size allowed from peer, bytes
	MaxMessageSize int64 // setting to not less than 4096 bytes

	//-------------------------

	//201222 08:17 At the moment there is only application: debug.PrintStack() when feeler catchs panic
	Debug int // 0 - no debug at all
	//-------------

	//TLS params; if (CertFile!="") then ListenAndServeTLS is run
	CertFile, KeyFile string
	//---------------

	//The Listening address; if "" then ":8080"
	Listening_address string
	//----------------

	Admins []int //Administrators, default Admins={0}

	//201222 06:53 -----------------
	//if =="" agents not supporting
	AgentPassword string //ControlPassword           string//201222 06:17 see agents.go checkAgent
	//---------------------

	RedirectOnNoAuthorisation string

	//--------------------- 181102 201221 06:23
	//201222 06:16 agentPassword is instead NoAgent bool // NotAgentDebugging bool
	//AgentFileDir      string //by default it is in the working directory with name "agents"
	//AgentWorkerDir    string //by default it is "" - that it is working directory
	//---------------------
	//--------------------- 181121_1
	WithoutHTTPActivity int // minutes - How many minutes some session may exist without HTTP activity provided there is not WS connection
	//"provided" here means if the parameter is elapsed but WS is lasted the session will last otherwse it will be ended
	//It may not be less then its default value - 15 minutes

	CheckWithoutHTTPActivityAfter int // Number of HubRunSleepTime periods which compile the period of checking HTTP activity
	//It may not be less then its default value - 100
	//---------------------
	//--------------------- 181228_1
	RegistrationThrouLogin        bool   //if it established a request of  "/login" will be treated as permitted way to create a session and using of CreateSess will be forbade
	OnFaultRegictrationRedirectTo string //Where to redirect if a request of  "/login" is faulted
	//if the registration is not faulted redirection to IndURL will be applied
	//if it == "" (default) the redirection will not be applied at all
	//---------------------
	//--------------------- 181228_2
	CleanUpNotDoneRequestStorage int //the period of cleaning up the global storage of not done requsts in millisecond
	//If it less tan 50 it will be set in 50 (the default value)
	//---------------------
	//-----------------------191223
	HurryForbidden bool
	//---------------------
	//201224 06:01 -----------------
	//Milliseconds of waiting a result from callback functions. Not less then 500
	CallBakTimeout int
	//------------------------------
}

/*
func (scp *SessConfigParams) actualPar() (asp string) {
	var val reflect.Value
	val = reflect.ValueOf(*scp)
	for i := 0; i < val.NumField(); i++ {
		asp = asp + val.Type().Field(i).Name + "==" +
			fmt.Sprint(val.Field(i).Interface()) +
			"\n"
	}
	asp = asp + openUsersLogResultMess + "\n"
	return
}
*/

func (scp *SessConfigParams) actualParWithLb(lb string) (asp string) { //181121_
	var val reflect.Value
	val = reflect.ValueOf(*scp)
	for i := 0; i < val.NumField(); i++ {
		asp = asp + val.Type().Field(i).Name + "==" +
			fmt.Sprint(val.Field(i).Interface()) +
			lb
	}
	//asp = asp + openUsersLogResultMess + lb
	asp = asp + lb
	return
}

func (scp *SessConfigParams) isAdmin(user_id int) bool {
	for _, v := range scp.Admins {
		if v == user_id {
			return true
		}
	}
	return false
}

func IsAdmin(user_id int) bool {
	return sessCP.isAdmin(user_id)
}

var ( //-m "190820(+2) Mending the problem of request counter"
	Version                      string = "190820" //"190715_closed190811" //"190425_closed190715" //"190402_closed190416" //"190128_developing" //"190124" // "190117" //"181228_developing" //"181128" //"181121" //"181102" //"181024" //"181022" //"181019" //"181005" "181003" //"180813"
	ErrUserAlreadyRegistered            = errors.New("var ErrUserAlreadyRegistered: User already registered")
	hubPtotector                        = sync.Mutex{}
	hubPtotectorForTagRegistered        = sync.Mutex{} //181228
	allowWS                      bool                  //if true the web socket is allowed

	hub *sessHub //It is the only structure for a session registration records
	// hub maintains the set of active clients

	flr           *feeler
	sessCP        *SessConfigParams //201221 07:20 Actual config params. set by CreateHub
	server        *http.Server
	ServerStopped chan struct{} = make(chan struct{})
	serverStopped chan string   = make(chan string) //181024_1

	usersLog *os.File //*sql.DB //201203 07:30

	serverStart string //Time of starting the application using this packet. See init function.
)

func GetActualSCP(lb string) string {
	if sessCP == nil {
		return ""
	}
	return sessCP.actualParWithLb(lb)
}

func init() {
	//generalLogFileName = "General" + time.Now().Format("20060102_150405") + ".log"
	/* 201204 07:51 moved to CreateHub
	if createGeneralLog(); err != nil {
		kerr.SysErrPrintf("createGeneralLog err=%v", err.Error())
	} else {
		gLog.run()
	}
	SendToGenLog("init()(ksess)", "general log created")
	*/
	serverStart = time.Now().Format(startFormat)
	//initGlobalNotDone()
	/*
		if err := restoreSessions(); err != nil {
			fmt.Printf("ksess.init: fatal exit code 1001; err = %v", err.Error())
			os.Exit(1001)
		}

		SendToGenLog("init", "ksess initialized")
	*/
}

//ServerStart returns time of initialization of package "ksess" in format of  "20060102_150405"
func ServerStart() string {
	return serverStart
}

//Returns true if CreateHub has been successfully called
func KsessRuns() bool {
	return hub != nil
}

func (h *sessHub) unregisterSess(user_id int) {
	//var cln *sessClient
	var mtx sync.Mutex
	mtx.Lock()
	for k, _ := range h.clients {
		if k.User_ID == user_id {
			if (k.conn != nil) && (k.send != nil) { //that is the client has WS
				close(k.send) //It will stop the client's writePump
			}
			delete(h.clients, k)
			deleteSavedSess(user_id)
			SendToGenLog("unregistered", k.String_190704())
			break
		}
	}
	mtx.Unlock()
}

func (h *sessHub) unregisterAll() {
	hubPtotector.Lock()
	for k, _ := range h.clients {
		if (k.conn != nil) && (k.send != nil) { //that is the client has WS
			close(k.send) //It will stop the client's writePump
		}
		time.Sleep(10000) //for unsetting WS
		delete(h.clients, k)
	}
	hubPtotector.Unlock()
}

func (h *sessHub) registerSess(cln *sessClient) {
	//cln.Since = time.Now()
	hubPtotector.Lock()
	if !h.clients[cln] {
		h.clients[cln] = true
		cln.LastHTTP = time.Now()
		saveSess(cln)
		SendToGenLog("registered", cln.String_190704())
	} else {
		kerr.SysErrPrintf("(h *sessHub) registerSess: Attempt to register a registered user (%v)", cln.User_ID)
	}
	hubPtotector.Unlock()
}

func (h *sessHub) unsetWS(c *sessClient) {
	var exist bool
	hubPtotector.Lock()
	defer hubPtotector.Unlock()
	if c.User_ID < 0 {
		kerr.SysErrPrintf("!!!! unsetWS: user<0; user_id=%v", c.User_ID)
		return
	}
	for cln := range h.clients {
		if c.User_ID == cln.User_ID {
			if cln.conn == nil {
				//kerr.SysErrPrintf("Attempt of unsetting WS for a client without WS; user_id=%v", c.User_ID)
				return
			}
			cln.conn.Close()
			cln.conn = nil
			close(cln.send) //!!! this stops writePump
			exist = true
			break
		}
	}
	if !exist {
		kerr.SysErrPrintf("Attempt of unsetting WS for a not registered  client; user_id=%v", c.User_ID)
	}
}

//returns 0 - if no cln registered (with given user_id and tag)
//		1 - if the cln registered but no socket connection
//		2 - if the cln registered and the socket connection exists
func (h *sessHub) clnRegistered(user_id, tag int) (res int, cln *sessClient) {
	if h == nil {
		panic("hub.clnRegistered: hub==nil")
	}
	//sessExMtx.Lock()
	//defer sessExMtx.Unlock()
	hubPtotector.Lock()
	defer hubPtotector.Unlock()
	for k, _ := range h.clients {
		if (k.User_ID == user_id) && (k.Tag == tag) {
			if k.conn == nil {
				return 1, k
			} else {
				return 2, k
			}
		}
	}
	return 0, nil

}

func (h *sessHub) userRegistered(user_id int) bool {
	if h == nil {
		panic("hub.clnRegistered: hub==nil")
	}
	hubPtotector.Lock()
	defer hubPtotector.Unlock()
	for k, _ := range h.clients {
		if k.User_ID == user_id {
			return true
		}
	}
	return false

}

func (h *sessHub) tagRegistered(tag string) bool {
	if h == nil {
		panic("hub.tagRegistered: hub==nil")
	}
	hubPtotectorForTagRegistered.Lock()
	defer hubPtotectorForTagRegistered.Unlock()
	for k, _ := range h.clients {
		if strconv.Itoa(k.Tag) == tag {
			return true
		}
	}
	return false

}

func (h *sessHub) clnlist(nl string) (lst string) {
	var in int //item number
	if h == nil {
		panic("hub.clnRegistered: hub==nil")
	}

	hubPtotector.Lock()
	defer hubPtotector.Unlock()

	if len(h.clients) == 0 {
		lst = "Ни одного сеанса не зарегистрировано."
		return
	}

	for k, _ := range h.clients {
		in++
		lst = lst + strconv.Itoa(in) + ") " + k.String(nl) + nl
	}
	return

}

func (h *sessHub) clnIdlenesslist(nl string) (lst string) { //181128
	var in int //item number
	if h == nil {
		panic("hub.clnIdleneslist: hub==nil")
	}

	hubPtotector.Lock()
	defer hubPtotector.Unlock()

	if len(h.clients) == 0 {
		lst = "Ни одного сеанса не зарегистрировано."
		return
	}

	for k, _ := range h.clients {
		in++
		lst = lst + strconv.Itoa(in) + ") " + k.String_181128_idleness(nl) + nl
	}
	return

}

//if no clients it returns nil
func (h *sessHub) clnts() (lst []int) { //181128
	//var in int //item number
	if h == nil {
		panic("hub.clnRegistered: hub==nil")
	}

	hubPtotector.Lock()
	defer hubPtotector.Unlock()

	if len(h.clients) == 0 {
		return
	}
	lst = make([]int, 0)
	for k, _ := range h.clients {
		//in++
		lst = append(lst, k.User_ID)
	}
	return
}

func (h *sessHub) clnt(user_id int, nl string) (info string) { //181128
	var noClnt = true //No such client
	if h == nil {
		panic("hub.clnRegistered: hub==nil")
	}

	hubPtotector.Lock()
	defer hubPtotector.Unlock()

	if len(h.clients) == 0 {
		info = "Ни одного клиента не зарегистрировано!" + nl
		return
	}

	for k, _ := range h.clients {
		if k.User_ID == user_id {
			noClnt = false
			info = k.String_181128(nl)
			return
		}
	}
	if noClnt {
		info = fmt.Sprintf("Нет сеанса с таким пользователем user_id = %v;%v", user_id, nl)
	}
	return
}

func (h *sessHub) notDoneList(nl string) (lst string) {
	var in int //item number
	if h == nil {
		panic("hub.clnRegistered: hub==nil")
	}

	hubPtotector.Lock()
	defer hubPtotector.Unlock()

	if len(h.clients) == 0 {
		lst = "Ни одного сеанса не зарегистрировано."
		return
	}
	lst = "Сеансы с невыполнеными запросами --------------------" + nl
	for k, _ := range h.clients {
		if k.isNotDone() {
			in++
			lst = lst + strconv.Itoa(in) + ") " + k.String(nl) + nl
		}
	}
	lst = lst + "--------------------------------"
	return

}

func (h *sessHub) notDoneForUser(user_id int, nl string) string {
	var notDone string
	if h == nil {
		panic("hub.clnRegistered: hub==nil")
	}
	hubPtotector.Lock()
	defer hubPtotector.Unlock()

	notDone = "Не выполненные  ---------" + nl

	for k, _ := range h.clients {
		if (k.User_ID == user_id) && k.isNotDone() {
			notDone = notDone + k.String(nl) + nl
		}
	}
	notDone = notDone + " ---------" + nl

	return notDone

}

func (h *sessHub) checkForIdleness() { //181121_1
	var ui int
	if h == nil {
		panic("hub.clnRegistered: hub==nil")
	}
	hubPtotector.Lock()
	defer hubPtotector.Unlock()

	for k, _ := range h.clients {
		//if time.Since(k.Since) > time.Duration(sessCP.WithoutHTTPActivity)*time.Minute {
		if time.Since(k.LastHTTP) > time.Duration(sessCP.WithoutHTTPActivity)*time.Minute { //181128
			if k.conn == nil {
				ui = k.User_ID
				h.unregisterSess(k.User_ID)
				kerr.SysErrPrintf("checkForIdleness: unregisterSess %v", ui)
			}
		}
	}

	return

}

//Public function
//returns 0 - if no session registered
//		1 - if the session registered but no socket connection
//		2 - if the session registered and the socket connection exists
func SessExists(user_id, tag int) int {
	res, _ := hub.clnRegistered(user_id, tag)
	return res
}

func GetClientList(nl string) string {
	if hub == nil {
		return "CreateHub was not called, the ksess system is not available"
	}
	return hub.clnlist(nl)
}

func GetClientsIdleness(nl string) string {
	if hub == nil {
		return "CreateHub was not called, the ksess system is not available"
	}
	return hub.clnIdlenesslist(nl)
}

func GetClients(nl string) string {
	var clnts []int
	if hub == nil {
		return "GetClients: CreateHub was not called, the ksess system is not available" + nl
	}
	clnts = hub.clnts()
	if clnts == nil {
		return "GetClients: Клиентов не зарегистрировано" + nl
	}
	return fmt.Sprintf("%v%v", clnts, nl)
}

func GetClientInfo(user_id, nl string) string {
	var userId int
	var err error
	if userId, err = strconv.Atoi(user_id); err != nil {
		return fmt.Sprintf("%v - это не идентификатор пользователя; Err=%v%v", user_id, err.Error(), nl)
	}
	return hub.clnt(userId, nl)
}

func (h *sessHub) sendTo(toClnts toClients) {
	//var clnt *sessClient
	if toClnts.users == nil {
		kerr.SysErrPrintln("toClnts.clients == nil")
		return
	}
	if len(toClnts.users) == 0 { //sending to all clients
		for clnt, _ := range h.clients {
			if !(len(clnt.send) < userSendChanLen) {
				kerr.SysErrPrintf("Channel for sending WS answer is overflow;user_id=%v", clnt.User_ID)
			}
			clnt.send <- toClnts.message
		}
		return
	}
	for _, user_id := range toClnts.users {
		clntRegistered := false
		for hClnt, _ := range h.clients {
			if user_id == hClnt.User_ID {
				clntRegistered = true
				if !(len(hClnt.send) < userSendChanLen) {
					kerr.SysErrPrintf("Channel for sending WS answer is overflow;user_id=%v", hClnt.User_ID)
				}
				if hClnt.conn != nil {
					hClnt.send <- toClnts.message
				}
			}
		}
		if !clntRegistered {
			kerr.SysErrPrintf("User from the sending list not registered, user_id=%v", user_id)
		}
	}
}

func (h *sessHub) run() {
	var iterationCounter int
	defer func() {
		if rec := recover(); rec != nil {
			kerr.SysErrPrintf("Hub,the abnormal exit from run: error = %v", rec)
			//The signal of correct stopping is closing the channel stop.
		}
	}()
	if h == nil {
		panic("h==nil")

	}
	for {
		iterationCounter++
		select {
		case toClnts := <-h.outChan:
			h.sendTo(toClnts)

		default: //181121_1
			if sessCP.WithoutHTTPActivity > 0 {
				if iterationCounter > sessCP.CheckWithoutHTTPActivityAfter {
					if (iterationCounter % sessCP.CheckWithoutHTTPActivityAfter) == 0 {
						h.checkForIdleness()
					}
				}
			}
			time.Sleep(time.Millisecond * time.Duration(sessCP.HubRunSleepTime))
		} //select
	} //for
} //run
/* removed since 180813
func logoutHandler(w http.ResponseWriter, r *http.Request) {
	logout(w, r)
}
*/
//Sends to the client or clients an own server message
func SendOSM(osm map[string]string) {
	var (
		toClnts toClients
		outMess []byte
	)

	if checkOSM(osm) != nil {
		kerr.SysErrPrintf("A bad OSM; err=%s", err.Error())
		return
	}

	outMess, err = json.Marshal(osm)
	if err != nil {
		kerr.SysErrPrintf("SendOSM: nothing sent; json.Marshal err=%s", err.Error())
		return
	}
	toClnts.users = extractUsers(osm)
	toClnts.message = outMess
	if hub != nil {
		hub.outChan <- toClnts
	}
}

//if (id==nil)&&(err==nil) then there is not initial data
//if (id==nil)&&(err!=nil) then there is an error
func GetSessInitData(user_id int) (id interface{}, err error) {
	var mux sync.Mutex
	var userIs bool

	if user_id < 0 {
		err = errors.New(fmt.Sprintf("GetSessInitData: user_id==%v(<0)", user_id))
		return
	}

	mux.Lock()
	defer mux.Unlock()

	for clnt, _ := range hub.clients {
		if clnt.User_ID == user_id {
			id = clnt.InitData
			userIs = true
		}
	}

	if !userIs {
		err = errors.New(fmt.Sprintf("GetSessInitData: such user (%v) have not a session", user_id))
	}
	return id, err
}

//if (d==nil)&&(err==nil) then there is not initial data
//if (d==nil)&&(err!=nil) then there is an error
func GetSessData(user_id int) (d interface{}, err error) {
	var mux sync.Mutex
	var userIs bool

	if user_id < 0 {
		err = errors.New(fmt.Sprintf("GetSessData: user_id==%v(<0)", user_id))
		return
	}

	mux.Lock()
	defer mux.Unlock()

	for clnt, _ := range hub.clients {
		if clnt.User_ID == user_id {
			d = clnt.Data
			userIs = true
		}
	}

	if !userIs {
		err = errors.New(fmt.Sprintf("GetSessData: such user (%v) have not a session", user_id))
	}
	return d, err
}

//if (d==nil) the function does nothing
//error returns when
//user_id<0
// not such user
func SetSessData(user_id int, d interface{}) (err error) {
	var mux sync.Mutex
	var userIs bool

	if user_id < 0 {
		err = errors.New(fmt.Sprintf("SetSessData: user_id==%v(<0)", user_id))
		return
	}

	mux.Lock()
	defer mux.Unlock()

	for clnt, _ := range hub.clients {
		if clnt.User_ID == user_id {
			clnt.Data = d
			updateSavedSess(clnt) //190702
			userIs = true
		}
	}

	if !userIs {
		err = errors.New(fmt.Sprintf("SetSessData: such user (%v) have not a session", user_id))
	}
	return err
}

func Stop_server(w http.ResponseWriter, r *http.Request) {
	var sd sessCookieData
	if _, sd, err = getSession(r); err != nil {
		s := fmt.Sprintf("The server gets intternal error\n%v", err.Error())
		kerr.SysErrPrintln(s)
		w.WriteHeader(500)
		w.Write([]byte(s))
		return
	}
	if !sessCP.isAdmin(sd.UserID) {
		s := fmt.Sprintf("The user %v does not have aministrative rights for stopping server", sd)
		w.WriteHeader(400)
		w.Write([]byte(s))
		return
	}
	ct := make([]string, 1)
	ct[0] = "text/html; charset=utf-8"
	w.Header()["Content-Type"] = ct
	w.Write([]byte("<h3>The server will be stopped and is not listening to anyone already now</h3>"))

	notDoneList := hub.notDoneList("<br>")
	w.Write([]byte("<p>" + notDoneList + "</p>"))
	fmt.Printf("--S-- Now we are going to call server.Shutdown ...\n")
	fmt.Printf("--S--  ...... before hub.unregisterAll() \n")
	hub.unregisterAll()
	go func() { //As it is proved you are a very big fool!
		server.Shutdown(context.Background())
		//ss := struct{}{}
		//ServerStopped <- ss
		close(ServerStopped)
	}()
	fmt.Printf("--S-- .............\n")
	time.Sleep(time.Millisecond * 50)
	fmt.Printf("--S-- ... done\n")
	return
}

func closeServer() (mess string) { //since 181024
	var err error
	hub.unregisterAll()
	err = server.Close()
	if err != nil {
		fmt.Printf("--S--  kerr.PrintDebugMsg was called with err=%v", err.Error())
	} else {
		fmt.Printf("--S--  kerr.PrintDebugMsg was called with err=%v", err)
	}

	//close(ServerStopped) //Is not it a very very big fooliness?
	if err != nil {
		mess = err.Error()
	} else {
		mess = "no message"
	}
	return
}

func stop_server_181019(w http.ResponseWriter, r *http.Request) { //since 181019

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte("<h3>The server will be stopped and is not listening to anyone already now</h3>"))

	notDoneList := hub.notDoneList("<br>")
	w.Write([]byte("<p>" + notDoneList + "</p>"))
	fmt.Printf("--S-- Now we are going to call stop_server ...\n")
	fmt.Printf("--S--  ...... before hub.unregisterAll() \n")
	hub.unregisterAll()
	go func() {
		server.Shutdown(context.Background())
		close(ServerStopped)
	}()
	fmt.Printf("--S-- .............\n")
	time.Sleep(time.Millisecond * 50)
	fmt.Printf("--S-- ... done\n")
	return
}

/* removed since 180813
func delete_session(w http.ResponseWriter, r *http.Request) {
	var (
		sd      sessCookieData
		user_id int
		err     error
	)
	if _, sd, err = getSession(r); err != nil {
		s := fmt.Sprintf("The server gets intternal error\n%v", err.Error())
		kerr.SysErrPrintln(s)
		w.WriteHeader(500)
		w.Write([]byte(s))
		return
	}
	if !sessCP.isAdmin(sd.UserID) {
		s := fmt.Sprintf("The user %v does not have aministrative rights for deleting session", sd.UserID)
		w.WriteHeader(400)
		w.Write([]byte(s))
		return
	}
	if err = r.ParseForm(); err != nil {
		panic(errors.New(fmt.Sprintf("The ksess delete_session err (of r.ParseForm()): %v", err.Error())))
	}
	userFormValue := r.FormValue("user")
	if userFormValue == "" {
		panic(errors.New(fmt.Sprint("The ksess delete_session err: not \"user\" form field ")))
	}

	if user_id, err = strconv.Atoi(userFormValue); err != nil {
		panic(errors.New(fmt.Sprintf("The ksess delete_session:  \"user\" (%v) is not integer ", userFormValue)))
	}

	if user_id < 0 {
		panic(errors.New(fmt.Sprintf("The ksess delete_session:  user_id (%v) < 0 ", user_id)))
	}

	if !hub.userRegistered(user_id) {
		panic(errors.New(fmt.Sprintf("The ksess delete_session:  user_id (%v)  not registered ", user_id)))
	}
	notDone := hub.notDoneForUser(user_id, "<br>")

	hub.unregisterSess(user_id)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	headMess := fmt.Sprintf("<h3>The session of  %v have been deleted </h3>", user_id)
	w.Write([]byte(headMess))
	w.Write([]byte("<p>" + notDone + "</p>"))
}
*/
func CreateSess(user_id int, initData interface{}, w http.ResponseWriter, r *http.Request) (err error) {
	var (
		cd sessCookieData
	)
	defer func() {
		if err != nil {
			kerr.SysErrPrintf(err.Error())
		}
	}()

	if hub == nil {
		err = errors.New("ksess.CreateSess: the CreateHub was not called properly")
		return
	}
	if sessCP.RegistrationThrouLogin {
		err = errors.New("ksess.CreateSess: Registration is allowed only through /login")
		return

	}

	if user_id < 0 {
		err = errors.New("ksess.CreateSess: user_id<0")
		return
	}
	if hub.userRegistered(user_id) {
		//err = errors.New(fmt.Sprintf("ksess.CreateSess: user with user_id==%v is already registered", user_id))
		err = ErrUserAlreadyRegistered
		return
	}
	cd.UserID = user_id
	cd.Tag, _ = kutils.TrueRandIntAsInt()
	//setSession(cd, initData, w, r.RemoteAddr, r.Host) //190105
	err = setSession(cd, initData, w, r.RemoteAddr, r.Host)
	return
}

func CreateSess_181228(sess SessionData, new_user_id int, initData interface{}, w http.ResponseWriter, r *http.Request) (err error) {
	var (
		cd sessCookieData
	)
	defer func() {
		if err != nil {
			kerr.SysErrPrintf(err.Error())
		}
	}()

	if hub == nil {
		err = errors.New("ksess.CreateSess_181228: the CreateHub was not called properly")
		return
	}
	if sessCP.RegistrationThrouLogin {
		err = errors.New("ksess.CreateSess_181228: Registration is allowed only through /login")
		return

	}

	if new_user_id < 0 {
		err = errors.New("ksess.CreateSess_181228: user_id<0")
		return
	}
	if hub.userRegistered(new_user_id) {
		//err = errors.New(fmt.Sprintf("ksess.CreateSess: user with user_id==%v is already registered", user_id))
		err = ErrUserAlreadyRegistered
		return
	}
	cd.UserID = new_user_id
	cd.Tag, _ = kutils.TrueRandIntAsInt()
	//setSession(cd, initData, w, r.RemoteAddr, r.Host) //190105
	if err = setSession(cd, initData, w, r.RemoteAddr, r.Host); err == nil {
		if sess.UserID >= 0 {
			hub.unregisterSess(sess.UserID)
		}
	}
	return
}

func DeleteSess(user_id int) (err error) {
	var ()
	if user_id < 0 {
		return errors.New(fmt.Sprintf("The ksess.DeleteSess:  user_id (%v) < 0 ", user_id))
	}

	if sessCP.RegistrationThrouLogin {
		err = errors.New("ksess.DeleteSess: logging out is allowed only through /logout")
		return
	}

	if !hub.userRegistered(user_id) {
		return errors.New(fmt.Sprintf("The ksess.DeleteSess:  user_id (%v)  not registered ", user_id))
	}

	hub.unregisterSess(user_id)
	return
}
