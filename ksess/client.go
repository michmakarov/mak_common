package ksess

import (
	//"bytes"
	"container/list"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	_ "github.com/mattn/go-sqlite3"

	"mak_common/kerr"
	"mak_common/khttputils"
	"mak_common/kutils"
)

const (
	startFormat = "20060102_150405"
)

type sessActivityWS struct {
	messCount int64 //The counter of incoming messages
	//maxNotDone int
	notDone *list.List
}

//GetStartFormat returns string that represents the format of timt.Time common for logging
func GetStartFormat() string {
	return startFormat
}

func (wsA sessActivityWS) wsActivity() string {
	return fmt.Sprint(wsA.messCount, "(notDone:", wsA.notDone.Len(), ")")
}

func (wsA sessActivityWS) wsActivityList() string {
	var al string

	for e := wsA.notDone.Front(); e != nil; e = e.Next() {
		al = al + "\n" + e.Value.(*userLogRecord).String()
	}
	return al
}

type sessActivityHTTP struct {
	reqCount int64 //The counter of incoming requests. It containes not dane requests.
	//maxNotDone int
	notDone *list.List
}

//func (hA sessActivityHTTP) httpActivity() string {
//	changeClient.Lock()
//	defer changeClient.Unlock()
//	return fmt.Sprint(hA.reqCount, "(notDone:", hA.notDone.Len(), ")")
//}

//
func (hA sessActivityHTTP) shortReport() string {
	return fmt.Sprintf("Всего:%v; из них не выполненных:%v", hA.reqCount, hA.notDone.Len())
}

func (hA sessActivityHTTP) fullReport(nl string) string {
	var fr string
	var in int // item number
	if hA.notDone.Len() == 0 {
		return fmt.Sprintf("HTTp активность: Всего:%v; из них не выполненных:%v", hA.reqCount, 0)
	}
	fr = fmt.Sprintf("HTTP активность: Всего:%v; из них не выполненны:%v", hA.reqCount, hA.notDone.Len())
	for e := hA.notDone.Front(); e != nil; e = e.Next() {
		in++
		fr = fr + nl + "________" + strconv.Itoa(in) + ") " + e.Value.(*userLogRecord).String()
	}
	return fr
}

func (wA sessActivityWS) fullReport(nl string) string {
	var fr string
	var in int = 1 // item number
	if wA.notDone.Len() == 0 {
		return fmt.Sprintf("WS активность: Всего:%v; из них не выполненных:%v", wA.messCount, 0)
	}
	fr = fmt.Sprintf("WS активность: Всего:%v; из них не выполненны:%v", wA.messCount, wA.notDone.Len())
	for e := wA.notDone.Front(); e != nil; e = e.Next() {
		in++
		fr = fr + "br" + strconv.Itoa(in) + ") " + e.Value.(*userLogRecord).String()
	}
	return fr
}

type sessClient struct {
	hub *sessHub

	// The websocket connection.
	conn *websocket.Conn

	// Buffered channel of outbound messages.
	send chan []byte

	Since    time.Time //181003
	LastHTTP time.Time //181128 The time the last HTTP request being received
	LastUrl  string    //190702 The last HTTP url - for what it needs since it has written into the front log?
	User_ID  int
	Tag      int //The tag uniquely identifiers different in time sessions with the same user id
	//It is the answer to situation (for example) when somebody has logouted from host 1, then has logined from host 2, then again sends requests from host1.

	activityHTTP *sessActivityHTTP
	activityWS   *sessActivityWS

	RemoteHTTP string //The address the registration has been tied
	Host       string //The host the registration has been tied
	//181231 What is it? Is it the server as it is appeared to the agent?

	InitData interface{} //An arbitrary data that have assosiated with the user on creation the session
	//It is set with the CreteHub through calls a callback function SetInitData and may be get by
	// function GetSessInitData
	Data interface{} //An arbitrary data have assosiated with the user
	//See SetSessData and GetSessData
}

