// hubCallBack
//Contains pablic types of callback function that that establishes by the CreateHub function
//and corresponding not public global variables
package ksess

import (
	"errors"
	"fmt"

	"net"
	"net/http"
	"time"

	//"golang.org/x/net/context"

	"mak_common/kerr"
	//"mak_common/kutils"
	//"github.com/gorilla/mux"
)

//201214 07:03 see also rels readme.txt 201214_
//it is an attempt to find a way to understand when and in what condition
//an incoming request enters the handler.
type ConnStateDescr struct {
	conn  net.Conn
	descr string
}

var currConnStateDescr *ConnStateDescr = &ConnStateDescr{nil, ""}

func connStateHook(conn net.Conn, state http.ConnState) {
	if currConnStateDescr.conn != conn {
		currConnStateDescr = &ConnStateDescr{
			conn:  conn,
			descr: "",
		}

	}
	currConnStateDescr.descr = fmt.Sprintf("%v-%v-%v", currConnStateDescr.descr, state, time.Now().Format("15:04:05.000"))
	return
}

//____________________201214 07:03

//It is a dispatcher which
//maps a incoming web socket message to the outcoming map accordong to SCEX
type ParserSocket func(user_id int, mess []byte) map[string]string

//It examines a request URL's path to allow doing the request regardless  existing the session
//see variable checkURLPath and parameter URLCheker of function CreateHub
//201223 06:15 see also OUTSESSION_REQEST into feeler.go_(f *feeler) ServeHTTP
type URLPathChecker func(path string) bool

//It takes a result of  the ParserSocket and extracts from it the list of users to which the result must be sent
//If no such function given then the extractUsersDefault is used
type ExtractUsers func(answer map[string]string) (users []int)

//see api.txt
type CheckUserCredentails func(action, userLogName, userPassword string) (user_id int, errMess string)

type GetInitData func(user_id int) (data interface{}, err error)

var (
	parserSocket         ParserSocket         //1
	extractUsers         ExtractUsers         //2
	checkUserCredentails CheckUserCredentails //3
	reqMultiplexer       http.Handler         //*mux.Router          //4
	checkURLPath         URLPathChecker       //5
	getInitData          GetInitData          //6
)

