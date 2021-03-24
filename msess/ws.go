package msess

import (
	//"bytes"
	"container/list"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	//"strconv"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	_ "github.com/mattn/go-sqlite3"

	"mak_common/kerr"
)

const MaxMessageSize = 4000

// Time allowed to read the next pong message from the peer, seconds
const PongWait = 10 //seconds
type WsMess map[string]string

// Send pings to peer with this period. Must be less than pongWait.
var PingPeriod time.Duration = (PongWait * 9) / 10
var inWsMessChan chan WsMess //All agents sends here incoming message

type sessActivityHTTP struct {
	reqCount int64 //The counter of incoming requests. It containes not dane requests.
	//maxNotDone int
	notDone *list.List
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
}

//this checks that the connection is living yet
//func pongHandler(appData string) error {
//	c.conn.SetReadDeadline(time.Now().Add(time.Second * time.Duration(sessCP.PongWait)))
//	return nil
//}

//The readPump is launched by serveWs in distinct goroutine
//It converts received messages of type 1 and 2 into WsMess and send the result to the inWsMess channel
//
func (a *Agent) readPump() {
	var inwm WsMess = make(map[string]string)
	var errVal string
	var err error

	defer func() {
		a.conn.Close()
		a.WsOut = nil
		a.conn = nil
	}()
	a.conn.SetReadLimit(MaxMessageSize)

	//Where is guarantee that a ping is already sent???
	a.conn.SetReadDeadline(time.Now().Add(time.Duration(PongWait) * time.Second))

	//c.conn.SetReadDeadline(zeroTime) //no deadline
	//a.conn.SetPongHandler(pongHandler)
	for {
		mesType, message, err := a.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
				kerr.SysErrPrintf(" readPump() websocket.IsUnexpectedCloseError: %v; tag=%v\n", err, a.Tag)
			} else {
				kerr.SysErrPrintf(" readPump() ReadMessage error: %v; tag=%v\n", err, a.Tag)
			}
			break
		} else {
			if (mesType != websocket.BinaryMessage) || (mesType != websocket.TextMessage) {
				continue
			}
		}
		if mesType == websocket.BinaryMessage {
			err = fmt.Sprintf("From tag=%v;user=%v binary data was received")
		}

		inWsMessChan <- inwm
	}
}

// writePump pumps messages from the hub to the websocket connection.
//
// A goroutine running writePump is started for each connection. The
// application ensures that there is at most one writer to a connection by
// executing all writes from this goroutine.
func (a *Agent) writePump() {

	ticker := time.NewTicker(PingPeriod)
	defer func() {
		ticker.Stop()
		a.conn.Close()
		a.conn = nil
		a.WsOut = nil
	}()
	for {
		select {
		case message, ok := <-a.WsOut:
			//c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			//c.conn.SetWriteDeadline(zeroTime) //no deadline
			if !ok {
				// The hub closed the channel.
				if a.conn != nil {
					a.conn.WriteMessage(websocket.CloseMessage, []byte{})
				}
				kerr.SysErrPrintf("(a *Agent) writePump(); not ok\n")
				return
			}

			w, err := a.conn.NextWriter(websocket.TextMessage)
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

// ServeWs handles websocket requests from some agent.
//It is a helper function of func (f *feeler) ServeHTTP as only the last detects that the request has come from some agent
func serveWs(w http.ResponseWriter, r *http.Request, a *Agent) {
	var err error

	//kerr.PrintDebugMsg(false, "ws", " serveWs HERE!")

	if a.conn != nil { //Why? I seemingly have said that the connection would be overrided by a next "/ws"
		//kerr.SysErrPrintf("serveWs : session already has WS; user_id=%v", c.User_ID)
		w.WriteHeader(400)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Write([]byte(fmt.Sprintf("serveWs : For agent (%v) it is permitted only one ws connection.", a.Tag)))
		return
	}

	a.conn, err = upgrader.Upgrade(w, r, nil)
	if err != nil {
		kerr.SysErrPrintf("serveWs : upgrader.Upgrade error=%v", err.Error())
		//w.WriteHeader(500)
		//w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		//w.Write([]byte(fmt.Sprintf("serveWs : upgrader.Upgrade error=%v", err.Error())))
		return
	}

	go a.writePump()
	go a.readPump()
}

//210319 14:23
//makeCopyAndCheck copies the mess to the copyMess then checks the mess for satisfaction of rules (see --WSMESS)
//if err!=nil then copyMess==nil
//210321 14:55 The func must be concurrent. Is it so?
func makeCopyAndCheck(mess WsMess) (copyMess WsMess, err error) {
	var ok bool
	var val string
	copyMess = make(WsMess)  //I hope that build in func are concurrent. That is I hope that developers of the language are not idiots
	for k, v := range mess { //I think the thereis not a large grief if not very actual data will be read
		copyMess[k] = v
	}
	val, ok = copyMess["err"] //	As about the make.
	if ok && (val != "") {
		return
	}

	if val, ok = copyMess["action"]; !ok {
		err = fmt.Errorf("makeCopyAndCheck: there is not a key \"action\"")
		copyMess = nil
		return
	}
	if val, ok = copyMess["from"]; !ok {
		err = fmt.Errorf("makeCopyAndCheck: there is not a key \"from\"")
		copyMess = nil
		return
	} else {
		if err = checkAgentAddr(val); err != nil {
			err = fmt.Errorf("makeCopyAndCheck: there is problem with a key \"from\"; err=%v", err.Error())
			copyMess = nil
			return
		}
	}

	if val, ok = copyMess["to"]; !ok {
		err = fmt.Errorf("makeCopyAndCheck: there is not a key \"to\"")
		copyMess = nil
		return
	} else {
		if err = checkAgentAddr(val); err != nil {
			err = fmt.Errorf("makeCopyAndCheck: there is problem with a key \"to\"; err=%v", err.Error())
			copyMess = nil
			return
		}
	}
	return
} //makeCopyAndCheck

func checkAgentAddr(addr string) (err error) {
	var fields []string
	fields = strings.Split(addr, ":")
	if fields[0] != "tag" || fields[0] != "user" {
		err = fmt.Errorf("checkAgentAddr:Instead \"tag\" or \"user\" field there is \"%v\"", fields[0])
		return
	}
	if len(fields) < 2 {
		err = fmt.Errorf("checkAgentAddr: there is not \":\" separator")
		return
	}
	if fields[1] == "" {
		err = fmt.Errorf("checkAgentAddr: there is not value of tag or user id")
		return
	}
	return
}
