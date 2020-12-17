// not_done_global_storage
package ksess

import (
	"container/list"
	"context"
	"fmt"
	"kot_common/kerr"
	"net/http"
	"strings"
	"sync"
	"time"
)

var globalNotDone = &GlobalNotDone{}
var gndMtx sync.Mutex

type GlobalNotDone struct {
	count   int64 //Counter of all chores that have been in the storage: that those had been performed and mayby been removed as well that those are performing
	NotDone *list.List
}

//type chore struct {//old definition. before 191017
//	id                 int64 //
//	w                  http.ResponseWriter
//	notKeep            bool //see GlobalNotDone.clean
//	start              time.Time
//	ulr                userLogRecord // a copy at the moment of creating the chore
//	err                error
//	RequestSouceClosed bool //190110 How does it coordinate with notKeep flag?
//	//maybe it is well; see clean method 190112 Nonsense! The w as well can be used as a flag
//	doneChan chan bool
//}

//A value of type "chore" is registration record of some work (a chore as it is named here) that is a function that is performing by a distinct goroutine
//So let let us say that a performer fulfills a work and performer is the goroutin
//Who is a employer of some work? It is represented by "userLogRecord" record that was formed before the work starts.
//Well
//How can we track that the employer is away now? One must receive a message from some channel!
//Let if this channel is closed the employer is away.
//Who will close it. That who holds the connection. The undelying http server or  readPump method.
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
	id       int64         //It is a value of  GlobalNotDone.count
	ulr      userLogRecord // it represents the employer
	emplAway <-chan bool   // for listenning. if it is closed the employer is away now
	//notKeep            bool          //see GlobalNotDone.clean
	cancel             context.CancelFunc
	doneChan           chan bool //for writing. if it is closed the work is done
	doneChan2          chan bool //As previous but for another listener
	start              time.Time
	dur                time.Duration
	err                error
	RequestSouceClosed bool //190110 How does it coordinate with notKeep flag?//191223 For what is it?
	//maybe it is well; see clean method 190112 Nonsense! The w as well can be used as a flag
}

func (chr *Chore) String(nl string) (s string) {
	s = s + "ulr:" + chr.ulr.String() + nl
	s = s + "dur:" + chr.dur.String() + nl
	return s
}

func initGlobalNotDone() {
	globalNotDone = &GlobalNotDone{}
	globalNotDone.NotDone = list.New()
	go globalNotDone.clean()
}

//This cleans the storage from chores that have its flag of notKeep = true
func (gnd *GlobalNotDone) clean() {
	var i, remainder int64
	//var scanned int
	var removed int
	for {
		if sessCP != nil { //sleep a bit
			time.Sleep(time.Duration(sessCP.CleanUpNotDoneRequestStorage) * time.Millisecond)
		} else {
			time.Sleep(time.Duration(2000 * time.Millisecond))
		}

		remainder = i % 10
		i++
		//scanned = 0
		removed = 0
		for e := gnd.NotDone.Front(); e != nil; e = e.Next() {
			select {
			case <-e.Value.(*Chore).doneChan2:
				gndMtx.Lock()
				gnd.NotDone.Remove(e)
				removed++
				gndMtx.Unlock()
			case <-e.Value.(*Chore).emplAway:
				gndMtx.Lock()
				e.Value.(*Chore).RequestSouceClosed = true
				gndMtx.Unlock()
			default:
				//break // go on
			} //select
		} //for (iterates the srorage)
		if (remainder == 0) || (removed != 0) {
			//fmt.Printf("--M-- Clean ; i=%v; scanned=%v;removed=%v\n", i, scanned, removed)
		}
	} //infinite loop
}

func (gnd *GlobalNotDone) AddHTTPChore(ulr *userLogRecord, w http.ResponseWriter, r *http.Request, cancel context.CancelFunc) (chrPtr *Chore) {
	var newChore Chore
	//newChore.w = w
	newChore.cancel = cancel
	newChore.start = time.Now()
	newChore.ulr = *ulr
	newChore.emplAway = w.(http.CloseNotifier).CloseNotify()
	newChore.doneChan, newChore.doneChan2 = func() (doneChan chan bool, doneChan2 chan bool) {
		doneChan = make(chan bool)
		doneChan2 = make(chan bool, 1)
		go func() {
			defer func() {
				close(doneChan)
				doneChan2 <- true
				//fmt.Printf("--M-- execution done;dur=%v;id=%v\n", newChore.dur, newChore.id)
			}()
			defer func() {
				if rec := recover(); rec != nil {
					newChore.err = kerr.GetRecoverError(rec)
				}
			}()
			reqMultiplexer.ServeHTTP(w, r)
			newChore.dur = time.Since(newChore.start)
		}()
		return
	}()
	gndMtx.Lock()
	gnd.count++
	newChore.id = gnd.count
	gnd.NotDone.PushBack(&newChore) //This will exist until notKeep=false
	gndMtx.Unlock()
	//fmt.Printf("--M-- PUSH has been done ;dur=%v;id=%v\n", newChore.dur, newChore.id)
	return &newChore
}

