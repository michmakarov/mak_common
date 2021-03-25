// API
package msess

import (
	"fmt"
)

func QQQ() {
	fmt.Println("Hello World! from msess API.")
}

var ServerStopped chan struct{} = make(chan struct{})

func MsessRuns() bool {
	if server != nil {
		return true
	}
	return false
}