func newClient(user_id, tag int, remoteHTTPAddr string, host string) (c *sessClient) {
	var aH = &sessActivityHTTP{}
	var aW = &sessActivityWS{}
	aH.notDone = list.New()
	aW.notDone = list.New()
	c = &sessClient{}
	c.Since = time.Now()
	c.hub = hub
	c.User_ID = user_id
	c.Tag = tag
	c.activityHTTP = aH
	c.activityWS = aW
	c.RemoteHTTP = remoteHTTPAddr
	c.Host = host
	return
}
func (c *sessClient) pongHandler(appData string) error {
	if sessCP.Debug != 0 {
		fmt.Printf("Pong: %s\n", appData)
	}
	c.conn.SetReadDeadline(time.Now().Add(time.Second * time.Duration(sessCP.PongWait))) //this checks that the connection is living yet
	return nil
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
}
var (
	//newline = []byte{'\n'}
	//space = []byte{' '}
	//zeroTime     time.Time
	changeClient sync.Mutex
	setConn      sync.Mutex
	//openUsersLogResultMess string //201203 15:30
)

func (c *sessClient) String(nl string) (s string) {
	var (
		ws             string
		wAc            string
		hAc            string
		initData, data string
		sinceDur       string //181003
	)
	if c == nil {
		return "? Session is nill"
	}
	if c.InitData != nil {
		initData = "Постоянные данные есть;"
	} else {
		initData = "Постоянных данных нет;"
	}
	if c.Data != nil {
		data = "Переменные данные есть;"
	} else {
		data = "Переменных данных нет;"
	}

	changeClient.Lock()
	defer changeClient.Unlock()

	if c.conn == nil {
		ws = "(Нет WS)"
	} else {
		ws = fmt.Sprintf("WS с %v", c.conn.RemoteAddr().String())
	}

	//hAc = c.activityHTTP.fullReport(nl)
	hAc = GetPerformingChoresOfUser(strconv.Itoa(c.User_ID), nl) //181228
	//wAc = c.activityWS.fullReport(nl)

	sinceDur = fmt.Sprintf("%v ...%v", c.Since.Format("20060102 15:04:05"), time.Since(c.Since))

	s = fmt.Sprintf("%v(%v)%v %v%v(зарегистрирован с %v--%v)%v",
		c.User_ID, c.Tag, ws, initData, data, c.RemoteHTTP, sinceDur, nl)
	s = s + hAc + nl + wAc + nl
	return s
}

func (c *sessClient) String_181128(nl string) (s string) { //181128 HTTP idlemess was foisted
	var (
		ws             string
		wAc            string
		hAc            string
		initData, data string
		sinceDur       string //181003
		noHTTPDur      string //181128
	)
	if c == nil {
		return "? Session is nill"
	}
	if c.InitData != nil {
		initData = "Постоянные данные есть;"
	} else {
		initData = "Постоянных данных нет;"
	}
	if c.Data != nil {
		data = "Переменные данные есть;"
	} else {
		data = "Переменных данных нет;"
	}

	changeClient.Lock()
	defer changeClient.Unlock()

	if c.conn == nil {
		ws = "(Нет WS)"
	} else {
		ws = fmt.Sprintf("WS с %v", c.conn.RemoteAddr().String())
	}

	hAc = c.activityHTTP.fullReport(nl)
	wAc = c.activityWS.fullReport(nl)

	sinceDur = fmt.Sprintf("%v ...%v", c.Since.Format("20060102 15:04:05"), time.Since(c.Since))
	noHTTPDur = fmt.Sprintf("%v ...%v", c.LastHTTP.Format("20060102 15:04:05"), time.Since(c.LastHTTP))

	s = fmt.Sprintf("%v(%v)%v %v%v(зарегистрирован с %v--%v; простой %v)%v",
		c.User_ID, c.Tag, ws, initData, data, c.RemoteHTTP, sinceDur, noHTTPDur, nl)
	s = s + hAc + nl + wAc + nl
	return s
}

