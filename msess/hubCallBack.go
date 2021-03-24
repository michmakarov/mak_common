// hubCallBack
//Contains pablic types of callback function that that establishes by the CreateHub function
//and corresponding not public global variables
package msess

import (
	"errors"
	"fmt"

	"net"
	"net/http"
	"time"

	"mak_common/kerr"
)

//201214 07:03 see also rels readme.txt 201214_
//it is an attempt to find a way to understand when and in what condition
//an incoming request enters the handler.
type ConnStateDescr struct {
	conn  net.Conn
	descr string
}

var serverStopped chan string = make(chan string)
var currConnStateDescr *ConnStateDescr = &ConnStateDescr{nil, ""}

//210101 if sessCP.Debug!=2 it does nothing
func connStateHook(conn net.Conn, state http.ConnState) {
	if !byteSet(byte(sessCP.Debug), 2) {
		return
	}
	if currConnStateDescr.conn != conn {
		currConnStateDescr = &ConnStateDescr{
			conn: conn,
			//descr: "",
		}

	}
	currConnStateDescr.descr = fmt.Sprintf("%v-%v-%v", currConnStateDescr.descr, state, time.Now().Format("15:04:05.000"))
	return
}

//____________________201214 07:03

//It is a dispatcher which
//maps a incoming web socket message to the outcoming map accordong to SCEX
type DoInWsMess func(mess map[string]string)

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
	cb_doInWsMess        DoInWsMess           //1
	checkUserCredentails CheckUserCredentails //2
	reqMultiplexer       http.Handler         //*mux.Router          //3
	checkURLPath         URLPathChecker       //4
	//getInitData          GetInitData          //5
)

var flr *feeler

func CreateHub(doInWsMess DoInWsMess, //1 may be nill
	cuc CheckUserCredentails, //2 not nill
	mx http.Handler, //201222 16:25 //mx *mux.Router, //3 not nill
	URLCheker URLPathChecker, //4 not nil
	scp *SessConfigParams) (err error) { //CreateHub body
	time.Sleep(time.Second) //191223 For what is it? 210322 Exposure time for ending setting in other goroutines
	//1 (callback)
	cb_doInWsMess = doInWsMess
	//2 (callback)
	if cuc == nil {
		err = errors.New("CreateHub: no function for checking credentials")
		return
	} else {
		checkUserCredentails = cuc
	}

	//3 (callback)
	if mx == nil {
		err = errors.New("CreateHub: no handler for incoming requests")
		return
	} else {
		reqMultiplexer = mx
	}

	//4 (callback)
	if URLCheker == nil {
		err = errors.New("CreateHub: no URLCheker")
		return
	} else {
		checkURLPath = URLCheker
	}

	//scp------

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

	if scp.WithoutHTTPActivity < 15 { //181121_1
		scp.WithoutHTTPActivity = 15
	}

	if scp.ServerReadTimeout < 1 {
		scp.ServerReadTimeout = 1
	}

	if scp.CleanUpNotDoneRequestStorage < 100 {
		scp.CleanUpNotDoneRequestStorage = 100
	}

	if scp.CallBakTimeout < 100 {
		scp.CallBakTimeout = 100
	}

	//sessCP = &SessConfigParams{}
	*sessCP = *scp //setting the global (in the packet) variable

	//-------scp

	//201204 07:58
	if createGeneralLog(); err != nil {
		//kerr.SysErrPrintf("createGeneralLog err=%v", err.Error())
		return
	} else {
		gLog.run()
	}
	SendToGenLog("init()(msess)", "general log created")

	//________________

	initGlobalNotDone()
	flr, err = createFeeler(mx)
	if err != nil {
		sessCP = nil
		err = errors.New(fmt.Sprintf("CreateHub: failure of creating feeler with err: %v", err.Error()))
		return
	}

	if err = createUsersLog(); err != nil { //see rules LOGGING
		sessCP = nil
		err = errors.New(fmt.Sprintf("CreateHub: failure of creating users log with err: %v", err.Error()))
		return
	}

	startAgentMonitor()

	server = &http.Server{
		ConnState:      connStateHook,
		Addr:           scp.Listening_address,
		Handler:        flr,
		ReadTimeout:    time.Second * time.Duration(sessCP.ServerReadTimeout),
		WriteTimeout:   0,
		MaxHeaderBytes: 1 << 20,
		ErrorLog:       httpServerLog,
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
		close(ServerStopped) //210323 16:43; for what? Idiot! It is public!
		serverStopped <- mess
	}()
	time.Sleep(time.Millisecond * 50) //210323 16:38; maybe after the delay the server will stop suddenly.

	select {
	case mess := <-serverStopped:
		err = fmt.Errorf("server.ListenAndServeTLS not start; err=%s", mess)
		return
	default: // The server has not stopped and the function ends with no error
	}

	//if errRS := restoreSessions(); errRS != nil {
	//	err = fmt.Errorf("CreateHub: restoreSessions err = %v", errRS.Error())
	//} else {
	//SendToGenLog("CreateHub", "restoreSessions has been done successfully")
	//}

	return
} //CreateHub
