// not_done_global_storage
// 210324 11:52 see blabla_210311, Developer_notes 210324 11:24
package msess

import (
	"container/list"
	"context"

	"fmt"
	"mak_common/kerr"
	"net/http"

	//"strings"
	//"sync"
	"time"
)

type GlobalNotDone struct {
	count   int64      //Counter of all chores that have been in the storage: that those had been performed and mayby been removed as well that those are performing
	notDone *list.List //of *Chore
}

type nd_MonitorResult struct {
	Err       error
	ExtraInfo string //210329 11:01 for passing results of scanning the list
	Data      interface{}
}
type nd_MonitorQuery struct {
	Action     string
	Data       interface{}
	ResultChan chan nd_MonitorResult
}

var globalNotDone *GlobalNotDone // it is received a vulue from func initGlobalNotDone()
var nd_mqChan chan nd_MonitorQuery

//A value of type "chore" is registration record of some work (a chore as it is named here)
//that is a function that is performing by a distinct goroutine
//So let us say that A PERFORMER fulfills a work and it is a goroutine.
//
//Who is a employer of a chore? Of course it is a request that have passed the feeler and reaches performing.
//It is represented by "userLogRecord" record that was formed before the work starts.
//
//How can we track that the employer is away now? One must receive a message from some channel!
//Let if this channel is closed the employer is away.
//Who will close it. That who holds the connection. The undelying http server.
//Who will be listenning it? The "clean" method, see further.
//Well
//What must be when a performer has done his work? But what may be done at the moment? Only to send to channel may be done!
//Let it be two such channels for two distinct listener
//Let those be closed when a performer has done his work.
//Who will close them. A peformer.
//Who will be listenning them? The "clean" method (see further) will listen first.
//A caller that calls methods, that create chores, can (and must) listen the second.
//Well
//A chore is only a record that must reflect the state of things Who is tracking this state of things?
//It is GlobalNotDone.clean method that is spinning in its goroutine.
type Chore struct {
	id        int64          //It is a value of  GlobalNotDone.count
	ulr       *userLogRecord // it represents the employer
	err       error          //
	emplAway  <-chan bool    // it is closed the employer through http.ResponseWriter
	cancel    context.CancelFunc
	doneChan  chan struct{}       // foe reading by the monitor
	doneChan2 chan *userLogRecord // foe reading by the calcHTTPResponse
}

func initGlobalNotDone() {
	globalNotDone = &GlobalNotDone{}
	nd_mqChan = make(chan nd_MonitorQuery)
	globalNotDone.notDone = list.New()
	go nd_Monitor()
}

func nd_Monitor() {
	var mQ nd_MonitorQuery
	var mR nd_MonitorResult
	var chore *Chore
	var begin time.Time

	for true {
		//sleep a bit
		time.Sleep(time.Duration(sessCP.CleanUpNotDoneRequestStorage) * time.Millisecond)

		select { //serving monitor queries and scanning the storage
		case mQ = <-nd_mqChan:
			switch mQ.Action {
			case "addChore":
				mR = globalNotDone.addChore(mQ.Data)
			case "cancelAllForAgent":
				mR = globalNotDone.cancelAllForAgent(mQ.Data)
			default:
				mR.Err = fmt.Errorf("agentsMonitor: illegal action (%v) of a query", mQ.Action)
			} //switch
			mQ.ResultChan <- mR
		default: //serving monitor queries is done, now scanning the storage
			for e := globalNotDone.notDone.Front(); e != nil; e = e.Next() { //scaning the list
				select {
				case <-e.Value.(*Chore).doneChan:
					chore = e.Value.(*Chore)
					if begin, err = time.Parse(timeFormat, chore.ulr.start); err != nil {
						panic("nd_Monitor: calculating the begin err=" + err.Error())
					}
					chore.ulr.dur = fmt.Sprintf("%v", time.Now().Sub(begin))
					chore.ulr.extraInfo = chore.ulr.extraInfo + "done;"
					chore.doneChan2 <- chore.ulr
					globalNotDone.notDone.Remove(e)
				case <-e.Value.(*Chore).emplAway:
					chore = e.Value.(*Chore)
					chore.ulr.extraInfo = chore.ulr.extraInfo + "cancelation of emplAway;"
					e.Value.(*Chore).cancel()
				default:
					// go on
				} //select
			} //for (scanning the storage)

		} //select; end of serving monitor queries and scanning the storage
	} //for
} //agentMonitor

func (gnd *GlobalNotDone) addHTTPChore(ulr *userLogRecord, w http.ResponseWriter,
	r *http.Request, cancel context.CancelFunc) (doneChan2 chan *userLogRecord, err error) {
	var newChore Chore
	var mQ nd_MonitorQuery
	var mR nd_MonitorResult

	newChore.cancel = cancel
	newChore.ulr = ulr
	newChore.emplAway = w.(http.CloseNotifier).CloseNotify()
	newChore.doneChan = make(chan struct{})
	newChore.doneChan2 = make(chan *userLogRecord)
	go func() { //start a performer
		defer close(newChore.doneChan)
		defer func() {
			if rec := recover(); rec != nil {
				newChore.err = kerr.GetRecoverError(rec)
			}
		}()
		reqMultiplexer.ServeHTTP(w, r)
	}()
	mQ = nd_MonitorQuery{"addChore", &newChore, make(chan nd_MonitorResult)}
	nd_mqChan <- mQ
	mR = <-mQ.ResultChan
	err = mR.Err
	doneChan2 = newChore.doneChan2
	return
}

//What kind can a chore be in general? It is defined by properties of its ulr
//OutSess - the tag of session is "?"
//Orphan - the tag of session is not "?" but the session is not registered now
//NotOrphan - the tag of session is not "?" and the session is registered now
//Besides above said a chore can be left or not. That means that the HTTP request that initiated the chore is closed now
//This will be expressed by suffics "Closed' e.g. "OutSessClosed"
//func (gnd *GlobalNotDone) StringMBD(nl string) string {

//190124

//210329 10:31
func cancelAll(a *Agent) {
	var mQ nd_MonitorQuery
	var mR nd_MonitorResult
	mQ = nd_MonitorQuery{"cancelAllForAgent", a, make(chan nd_MonitorResult)}
	nd_mqChan <- mQ
	mR = <-mQ.ResultChan
	err = mR.Err
	return
}