//What kind can a chore be in general? It is defined by properties of its ulr
//OutSess - the tag of session is "?"
//Orphan - the tag of session is not "?" but the session is not registered now
//NotOrphan - the tag of session is not "?" and the session is registered now
//Besides above said a chore can be left or not. That means that the HTTP request that initiated the chore is closed now
//This will be expressed by suffics "Closed' e.g. "OutSessClosed"
func (gnd *GlobalNotDone) String(nl string) string {
	var orphans string
	var notOrphans string
	var outSess string
	var preamb string
	var closed string

	gndMtx.Lock()
	defer gndMtx.Unlock()

	preamb = fmt.Sprintf("%v", gnd.NotDone.Len())

	for e := gnd.NotDone.Front(); e != nil; e = e.Next() {
		closed = ""
		if e.Value.(*Chore).RequestSouceClosed {
			closed = "(closed!)"
		}
		if e.Value.(*Chore).ulr.tag != "?" {
			if hub.tagRegistered(e.Value.(*Chore).ulr.tag) {
				notOrphans = notOrphans + e.Value.(*Chore).ulr.String() + closed + nl
			} else {
				orphans = orphans + e.Value.(*Chore).ulr.String() + closed + nl
			}
		} else {
			outSess = outSess + e.Value.(*Chore).ulr.String() + closed + nl
		}

	}
	return "Выполняемые запросы(" + preamb + ")" + nl + "Orphans:" + nl + orphans + nl +
		"not Orphans:" + nl + notOrphans + nl + "outSess:" + nl + outSess + nl +
		"________________"
}

//190124
func (gnd *GlobalNotDone) String2(nl string) string {

	var preamb string
	var closed string
	var whoIsIt string
	var list string

	gndMtx.Lock()
	defer gndMtx.Unlock()

	preamb = fmt.Sprintf("%v", gnd.NotDone.Len())

	for e := gnd.NotDone.Front(); e != nil; e = e.Next() {
		closed = ""
		whoIsIt = ""
		if e.Value.(*Chore).RequestSouceClosed {
			closed = "брошен"
		} else {
			closed = "ждут"
		}
		if e.Value.(*Chore).ulr.tag != "?" {
			if hub.tagRegistered(e.Value.(*Chore).ulr.tag) {
				whoIsIt = "не сиротка"
			} else {
				whoIsIt = "сиротка"
			}
		} else {
			whoIsIt = "outSess"
		}
		list = list + fmt.Sprintf("id=%v(%v):%v(%v)%v", e.Value.(*Chore).id, whoIsIt, e.Value.(*Chore).ulr.String(), closed, nl)
		//e.Value.(*Chore).id+ whoIsIt + e.Value.(*Chore).ulr.String() + closed + nl

	}
	return "Выполняемые запросы(" + preamb + ")" + nl + list +
		"________________"
}

func (gnd *GlobalNotDone) StringOfUserMBD(userId string, nl string) string {
	var orphans string
	var notOrphans string
	var outSess string
	var preamb string
	var closed string
	var count int

	gndMtx.Lock()
	//defer gndMtx.Unlock()

	for e := gnd.NotDone.Front(); e != nil; e = e.Next() {
		if e.Value.(*Chore).ulr.user_id != userId {
			continue
		}
		count++
		closed = ""
		if e.Value.(*Chore).RequestSouceClosed {
			closed = "(closed!)"
		}
		if e.Value.(*Chore).ulr.tag != "?" {
			if hub.tagRegistered(e.Value.(*Chore).ulr.tag) {
				notOrphans = notOrphans + e.Value.(*Chore).ulr.String() + closed + nl
			} else {
				orphans = orphans + e.Value.(*Chore).ulr.String() + closed + nl
			}
		} else {
			outSess = outSess + e.Value.(*Chore).ulr.String() + closed + nl
		}

	}
	gndMtx.Unlock()

	preamb = fmt.Sprintf("%v;user==%v", count, userId)

	return "Выполняемые запросы(" + preamb + ")" + nl + "Orphans:" + nl + orphans + nl +
		"not Orphans:" + nl + notOrphans + nl + "outSess:" + nl + outSess + nl +
		"________________"
}