func (c *sessClient) String_181128_idleness(nl string) (s string) { //181128 HTTP idlemess was foisted
	var (
		noHTTPDur string //181128
	)
	if c == nil {
		return "? Session is nill"
	}

	changeClient.Lock()
	defer changeClient.Unlock()

	noHTTPDur = fmt.Sprintf("%v ...%v", c.LastHTTP.Format("20060102 15:04:05"), time.Since(c.LastHTTP))

	s = fmt.Sprintf("%v простой %v)%v",
		c.User_ID, noHTTPDur, nl)
	return s
}

func (c *sessClient) isHTTPNotDone() bool {
	if c == nil {
		panic("*sessClient.isHTTPNotDone(): ? Session is nill")
	}
	changeClient.Lock()
	defer changeClient.Unlock()
	return c.activityHTTP != nil
}

func (c *sessClient) isWsNotDone() bool {
	if c == nil {
		panic("*sessClient.isWsNotDone(): ? Session is nill")
	}
	changeClient.Lock()
	defer changeClient.Unlock()
	return c.activityWS != nil
}

func (c *sessClient) isNotDone() bool {
	if c == nil {
		panic("*sessClient.isWsNotDone(): ? Session is nill")
	}
	changeClient.Lock()
	defer changeClient.Unlock()
	return (c.activityWS.notDone.Len() != 0) || (c.activityHTTP.notDone.Len() != 0)
}

func (c *sessClient) user_idAsString() string {
	if c == nil {
		return "?User"
	} else {
		return strconv.Itoa(c.User_ID)
	}
}
func (c *sessClient) tagAsString() string {
	if c == nil {
		return "?Tag"
	} else {
		return strconv.Itoa(c.Tag)
	}
}

// readPump pumps messages from the websocket connection to the hub.
//
// The application runs readPump in a per-connection goroutine. The application
// ensures that there is at most one reader on a connection by executing all
// reads from this goroutine.
func (c *sessClient) readPump() {

	defer func() {
		c.hub.unsetWS(c)
	}()
	c.conn.SetReadLimit(sessCP.MaxMessageSize)

	//180726 !? couses hijacking the comnection
	//c.conn.SetReadDeadline(time.Now().Add(time.Duration(sessCP.PongWait))) //Where is guarantee that a ping is already sent???
	c.conn.SetReadDeadline(time.Now().Add(time.Duration(sessCP.PongWait) * time.Second)) //Where is guarantee that a ping is already sent???

	//c.conn.SetReadDeadline(zeroTime) //no deadline
	c.conn.SetPongHandler(c.pongHandler)
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
				kerr.SysErrPrintf(" readPump() websocket.IsUnexpectedCloseError: %v; user_id", err, c.User_ID)
			}
			break
		}

		//180720 For what is this? Let's get it away
		//message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))

		//Here we must calculate the response with a partial routine, which sends the result to c.hub.outChan
		go calcWSResponse(c, message)
	}
}

