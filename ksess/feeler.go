package ksess

import (
	//"errors"
	//"database/sql"
	"fmt"
	"strings"

	"mak_common/kerr"
	"mak_common/khttputils"

	"context"
	//"encoding/json" //see History 201203 06:46
	"log"
	"net/http"
	"os"
	"runtime/debug"
	"strconv"
	"sync/atomic"
	"time"
)

//CtxParType is the type of context parameters which assigning to incoming requests
//
type ctxParType string

const (
	//Next keys are established by func (f *feeler) ServeHTTP
	UserIdCtxKey ctxParType = "UserId"
	NumberCtxKey ctxParType = "Number"
	URLCtxKey    ctxParType = "URL"
)

//GetCtxStrPar returns ok==true if there is a parameter corresponded ctxKey
//The val is the value of the parameter or val=="" if ok==false
//201203 06:37
func GetCtxStrPar(ctx context.Context, ctxKey string) (val string, ok bool) {
	switch ctxKey {
	case "UserId":
		val, ok = ctx.Value(UserIdCtxKey).(string)
	case "Number":
		val, ok = ctx.Value(NumberCtxKey).(string)
	case "URL":
		val, ok = ctx.Value(URLCtxKey).(string)
	default:
		val = ""
		ok = false
	}
	return
}

/* see History 201203 06:46
// Control request answers
type HttpPingAnswer struct {
	PingTag      string
	From         string
	Answertime   string
	RequestCount int64
}
type HttpCloseKotAnswer struct {
	PingTag      string
	From         string
	Answertime   string
	RuquestCount int64
	Message      string
}
*/

//what is the Feeler
//It is the front filter of HTTP requests
//That is that the Feeler analysing incoming requests and rejecting all besides the allowed ones
type feeler struct {
	h           http.Handler
	feelerCount int64
	guardianTag string
	flgr        *feelerLogger
}

func (f *feeler) feelerCountAsString() string {
	return strconv.FormatInt(f.feelerCount, 10)
}

type requestRecord struct {
	count     int64
	start     time.Time
	label     string
	user_id   int
	what      string //"refused" or ''accepted"
	connState string
}

var (
	err            error
	printToConsole bool
	flrLogFileName string
)

//GetFrontLogName returns  a current front log (of the Feeler) file name
//201203 07:09
func GetFrontLogName() string {
	if flr == nil {
		panic("ksess.GetFrontLogName: the ksess feeler is not created yet.")
	}
	return flrLogFileName
}
func createFeeler(h http.Handler) (f *feeler, err error) {
	//var FlrLogFileName string

	if !sessCP.NotAgentDebugging {
		if err = setEmptyAgents(); err != nil {
			return
		}
	}

	if sessCP.Debug != 0 {
		printToConsole = true
	}
	f = &feeler{}
	f.h = h
	flrLogFileName = "Feeler" + time.Now().Format("20060102_150405") + ".log"

	if f.flgr, err = createFlrLog(flrLogFileName, printToConsole); err != nil {
		return nil, err
	} else {
		go f.flgr.Run()
		//return f, nil
	}

	SendToGenLog("Feeler", " started")
	return f, nil
}

/* 201203 06:46
func (f *feeler) checkCommandRequst(w http.ResponseWriter, r *http.Request) (ok bool) {
	var controlPassword string
	if controlPassword = strings.TrimSpace(r.FormValue("CONTROL_PASSWORD")); controlPassword == "" {
		w.WriteHeader(400)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Write([]byte("There is not CONTROL_PASSWORD parameter or it is empty"))
		return
	}

	if controlPassword != sessCP.ControlPassword {
		w.WriteHeader(400)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Write([]byte("The parameter CONTROL_PASSWORD does not equal sessCP.ControlPassword"))
		return
	}

	ok = true
	return
}

func (f *feeler) checkGuardianTag(w http.ResponseWriter, r *http.Request) (ok bool) {
	var guardianTag string
	guardianTag = strings.TrimSpace(r.FormValue("GUARDIAN_TAG"))

	if f.guardianTag == "" {
		f.guardianTag = guardianTag
	} else {
		if f.guardianTag != guardianTag {
			w.WriteHeader(400)
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.Write([]byte("The parameter GUARDIAN_TAG does not equal the established"))
			return
		}
	}
	ok = true
	return
}
*/