//190123_I181121
func (gnd *GlobalNotDone) StringOfUser_2(userId string, nl string) string {
	var preamb string
	var closed string
	var res string
	var count int

	gndMtx.Lock()

	for e := gnd.NotDone.Front(); e != nil; e = e.Next() {
		if e.Value.(*Chore).ulr.user_id != userId {
			continue
		}
		count++
		closed = ""
		if e.Value.(*Chore).RequestSouceClosed {
			closed = "(closed!)"
		}
		if e.Value.(*Chore).ulr.tag != "?" {
			if hub.tagRegistered(e.Value.(*Chore).ulr.tag) {
				res = res + "не сиротка " + e.Value.(*Chore).ulr.String() + closed + nl
			} else {
				res = res + "сиротка " + e.Value.(*Chore).ulr.String() + closed + nl
			}
		} else {
			res = res + "вне сессии" + e.Value.(*Chore).ulr.String() + closed + nl
		}

	}
	gndMtx.Unlock()

	preamb = fmt.Sprintf("%v;user==%v", count, userId)

	return "Выполняемые запросы(" + preamb + ")" + nl + res + nl +
		"________________"
}

// URL_InDoing returns not empty string if the URL is not performed
//191223 for HurryForbidden
func (gnd *GlobalNotDone) URL_InDoing(userId string, URL string) string {
	//var preamb string
	var closed string
	var res string
	var count int

	var getPath = func(url string) string {
		urlParts := strings.Split(url, "?")
		return urlParts[0]
	}

	gndMtx.Lock()

	for e := gnd.NotDone.Front(); e != nil; e = e.Next() {

		kerr.PrintDebugMsg(false, "HurryForbidden",
			fmt.Sprintf("URL_InDoing:%v<>%v,%v<>%v", getPath(e.Value.(*Chore).ulr.url), URL, e.Value.(*Chore).ulr.user_id, userId))

		if (getPath(e.Value.(*Chore).ulr.url) != URL) || (e.Value.(*Chore).ulr.user_id != userId) {
			continue
		}
		count++
		closed = ""
		if e.Value.(*Chore).RequestSouceClosed {
			closed = "(closed!)"
		}
		if e.Value.(*Chore).ulr.tag != "?" {
			if hub.tagRegistered(e.Value.(*Chore).ulr.tag) {
				res = res + "не сиротка " + e.Value.(*Chore).ulr.String() + closed
			} else {
				res = res + "сиротка " + e.Value.(*Chore).ulr.String() + closed
			}
		} else {
			res = res + "вне сессии" + e.Value.(*Chore).ulr.String() + closed
		}

	}
	gndMtx.Unlock()

	//preamb = fmt.Sprintf("%v;user==%v", count, userId)

	return res
}

func (gnd *GlobalNotDone) choresOfUserCount(userId string) (count int) {
	gndMtx.Lock()
	defer gndMtx.Unlock()

	for e := gnd.NotDone.Front(); e != nil; e = e.Next() {
		if e.Value.(*Chore).ulr.user_id != userId {
			break
		}
		count++
	}
	return count
}

//func GetPerformingChores(nl string) string {
//	return globalNotDone.String(nl)
//}

//201208 18:25 In what is difference with GetPerformingChores?
func GetPerformingChores2(nl string) string {
	return globalNotDone.String2(nl)
}

func GetPerformingChoresOfUser(userId string, nl string) (res string) {
	res = globalNotDone.StringOfUser_2(userId, nl)
	return //globalNotDone.StringOfUser(userId, nl)
}

func GetCountOfPerformingChoresOfUser(userId string) int {
	return globalNotDone.choresOfUserCount(userId)
}

//190124
func (gnd *GlobalNotDone) cancel(id int64) (res bool) {

	gndMtx.Lock()
	defer gndMtx.Unlock()

	for e := gnd.NotDone.Front(); e != nil; e = e.Next() {
		if (e.Value.(*Chore).cancel != nil) && (e.Value.(*Chore).id == id) {
			e.Value.(*Chore).cancel()
			res = true
			return
		}
	}

	return
}

func CancelChore(id int64) (res bool) {
	res = globalNotDone.cancel(id)
	return
}
