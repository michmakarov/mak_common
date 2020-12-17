// utils
package ksess

import (
	"fmt"
	"os"
	"strings"
)

//201203 16:29
//A logsDir paramrter must
func checkLogsDir(logsDir string) (err error) {
	if logsDir == "" {
		return
	}
	switch string(os.PathSeparator) {
	case "/":
		if !strings.HasSuffix(logsDir, "/") {
			err = fmt.Errorf("Checking a log dir error: the dir not ended by /")
			return
		}
	case "\\":
		if !strings.HasSuffix(logsDir, "\\") {
			err = fmt.Errorf("Checking a log dir error: the dir not ended by \\")
			return
		}
	default:
		err = fmt.Errorf("Checking a log dir error: Illegal value of the path separator= %v", os.PathSeparator)
		return
	}
	return
}
