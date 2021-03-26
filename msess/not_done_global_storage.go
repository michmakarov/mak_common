// not_done_global_storage
// 210324 11:52 see blabla_210311, Developer_notes 210324 11:24
package msess

import (
	"container/list"
	"context"

	//"fmt"
	"mak_common/kerr"
	"net/http"

	//"strings"
	//"sync"
	"time"
)

var globalNotDone = *GlobalNotDone // it is received a vulue from func initGlobalNotDone()

type GlobalNotDone struct {
	count         int64 //Counter of all chores that have been in the storage: that those had been performed and mayby been removed as well that those are performing
	addChoreChan  chan *Chore
	cancelAllChan chan *Agent //210326 18:33 It cancels all where Chore.ulr.tag==Agent.Tag
	notDone       *list.List  //of *Chore
}

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
	id       int64          //It is a value of  GlobalNotDone.count
	ulr      *userLogRecord // it represents the employer
	err      error          //
	emplAway <-chan bool    // it is closed the employer through http.ResponseWriter
	cancel   context.CancelFunc
	doneChan chan struct{} // it will be closed by a performer
}

func initGlobalNotDone() {
	globalNotDone = &GlobalNotDone{}
	globalNotDone.addChoreChan = make(chan *Chore)
	globalNotDone.cancelAllChan = make(chan *Agent)
	globalNotDone.notDone = list.New()
	go globalNotDone.run()
}

//
func (gnd *GlobalNotDone) run() {
	var i int64
	for {
		//sleep a bit
		time.Sleep(time.Duration(sessCP.CleanUpNotDoneRequestStorage) * time.Millisecond)
		i++
		select {
		case newChore := <-gnd.addChoreChan:
			gnd.notDone.PushBack(newChore)
			gnd.count++
		case a := <-gnd.cancelAll:
			gnd.cancelAll(a)
		default: //go on
		}
		for e := gnd.notDone.Front(); e != nil; e = e.Next() { //scaning the list
			select {
			case <-e.Value.(*Chore).doneChan:
				gnd.notDone.Remove(e)
			case <-e.Value.(*Chore).emplAway:
				e.Value.(*Chore).cancel()
			default:
				// go on
			} //select
		} //for (scanning the storage)
	} //infinite loop
}

func (gnd *GlobalNotDone) addHTTPChore(ulr *userLogRecord, w http.ResponseWriter,
	r *http.Request, cancel context.CancelFunc) (dc chan struct{}) {
	var newChore Chore
	newChore.cancel = cancel
	newChore.ulr = ulr
	newChore.emplAway = w.(http.CloseNotifier).CloseNotify()
	newChore.doneChan = make(chan struct{})
	go func() { //start a performer
		defer close(newChore.doneChan)
		defer func() {
			if rec := recover(); rec != nil {
				newChore.err = kerr.GetRecoverError(rec)
			}
		}()
		reqMultiplexer.ServeHTTP(w, r)
	}()
	gnd.addChoreChan <- &newChore
	return newChore.doneChan
}

//What kind can a chore be in general? It is defined by properties of its ulr
//OutSess - the tag of session is "?"
//Orphan - the tag of session is not "?" but the session is not registered now
//NotOrphan - the tag of session is not "?" and the session is registered now
//Besides above said a chore can be left or not. That means that the HTTP request that initiated the chore is closed now
//This will be expressed by suffics "Closed' e.g. "OutSessClosed"
//func (gnd *GlobalNotDone) StringMBD(nl string) string {

//190124

//210326 17:35
func cancelChores(a *Agent) {

}