// writePump pumps messages from the hub to the websocket connection.
//
// A goroutine running writePump is started for each connection. The
// application ensures that there is at most one writer to a connection by
// executing all writes from this goroutine.
func (c *sessClient) writePump() {

	ticker := time.NewTicker(sessCP.PingPeriod)
	defer func() {
		ticker.Stop()
		c.hub.unsetWS(c)
	}()
	for {
		select {
		case message, ok := <-c.send:
			//c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			//c.conn.SetWriteDeadline(zeroTime) //no deadline
			if !ok {
				// The hub closed the channel.
				if c.conn != nil {
					c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				}
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			//c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			//c.conn.SetWriteDeadline(zeroTime) //no deadline
			pingMess := fmt.Sprintf(" from user %v", c.User_ID)
			if c.conn != nil {
				if err := c.conn.WriteMessage(websocket.PingMessage, []byte(pingMess)); err != nil {
					return
				}
			}
		} //select
	}
}

// ServeWs handles websocket requests from the peer.
//It does not create a client (that is an instance of sessClient type), contrary the instance must be already
//See the sesSession within coocies.go, it does the creation
func serveWs(hub *sessHub, w http.ResponseWriter, r *http.Request) {
	var (
		//sessData SessionData
		err error
		//res      int
		c *sessClient
	)

	kerr.PrintDebugMsg(false, "ws", " serveWs HERE!")

	if c, _, err = getSession(r); err != nil {
		kerr.SysErrPrintf("serveWs : error=%v", err.Error())
		w.WriteHeader(400)
		w.Write([]byte(fmt.Sprintf("serveWs : error=%v", err.Error())))
		return
	}

	if c == nil {
		kerr.SysErrPrintf("serveWs : session does not registered; Request =%v", khttputils.ReqLabel(r))
		w.WriteHeader(400)
		w.Write([]byte(fmt.Sprintf("serveWs : session does not registered; Request =%v", khttputils.ReqLabel(r))))
		return
	}
	if c.conn != nil { //Why? I seemingly have said that the connection would be overrided by a next "/ws"
		kerr.SysErrPrintf("serveWs : session already has WS; user_id=%v", c.User_ID)
		w.WriteHeader(400)
		w.Write([]byte(fmt.Sprintf("serveWs : session already has WS; user_id=%v", c.User_ID)))
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		kerr.SysErrPrintf("serveWs : upgrader.Upgrade error=%v", err.Error())
		return
	}

	setConn.Lock()
	c.conn = conn
	c.send = make(chan []byte, userSendChanLen)
	setConn.Unlock()

	go c.writePump()
	go c.readPump()
}
func parserGuard(user_id int, inMess []byte) (resp map[string]string) {
	defer func() {
		if rec := recover(); rec != nil {
			//kerr.SysErrPrintf("!!calcWSResponse cought err = %v", rec)
			resp = sso10(strconv.Itoa(user_id), kerr.GetRecoverError(rec).Error())
		}
	}()
	resp = parserSocket(user_id, inMess)
	return
}

//this sends a toClients structure to the hub (c.hub.outChan <- toClnts)
func calcWSResponse(c *sessClient, inMess []byte) {
	var (
		start string
		begin time.Time
		//dur         time.Duration
		resp        map[string]string
		ulr         *userLogRecord
		recId       string //idendifier of a record in SQLite
		rA          []string
		ip          string
		port        string
		err         error
		user_id     string
		action_name string
		toClnts     toClients
		outMess     []byte
	)
	defer func() {
		if rec := recover(); rec != nil {
			kerr.SysErrPrintf("!!calcWSResponse cought err = %v", rec)
		}
	}()

	user_id = c.user_idAsString()

	rA = strings.Split(c.conn.RemoteAddr().String(), ":")
	if len(rA) == 2 {
		ip = rA[len(rA)-1]
		port = rA[len(rA)-2]
	} else {
		ip = c.conn.RemoteAddr().String()
		port = "?"
	}
	recId, err = kutils.TrueRandInt()
	if err != nil {
		kerr.SysErrPrintf("calcHTTPResponse: kutils.TrueRandInt() returns error")
	}
	action_name = "unknown"
	begin = time.Now()
	start = begin.Format(startFormat)
	ulr = newUserLogRecord(recId, user_id, c.tagAsString(), "ws", ip, port, action_name, start)

	changeClient.Lock() //Lock ----------------
	insertUserLogRecord(ulr)
	c.activityWS.messCount++
	notDoneEl := c.activityWS.notDone.PushBack(ulr)
	changeClient.Unlock() //Unlock ------------------

	resp = parserGuard(c.User_ID, inMess)
	resp, action_name = checkWSResponse(c, resp)

	//dur = time.Now().Sub(begin)

	outMess, err = json.Marshal(resp)
	if err != nil {
		outMess, err = json.Marshal(sso5(user_id, err.Error()))
		if err != nil {
			kerr.SysErrPrintf("calcWSResponse, nothing sent: sso5 not marshalling; err=%s", err.Error())
			outMess = nil
		}
	}

	changeClient.Lock() //Lock ----------------
	c.activityWS.notDone.Remove(notDoneEl)
	//updateUserLogRecordWs(action_name, int64(dur), int64(len(inMess)), int64(len(outMess)), 1, recId)
	changeClient.Unlock() //Unlock ------------------

	//---------- to whom must It be sent?
	toClnts.users = extractUsers(resp)
	//----------

	toClnts.message = outMess
	c.hub.outChan <- toClnts
} // calcWSResponse

//Checks for according to KSCEX (Kot protocol of data exchange between a server and client)
//If a offense has been found the given map is replaced with the corresponding sso map
//see the functions sso2, ..., sso4, sso6, sso7
func checkWSResponse(c *sessClient, parsed map[string]string) (resp map[string]string, action_name string) {
	var (
		user_id string
		errMess string
	)
	user_id = c.user_idAsString()

	if _, ok := parsed["err_mess"]; ok {
		if parsed["res_code"] == "-1" { //A result of  Sso1 took place
			return parsed, "ParsingErr"
		}
	}

	if parsed["user_id"] == "" {
		errMess = fmt.Sprintf("In the answer there is not the user id field, user_id=%s", user_id)
		return sso2(user_id, errMess), "NotUser_id"
	}

	action_name = parsed["action_name"]
	if action_name == "" {
		errMess = fmt.Sprintf("In the answer there is not the action name field, user_id=%s", user_id)
		return sso3(user_id, errMess), "NotAction_name"
	}

	if parsed["user_id"] != user_id {
		errMess = fmt.Sprintf("In the answer there is mismatch of user_id=%s", user_id)
		return sso4(user_id, errMess), "User_idNotMatch"
	}

	if parsed["source"] == "" {
		errMess = fmt.Sprintf("In the answer there is not source fiel; user_id=%s", user_id)
		return sso6(user_id, errMess), "NoSourceFiedl"
	}

	if _, ok := parsed["err_mess"]; ok {
		errMess = fmt.Sprintf("The answer has the not allowed  err_mess fiel; user_id=%s", user_id)
		return sso7(user_id, errMess), "NotAllowedErr_mesField"
	}

	return parsed, action_name
}

func checkOSM(parsed map[string]string) (err error) {
	var (
		user_id int
	)

	if parsed["user_id"] == "" {
		err = errors.New(fmt.Sprint("In the OSM there is not the user id field"))
		return
	}

	if parsed["action_name"] == "" {
		err = errors.New(fmt.Sprint("In the OSM there is not the action name field"))
		return
	}

	user_id, err = strconv.Atoi(parsed["user_id"])
	if err != nil {
		err = errors.New(fmt.Sprint("In the OSM user_id==%s has bad format", parsed["user_id"]))
		return
	}
	if user_id < 0 {
		err = errors.New(fmt.Sprint("In the OSM user_id==%s < 0", user_id))
		return
	}

	if parsed["source"] == "" {
		err = errors.New(fmt.Sprintf("In the OSM there is not source fiel; user_id=%s", user_id))
		return
	}

	if _, ok := parsed["err_mess"]; ok {
		err = errors.New(fmt.Sprintf("The answer has the not allowed  err_mess fiel; user_id=%s", user_id))
		return
	}

	return
}

//201203 04:18 What do the func do?
//Case 1 - users!=nil and len(users)>0 - the message must be sent to all contained users
//Case 2 - users!=nil and len(users)=0 - the message must be sent to all registered users
//Case 3 - users==nil - the message cannot be sent any users
//The case 1  with single user (len(users)==1) take place if there if a field "user_id" and it is a valid integer > -1
//Othewise the case 3 take place.
//The case 2 impossible here.
func extractUsersDefault(answer map[string]string) (users []int) {
	var (
		user_id int
		err     error
		ok      bool
	)
	if _, ok = answer["action_name"]; ok {
		if _, ok = answer["err_msg"]; ok {
			kerr.SysErrPrintf("extractUsersDefault: simultaneous  existing of fields action_name and err_msg is not allowed\n")
			return // the case 3
		}
	}

	if _, ok = answer["user_id"]; ok {
		user_id, err = strconv.Atoi(answer["user_id"])
		if err != nil {
			kerr.SysErrPrintf("extractUsersDefault: strconv.Atoi error=%v\n", err.Error())
			return // the case 3
		} else {
			users = make([]int, 1)
			users[0] = user_id
			return //the case 1
		}
	} else { // no "user_id"
		kerr.SysErrPrintf("extractUsersDefault: no user_id")
		return // the case 3
	}

}

func defaultURLCheker(path string) bool {
	return strings.HasPrefix(path, "/app")
}

//For what does the function need?
func calcHTTPResponse(c *sessClient, w http.ResponseWriter, r *http.Request, cancel context.CancelFunc) {
	var (
		ulr   *userLogRecord
		start string
		begin = time.Now()
		//dur          time.Duration
		ip, port     string
		user_id, tag string
		recId        string //idendifier of a record in SQLite
		chr          *Chore
		//err          error
	)

	kerr.PrintDebugMsg(false, "ServeHTTP_201203_1129", fmt.Sprintf("calcHTTPResponse:very start; c=%v", c))

	//start = time.Now().Format(startFormat)
	start = begin.Format(startFormat)

	ip, port = khttputils.Grt_IP_Port(r)

	//recId, err = kutils.TrueRandInt() //201203 20:45 Now in it there is not some need
	//if err != nil {
	//	kerr.SysErrPrintf("calcHTTPResponse: kutils.TrueRandInt() returns error")
	//}

	if c != nil {
		user_id = c.user_idAsString()
		tag = c.tagAsString()
	} else {
		user_id = "?"
		tag = "?"
	}
	changeClient.Lock() //Lock 201203 11:57
	ulr = newUserLogRecord(recId, user_id, tag, "http", ip, port, r.URL.String(), start)
	changeClient.Unlock() //Unlock  201203 11:57

	//kerr.PrintDebugMsg(false, "ServeHTTP_201203_1129", fmt.Sprintf("calcHTTPResponse: before globalNotDone.AddHTTPChore; ulr=%v", ulr))

	chr = globalNotDone.AddHTTPChore(ulr, w, r, cancel)

	<-chr.doneChan

	ulr.dur = int64(time.Now().Sub(begin))

	//kerr.PrintDebugMsg(false, "ServeHTTP_201203_1129", fmt.Sprintf("calcHTTPResponse: after chr.doneChan; dur=%v; chr=%v", dur, chr))

	changeClient.Lock() //Lock ----------------
	//updateUserLogRecord(int64(dur), -1, -1, 1, recId)
	insertUserLogRecord(ulr)
	changeClient.Unlock() //Unlock ------------------

}

func (c *sessClient) String_190704() (s string) {
	var (
		ws string
		//wAc            string
		hAc            string
		initData, data string
		sinceDur       string
	)
	if c == nil {
		return "? Session is nill"
	}
	if c.InitData != nil {
		initData = "ID is;"
	} else {
		initData = "ID is not;"
	}
	if c.Data != nil {
		data = "VD is;"
	} else {
		data = "VD is not;"
	}

	changeClient.Lock()
	defer changeClient.Unlock()

	if c.conn == nil {
		ws = "(Нет WS)"
	} else {
		ws = fmt.Sprintf("WS с %v", c.conn.RemoteAddr().String())
	}

	//hAc = c.activityHTTP.fullReport(nl)
	hAc = GetPerformingChoresOfUser(strconv.Itoa(c.User_ID), "--")
	//wAc = c.activityWS.fullReport(nl)

	sinceDur = fmt.Sprintf("%v ...%v", c.Since.Format("20060102 15:04:05"), time.Since(c.Since))

	s = fmt.Sprintf("%v(%v)%v %v%v(с %v(%v))--",
		c.User_ID, c.Tag, ws, initData, data, c.RemoteHTTP, sinceDur)
	s = s + hAc + "--" //+ wAc
	return s
}
