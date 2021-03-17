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

//210101 The requestRecord is incoming request descriptor.
//Why does it need if there is r *http.Request?
//It bear info which there are not in http.Request
type requestRecord struct {
	count       int64
	start       string
	reqDescr    string
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
		go f.flgr.Run()
	}

	SendToGenLog("Feeler", " started")
	return f, nil
}

func (f *feeler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var (
		err           error
		cD            *SessCookieData
		clientErrMess string
		clientErrCode int

		agent             *Agent
		cancel            context.CancelFunc
		ctx               context.Context
		requestCouter     int64 //!!! 190820_2 The problem of requests counter
		OUTSESSION_REQEST bool  //201223 05:45
	)
	defer func() {
		var rec interface{}
		if rec = recover(); rec != nil {
			kerr.SysErrPrintf("feeler ServeHTTP coughts panic = %v", rec)
			if sessCP.Debug != 0 {
				debug.PrintStack()
			}
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.WriteHeader(500)
			w.Write([]byte(fmt.Sprintf("feeler panic err = %v", rec)))
		}
	}()

	var WriteToLog = func(user, do string) { //210309 16:46
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

	requestCouter = atomic.AddInt64(&f.feelerCount, 1)

	//r.Method = strings.ToUpper(r.Method)

	//1. Outsession requests pass without any hinder
	if checkURLPath(r.URL.Path) {
		goto gettingResponse
	}

	//2. to check agent coockie and getting the current agent
	if cD, err = getCookieData(r); err != nil {
		if agent, err = agentRegistered(cD, r); err != nil { //no agent (or it was forgeded)
			if r.URL.Path == "/" { //the http client will be given a new agent
				indexHandler(w, r)
				return
			} else { //no agent: all requests excluding "/" are forbidden
				clientErrMess = fmt.Sprintf("getCookieData: no agent err=%v\n", err.Error())
				clientErrCode = 403
				goto exitOnErr
			}
		}

	}

gettingResponse:
	r = r.WithContext(context.WithValue(r.Context(), NumberCtxKey, strconv.FormatInt(requestCouter, 10)))
	r = r.WithContext(context.WithValue(r.Context(), UserIdCtxKey, cookData.UserIDAsString()))
	r = r.WithContext(context.WithValue(r.Context(), URLCtxKey, r.RequestURI)) //190408
	ctx, cancel = context.WithCancel(r.Context())
	r = r.WithContext(ctx)
	calcHTTPResponse(c, w, r, cancel)
	return

exitOnErr:
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(clientErrCode)
	w.Write([]byte(clientErrMess))

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

	FlrLog = &feelerLogger{}
	if f, err = os.Create("logs/" + fileName); err != nil {
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

//210101 //210104 14:12
func (fl *feelerLogger) getFlrlogMess(rr requestRecord) (mess string) {
	var additionalMess string

	//210101 here mess obtains its minimal default value
	mess = fmt.Sprintf("flr: (CNT==%v;user_id=%v)%v -- %v; what:%v;\n",
		rr.count, rr.user_id, rr.label, rr.start, rr.what)
	if byteSet(fl.mode, 2) {
		additionalMess = fmt.Sprintf("%v%v\n", additionalMess, rr.connState)
	}
	if byteSet(fl.mode, 3) {
		additionalMess = fmt.Sprintf("%v%v\n", additionalMess, rr.contentType)
	}
	mess = mess + additionalMess
	return
}

func DoIndexRequest(w http.ResponseWriter, r *http.Request) {
	var cookieData *sessCookieData

	if r.URL.Path != "/" {
		return
	}
	if cookieData, err = getSession(r); err != nil {
	}
}

//210316 14:52 How would not forget that there is this function!
func getRequestDescr(r *http.Request) string {
	return fmt.Sprintf("(%v-%v-%v)", r.Method, r.RemoteAddr, r.RequestURI)
}
