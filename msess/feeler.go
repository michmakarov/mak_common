package msess

import (
	//"errors"
	//"database/sql"
	"fmt"
	//"strings"

	"mak_common/kerr"
	//"mak_common/khttputils"

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

const timeFormat = "20060102_150405"

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

//210323 05:34 it was reworked under the agent conception.
//210101 The requestRecord is incoming request descriptor.
//Why does it need if there is r *http.Request?
//It bear info which there are not in http.Request
type requestRecord struct {
	count       int64
	start       string
	reqDescr    string
	agent       string //agent.shortDecr()
	what        string //"refused" or ''accepted"
	connState   string
	contentType string
}

var (
	err            error
	flrLogFileName string
)

//GetFrontLogName returns  a current front log (of the Feeler) file name
//201203 07:09
func GetFrontLogName() string {
	return flrLogFileName
}

//210101
func createFeeler(h http.Handler) (f *feeler, err error) {
	f = &feeler{}
	f.h = h
	flrLogFileName = "Feeler" + time.Now().Format("20060102_150405") + ".log"

	//kerr.PrintDebugMsg(false, "DFLAG210102", fmt.Sprintf("createFeeler: sessCP.Debug: %v", sessCP.Debug))

	if f.flgr, err = createFlrLog(flrLogFileName, uint8(sessCP.Debug)); err != nil {
		return nil, err
	} else {
		if f.flgr != nil { //210603 06:58 It may be nil if stringSet(sessCP.Loggers, f)!=true
			go f.flgr.Run()
		}
	}

	if f.flgr != nil {
		SendToGenLog("Feeler", " started with logging")
	} else {
		SendToGenLog("Feeler", " started without logging")
	}
	return f, nil
}

func (f *feeler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var (
		//err           error
		//cD            *SessCookieData
		clientErrMess string
		clientErrCode int

		a             *Agent
		cancel        context.CancelFunc
		ctx           context.Context
		requestCouter int64
	)
	defer func() {
		var rec interface{}
		if rec = recover(); rec != nil {
			var msg string
			msg = fmt.Sprintf("feeler panic %v (%v)", rec, getRequestBrief(r))
			kerr.SysErrPrintln(msg)
			if byteSet(sessCP.Debug, 4) {
				debug.PrintStack()
			}
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.WriteHeader(500)
			w.Write([]byte(msg))
		}
	}()

	var WriteToLog = func(do string) { //210325 04:32 //210309 16:46 //210607 18:29
		if f.flgr == nil {
			return
		} //210603 06:46 See LOGGING definition and func createFlrLog
		var reqestDescr string = getRequestDescr(r)
		rr := requestRecord{
			count:       requestCouter,
			start:       time.Now().Format(timeFormat),
			reqDescr:    reqestDescr,
			what:        do,
			connState:   "CS:" + currConnStateDescr.descr,
			contentType: "CT:" + r.Header.Get(http.CanonicalHeaderKey("Content-Type")),
		}
		f.flgr.send <- rr
	}

	kerr.PrintDebugMsg(true, "DFLAG210610", fmt.Sprintf("feeler.ServeHTTP start RequestURI=%v", r.RequestURI))
	requestCouter = atomic.AddInt64(&f.feelerCount, 1)

	// 210607 18:36 I think that a handler may do this it it needs
	//r.Method = strings.ToUpper(r.Method)

	//1. 210607 15:59 intercepting 210609 06:20
	//Each case caurses sending some aswer, which is determined entirely by the handler.
	//The handler has only a request and forms an answer on its own from the request
	switch r.URL.Path {
	case "/":
		WriteToLog("intercepted")
		indexHandler(w, r)
		return
	case "/agent":
		WriteToLog("intercepted")
		agentHandler(w, r)
		return
	case "/ws":
		serveWs(w, r, a)
		WriteToLog("intercepted")
		return
	case "/login":
		login(w, r, a)
		WriteToLog("intercepted")
		return
	case "/logout":
		logout(w, r, a)
		WriteToLog("intercepted")
		return
	}

	//2. Outsession requests pass without any hinder
	if checkURLPathEnv(r.URL.Path) {
		//kerr.PrintDebugMsg(true, "DFLAG210604-0308", fmt.Sprintf("feeler.ServeHTTP after checking RequestURI=%v", r.RequestURI))
		goto gettingResponse //210604 18:21 Now agent==nil as the mark of outsession request
	}

	//3. Determing agent
	if a = agentRegistered(r); a == nil {
		clientErrMess = fmt.Sprintf("No agent\n")
		clientErrCode = 403
		goto exitOnErr
	}

gettingResponse:
	r = r.WithContext(context.WithValue(r.Context(), NumberCtxKey, strconv.FormatInt(requestCouter, 10)))
	if a == nil {
		r = r.WithContext(context.WithValue(r.Context(), UserIdCtxKey, "-1"))
	} else {
		r = r.WithContext(context.WithValue(r.Context(), UserIdCtxKey, a.UserId))
	}
	r = r.WithContext(context.WithValue(r.Context(), URLCtxKey, r.RequestURI)) //190408
	ctx, cancel = context.WithCancel(r.Context())
	r = r.WithContext(ctx)
	if a == nil {
		WriteToLog("OUTSESSION")
	} else {
		WriteToLog("REGULAR")
	}
	//kerr.PrintDebugMsg(true, "DFLAG210604-0308", fmt.Sprintf("feeler.ServeHTTP next calcHTTPResponse; agent=%v", a))
	calcHTTPResponse(f.feelerCount, a, w, r, cancel)
	return

exitOnErr:
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(clientErrCode)
	w.Write([]byte(clientErrMess))
	WriteToLog("refused")

} //(f *Feeler) ServeHTTP

type feelerLogger struct {
	log *log.Logger
	//printToConsole bool
	mode uint8
	send chan requestRecord
}

//210101 createFlrLog returns an error if is not success
func createFlrLog(fileName string, mode uint8) (FlrLog *feelerLogger, err error) {
	var f *os.File
	//kerr.PrintDebugMsg(false, "DFLAG210102", fmt.Sprintf("createFlrLog: mode: %b", mode))
	if stringSet(sessCP.Loggers, "f") != true {
		return
	}
	FlrLog = &feelerLogger{}
	if f, err = os.Create("logs/f" + fileName); err != nil {
		kerr.SysErrPrintf("Не удалось создать lrLog - %v\n ", err.Error())
		if mode > 0 {
			fmt.Printf("Не удалось создать lrLog - %v\n ", err.Error())
		}
		FlrLog = nil
		return
	} else {
		FlrLog.log = log.New(f, "", log.LstdFlags)
	}
	//FlrLog.printToConsole = printToConsole
	FlrLog.mode = mode
	FlrLog.send = make(chan requestRecord, 253)
	return
}

//210101
func (fl *feelerLogger) Run() {
	var rr requestRecord
	var msg string
	if fl == nil {
		return
	}
	for {
		rr = <-fl.send

		msg = fl.getFlrlogMess(rr)
		fl.log.Print(msg)
		if byteSet(fl.mode, 1) {
			fmt.Println(msg)
		}
	}
}

//210323 05:23
//getFlrlogMess converts the rr to a string for writing to the front log.
//210101 //210104 14:12
func (fl *feelerLogger) getFlrlogMess(rr requestRecord) (mess string) {
	var additionalMess string

	//210101 here mess obtains its minimal default value
	mess = fmt.Sprintf("flr: (CNT==%v;user=%v)%v -- %v; what:%v;\n",
		rr.count, rr.agent, rr.reqDescr, rr.start, rr.what)
	if byteSet(fl.mode, 2) {
		additionalMess = fmt.Sprintf("%v%v\n", additionalMess, rr.connState)
	}
	if byteSet(fl.mode, 3) {
		additionalMess = fmt.Sprintf("%v%v\n", additionalMess, rr.contentType)
	}
	mess = mess + additionalMess
	return
}

//210316 14:52 How would not forget that there is this function!
func getRequestDescr(r *http.Request) string {
	return fmt.Sprintf("(%v-%v-%v)", r.Method, r.RemoteAddr, r.RequestURI)
}
