// httpserver_log
//210602 08:18 For what it is needed since there is the front log? See type feeler struct.
//_______13:20 The front log does not has any matter here. It is error log.
package msess

import (
	"fmt"
	"log"
	"os"
	"time"
)

var httpServerLog *log.Logger

func createHttpserverLog() { //210603 09:49 I think an init() here does not good
	var err error
	var out *os.File
	var httpServerLogFileName = "Servlog" + time.Now().Format("20060102_150405") + ".log"
	if stringSet(sessCP.Loggers, "h") != true {
		return
	}
	if out, err = os.Create("logs/h" + httpServerLogFileName); err != nil {
		panic(fmt.Sprintf("init (httpserver_log): os.Create err=%v", err.Error()))
	}

	httpServerLog = log.New(out, "", log.LstdFlags)
}