func (f *feeler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var (
		cookData      sessCookieData
		c             *sessClient
		cancel        context.CancelFunc
		ctx           context.Context
		err           error
		requestCouter int64 //!!! 190820_2 The problem of requests counter
	)
	defer func() {
		if rec := recover(); rec != nil {
			//kerr.PrintDebugMsg(false, "checker", "ServeHTTP panic")
			kerr.SysErrPrintf("feeler cought err = %v", rec)
			if sessCP.Debug != 0 {
				debug.PrintStack()
			}
			w.WriteHeader(500)
			w.Write([]byte(fmt.Sprintf("feeler cought err = %v", rec)))
		}
	}()

	var WriteToLog = func(do string) { //180808 Why does not it  have a parameter???
		if sessCP.NotFeelerLogging {
			return
		}
		//rr := requestRecord{count, time.Now(), khttputils.ReqLabel(r), cookData.UserID, do}
		//!!! 4) A request counter problem(ksess_190820:190823; ksodd_190819:190823)
		//rr := requestRecord{atomic.LoadInt64(&f.feelerCount), time.Now(), khttputils.ReqLabel(r), cookData.UserID, do}
		rr := requestRecord{requestCouter, time.Now(), khttputils.ReqLabel(r), cookData.UserID, do, currConnStateDescr.descr}
		f.flgr.send <- rr
	}

	/* 201203 06:46

	var WriteControlReqToLog = func(do string) { //181016
		if sessCP.NotFeelerLogging {
			return
		}
		//!!! 4) A request counter problem(ksess_190820:190823; ksodd_190819:190823)
		//rr := requestRecord{f.feelerCount, time.Now(), khttputils.ReqLabel(r), -100, do}
		rr := requestRecord{requestCouter, time.Now(), khttputils.ReqLabel(r), -100, do}
		f.flgr.send <- rr
	}
	*/

	requestCouter = atomic.AddInt64(&f.feelerCount, 1)

	r.Method = strings.ToUpper(r.Method)
	/* //see History 201203 06:46
	if r.URL.Path == "/ping" { //181019//181022//181024//181102
		var ok bool
		var err error
		//var errCode int
		var pingTag string
		//var pa HttpPingAnswer //see History 201203 06:46
		var answer []byte
		if ok = f.checkCommandRequst(w, r); ok {
			WriteControlReqToLog("accepted")
		} else {
			WriteControlReqToLog("refused")
			return
		}
		if ok = f.checkGuardianTag(w, r); !ok {
			return
		}

		pingTag = strings.TrimSpace(r.FormValue("PING_TAG"))
		if pingTag == "" {
			pingTag = "no ping tag"
		}
		pa.PingTag = pingTag
		if pa.From, err = os.Hostname(); err != nil {
			pa.From = err.Error()
		}
		pa.Answertime = time.Now().Format(startFormat)
		pa.RequestCount = f.feelerCount
		answer, _ = json.Marshal(pa)
		w.Header().Set("Contenr-Type", "application/json; charset=utf-8")
		w.Write(answer)
		return
	}

	if r.URL.Path == "/close_kot" { //181024//161102
		var ok bool
		//var err error
		//var errCode int
		var pingTag string
		//var pa HttpCloseKotAnswer //see History 201203 06:46
		var answer []byte

		if ok = f.checkCommandRequst(w, r); ok {
			WriteControlReqToLog("accepted")
		} else {
			WriteControlReqToLog("refused")
			return
		}

		pingTag = strings.TrimSpace(r.FormValue("PING_TAG"))
		if pingTag == "" {
			pingTag = "no ping tag"
		}
		pa.PingTag = pingTag
		pa.Answertime = time.Now().Format(startFormat)
		pa.RuquestCount = f.feelerCount
		pa.Message = closeServer()
		answer, _ = json.Marshal(pa)
		w.Header().Set("Contenr-Type", "application/json; charset=utf-8")
		w.Write(answer)
		return
	} //r.URL.Path == "/close_kot"
	if doControlRequest(w, r, WriteControlReqToLog) {
		return
	}
	*/

	c, cookData, err = getSession(r)
	if err != nil {
		panic("Feeler, GetSession return error =" + err.Error())
	}

	//191223 - hurry forbidden
	if sessCP.HurryForbidden {
		var hurry, user string
		if c == nil {
			user = "?"
		} else {
			user = c.user_idAsString()
		}
		hurry = globalNotDone.URL_InDoing(user, r.URL.Path)
		kerr.PrintDebugMsg(false, "HurryForbidden", fmt.Sprintf("feeler:hurry=%v", hurry))
		if hurry != "" {
			w.WriteHeader(409)
			w.Write([]byte(fmt.Sprintf("<p>No hurry:%v</p>", hurry)))
			return
		}

	}

	kerr.PrintDebugMsg(false, "ServeHTTP_201203_1129", fmt.Sprintf("ServeHTTP:before if doHijackedRequest; cookData=%v, c=%v", cookData, c))

	if doHijackedRequest(w, r, cookData, c) {
		WriteToLog("accepted")
		return
	} else {
		if cookData.UserID > -1 {
			WriteToLog("accepted") //pass to calcHTTPResponse
		} else {
			if !checkURLPath(r.URL.Path) {
				WriteToLog("refused")
				if sessCP.RedirectOnNoAuthorisation == "" { //181019
					w.WriteHeader(401)
					w.Write([]byte(fmt.Sprintf("<p>Access without authorization forbade - %v</p>", khttputils.ReqLabel(r))))
				} else {
					http.Redirect(w, r, sessCP.RedirectOnNoAuthorisation, 303) //Why 303?
				}
				return
			} else {
				WriteToLog("accepted")
				//pass to calcHTTPResponse
			}
		} // else of cookData.UserID>-1
	} //else of doHijackedRequest worked it

	if c != nil {
		c.LastHTTP = time.Now()
		c.LastUrl = r.RequestURI
		updateSavedSess(c) //190702
	} //181128 And question - may it be that another gorouring will want to do the same
	//r = r.WithContext(context.WithValue(r.Context(), NumberCtxKey, f.feelerCountAsString()))
	r = r.WithContext(context.WithValue(r.Context(), NumberCtxKey, strconv.FormatInt(requestCouter, 10)))
	r = r.WithContext(context.WithValue(r.Context(), UserIdCtxKey, cookData.UserIDAsString()))
	r = r.WithContext(context.WithValue(r.Context(), URLCtxKey, r.RequestURI)) //190408
	ctx, cancel = context.WithCancel(r.Context())
	r = r.WithContext(ctx)
	calcHTTPResponse(c, w, r, cancel)
} //(f *Feeler) ServeHTTP

