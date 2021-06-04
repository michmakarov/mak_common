// gLogger ( stands for general logger) is intended for logging reports of events that do not match event of occurring a system error or coming a request
// 210316 14:29 It have been pulled from the ksess
package msess

import (
	"fmt"
	"log"
	"mak_common/kerr"
	"os"
	"time"
)

var gLog *generalLogger
var generalLogFileName string

//generalLogger is intended for logging events of not special character that are not such as  arising a system errors or coming a request
type generalLogger struct {
	log      *log.Logger
	sendChan chan string
}

//GetGeneralLogFileName returns  a current general log file nsme
//190715_2
func GetGeneralLogFileName() string {
	if gLog == nil {
		panic("ksess.GetGeneralLogFileName:General logger is not created yet.")
	}
	return generalLogFileName
}

func createGeneralLog() (err error) {
	var f *os.File

	//kerr.PrintDebugMsg(false, "restoreSess", fmt.Sprintf(" createGeneralLog: fileName=%v", fileName))

	if stringSet(sessCP.Loggers, "g") != true {
		return
	}

	generalLogFileName = "GLog" + time.Now().Format("20060102_150405") + ".log"

	gLog = &generalLogger{}
	if f, err = os.Create("logs/g" + generalLogFileName); err != nil {
		gLog = nil
		return
	} else {
		gLog.log = log.New(f, "", log.LstdFlags)
	}
	gLog.sendChan = make(chan string, 253)
	return
}

func (gLog *generalLogger) run() {
	if gLog == nil {
		kerr.SysErrPrintf("An attempt to run nil general log")
		return
	}
	go func() {
		for {
			mess := <-gLog.sendChan
			gLog.log.Println(mess)
		}
	}()
}

func SendToGenLog(tp string, mess string) {
	if gLog == nil {
		return
	}
	fulMess := fmt.Sprintf("%v:%v", tp, mess)
	gLog.sendChan <- fulMess
}
