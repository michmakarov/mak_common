// httpserver_log
package msess

import (
	"fmt"
	"log"
	"os"
	"time"
)

var httpServerLog *log.Logger

func init() {
	var err error
	var out *os.File
	var httpServerLogFileName = "Servlog" + time.Now().Format("20060102_150405") + ".log"
	if out, err = os.Create("logs/" + httpServerLogFileName); err != nil {
		panic(fmt.Sprintf("init (httpserver_log): os.Create err=%v", err.Error()))
	}

	httpServerLog = log.New(out, "", log.LstdFlags)
}