type feelerLogger struct {
	log            *log.Logger
	printToConsole bool
	send           chan requestRecord
}

func createFlrLog(fileName string, printToConsole bool) (FlrLog *feelerLogger, err error) {
	var f *os.File
	//if sessCP.
	FlrLog = &feelerLogger{}
	if f, err = os.Create(sessCP.LogsDir + fileName); err != nil {
		kerr.SysErrPrintf("Не удалось создать lrLog - %v\n ", err.Error())
		if printToConsole {
			fmt.Printf("Не удалось создать lrLog - %v\n ", err.Error())
		}
		FlrLog = nil
		return
	} else {
		FlrLog.log = log.New(f, "", log.LstdFlags)
	}
	FlrLog.printToConsole = printToConsole
	FlrLog.send = make(chan requestRecord, 253)
	return
}

func (fl *feelerLogger) Run() {
	if fl == nil {
		return
	}
	for {
		rr := <-fl.send
		fl.log.Printf("flr: (CNT==%v;user_id=%v)%v -- %v; what: %v \n", rr.count, rr.user_id, rr.label, rr.start, rr.what)
		if fl.printToConsole {
			fmt.Printf("flr: (CNT==%v;user_id=%v)%v -- %v --%v --%v\n ", rr.count, rr.user_id, rr.label, rr.start, rr.what, rr.connState)
		}
	}
}
