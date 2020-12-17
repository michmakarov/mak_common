// sysErrLogging
package kerr

import (
	"errors"
	"fmt"
	"log"
	"os"
	"time"
)

var sysErrLog *log.Logger
var sysErrLogFileName string
var sysErrLogFileNamePrefix string = "KotSysErr"

func init() {
	if sysErrLog != nil {
		return
	}
	//var logFileName string
	var err error
	sysErrLogFileName = "KotSysErr" + time.Now().Format("20060102_150405") + ".log"
	if err = createErrSysLog(sysErrLogFileName); err != nil {
		//fmt.Printf("Ouch! sysErrLog not creared err=%v", err.Error())
		//190721
		panic(fmt.Sprintf("kerr.init: sysErrLog not creared err=%v", err.Error()))
	}
}

func createErrSysLog(logFileName string) error {
	var logFile *os.File
	var err error
	if sysErrLog != nil {
		return errors.New("SysErrLogging: The system error logger already exist.")
	}
	if logFile, err = os.Create(logFileName); err != nil {
		return errors.New("SysErrLogging: Cannot create a log file " + err.Error())
	}
	sysErrLog = log.New(logFile, "", log.LstdFlags)
	return nil
}

func SysErrPrintf(fMess string, par ...interface{}) {
	if sysErrLog == nil {
		fmt.Println("Ouch! sysErrLog not creared. You have wanted to say ", fMess)
		return
	}
	sysErrLog.Printf(fMess, par...)
}

func SysErrPrintln(par ...interface{}) {
	if sysErrLog == nil {

		//fmt.Println("Ouch! sysErrLog not creared. You have wanted to say ", par)
		ProcessingError(-1001, fmt.Sprintf("Ouch! sysErrLog not creared. You have wanted to say ", par))
		//return
		os.Exit(1001)
	}
	sysErrLog.Println(par)
}

func GetSysErrLogFileName() string {
	if sysErrLog == nil {
		panic("kerr.GetSysErrLogFileName: systen errors log is not created yet.")
	}
	return sysErrLogFileName
}