//A programmer may call or not call this function
//If he want to uses the ksess system he must call this function
//In not the case all the functionality of this package will not be available for him
//That is that this function is the central and the only public entrance  to ksess system
//190702 It initializes the hub, that makes the KsessRuns function to return true.
func CreateHub(ps ParserSocket, //1 not nill
	exUsers ExtractUsers, //2
	cuc CheckUserCredentails, //3 not nill
	mx http.Handler, //201222 16:25 //mx *mux.Router, //4 not nill
	URLCheker URLPathChecker, //5
	initDataGetter GetInitData, //6
	scp *SessConfigParams) (err error) { //CreateHub body
	time.Sleep(time.Second) //191223 For what is it?
	//1 (callback)
	if ps == nil {
		err = errors.New("CreateHub: no ParserSocket")
		return
	} else {
		parserSocket = ps
	}

	//2 (callback)
	if exUsers == nil {
		extractUsers = extractUsersDefault
	} else {
		extractUsers = exUsers
	}

	//3 (callback) from 180808 cuc==nil is not errror since it now just forbid "post login"
	//if cuc == nil {
	//	err = errors.New("CreateHub: no function for checking credentials")
	//	return
	//} else {
	checkUserCredentails = cuc // confurmed 181228, see also SessConfigParams.RegistrationThrouLogin
	//}

	// 201207 07:09
	if (scp.RegistrationThrouLogin == true) && (checkUserCredentails == nil) {
		err = errors.New("CreateHub: scp.RegistrationThrouLogin==true but checkUserCredentails==nil")
		return
	}
	if (scp.RegistrationThrouLogin == false) && (checkUserCredentails != nil) {
		err = errors.New("CreateHub: scp.RegistrationThrouLogin==false but checkUserCredentails!=nil")
		return
	}

	//4 (callback)
	if mx == nil {
		err = errors.New("CreateHub: no handler for incoming requests")
		return
	} else {
		reqMultiplexer = mx
	}

	//5 (callback)
	if URLCheker == nil {
		checkURLPath = defaultURLCheker //permits /app
	} else {
		checkURLPath = URLCheker
	}

	//6 (callback)
	getInitData = initDataGetter

	//scp------

	if scp.CallBakTimeout < 500 {
		scp.CallBakTimeout = 500
	}
	if scp.LogsDir != "" {
		if err = checkLogsDir(scp.LogsDir); err != nil {
			err = fmt.Errorf("CreateHub: LogsDir err=%v", err.Error())
		}
	}

	if scp.IndURL == "" {
		scp.IndURL = "/"
	}

	if scp.HubRunSleepTime < 10 {
		scp.HubRunSleepTime = 10
	}

	if scp.PongWait < 60 {
		scp.PongWait = 60
	}

	if (scp.CertFile != "") && (scp.KeyFile == "") {
		err = errors.New("Two TLS files must  be provided both or they must not be at all.")
		return
	}
	if (scp.CertFile == "") && (scp.KeyFile != "") {
		err = errors.New("Two TLS files must  be provided both or they must not be at all.")
		return
	}

	if scp.Listening_address == "" {
		if scp.CertFile != "" {
			scp.Listening_address = ":443"
		} else {
			scp.Listening_address = ":8080"
		}

	}

	scp.PingPeriod = ((time.Duration(scp.PongWait) * time.Second) / 10) * 9

	if scp.MaxMessageSize < 4096 {
		scp.MaxMessageSize = 4096
	}

	if scp.Admins == nil {
		scp.Admins = []int{0}
	}

	//201222 06:47 agentPassword is instead;
	//if scp.ControlPassword == "" { //181024_2 //201221 07:44 Let's it be
	//	scp.ControlPassword = kutils.TrueRandInt()
	//}

	//if scp.AgentFileDir == "" { //181102 /201221 06:38
	//	scp.AgentFileDir = "agents"
	//}

	if scp.WithoutHTTPActivity > 0 {
		if scp.WithoutHTTPActivity < 15 { //181121_1
			scp.WithoutHTTPActivity = 15
		}
	}

	if scp.CheckWithoutHTTPActivityAfter < 100 { //181121_1
		scp.CheckWithoutHTTPActivityAfter = 100
	}

	if scp.CleanUpNotDoneRequestStorage < 50 { //181228_2
		scp.CleanUpNotDoneRequestStorage = 50
	}
	//---------------------

	sessCP = &SessConfigParams{}
	*sessCP = *scp //setting the global (in the packet) variable
	//-------scp

	//201204 07:58
	if createGeneralLog(); err != nil {
		//kerr.SysErrPrintf("createGeneralLog err=%v", err.Error())
		return
	} else {
		gLog.run()
	}
	SendToGenLog("init()(ksess)", "general log created")

	//________________

	initGlobalNotDone()
	flr, err = createFeeler(mx)
	if err != nil {
		sessCP = nil
		err = errors.New(fmt.Sprintf("CreateHub: failure of creating feeler with err: %v", err.Error()))
		return
	}

	kerr.PrintDebugMsg(false, "DFLAG201204_0638", fmt.Sprintf("CreateHub: before if err = createUsersLog()"))
	if err = createUsersLog(); err != nil { //see rules LOGGING
		sessCP = nil
		err = errors.New(fmt.Sprintf("CreateHub: failure of creating users log with err: %v", err.Error()))
		return
	}
	kerr.PrintDebugMsg(false, "DFLAG201204_0638", fmt.Sprintf("CreateHub: after if err = createUsersLog()"))

	//reqMultiplexer.HandleFunc("/login", loginHandler)
	//reqMultiplexer.HandleFunc("/logout", logoutHandler)
	//reqMultiplexer.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
	//	serveWs(hub, w, r)
	//})

	server = &http.Server{
		ConnState:      connStateHook,
		Addr:           scp.Listening_address,
		Handler:        flr,
		ReadTimeout:    0, //60 * time.Second,
		WriteTimeout:   0, //60 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	//server.RegisterOnShutdown(nil)

	go func() {

		if scp.CertFile == "" {
			fmt.Printf("--S--  WITHOUT TLS\n")
			err = server.ListenAndServe()
		} else {
			fmt.Printf("--S--  WITH TLS\n")
			err = server.ListenAndServeTLS(scp.CertFile, scp.KeyFile)
		}
		kerr.SysErrPrintf("server.ListenAndServe stopped with message %s", err.Error())
		mess := fmt.Sprintf("--S-- server.ListenAndServe stopped with message %s", err.Error())
		close(ServerStopped)
		serverStopped <- mess
	}()
	time.Sleep(time.Millisecond * 50)
	select {
	case mess := <-serverStopped:
		err = fmt.Errorf("server.ListenAndServeTLS not start; err=%s", mess)
		return
	default:
	}

	hub = &sessHub{
		outChan: make(chan toClients),
		clients: make(map[*sessClient]bool),
	}
	go hub.run()

	if errRS := restoreSessions(); errRS != nil {
		err = fmt.Errorf("CreateHub: restoreSessions err = %v", errRS.Error())
	} else {
		SendToGenLog("CreateHub", "restoreSessions has been done successfully")
	}

	return
} //CreateHub
