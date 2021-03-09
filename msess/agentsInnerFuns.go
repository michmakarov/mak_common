// agentsInnerFuns
//210306 10:53 The place for functions using immediately by function agentMonitor for modification agests slice.
//That is the functions must not be called from other places.
package msess

import (
	"fmt"
)

func qqqmain() {
	fmt.Println("Hello World!")
}

func register(data interface{}) (res MonitorResult) {
	var a *Agent
	var ok bool
	if a, ok = data.(*Agent); !ok {
		res.Err = fmt.Errorf("register: Given data for registration is not converted to *Agent")
		return
	}
	for _, item := range agents {
		if item.Tag == a.Tag {
			res.Err = fmt.Errorf("register: Agent %v already registered", a.String())
			return
		}
	}
	agents = append(agents, a)
	return
}

func unregister(data interface{}) (res MonitorResult) {
	var a *Agent
	var ok bool
	var ind int
	var item *Agent
	if a, ok = data.(*Agent); !ok {
		res.Err = fmt.Errorf("unregister: Given data for registration is not converted to *Agent")
		return
	}
	for ind, item = range agents {
		if item.Tag == a.Tag {
			res.Err = fmt.Errorf("unregister: Agent %v already registered", a.String())
			break
		}
	}
	if ind >= len(agents)-1 {
		res.Err = fmt.Errorf("unregister: Agent %v is not registered", a.String())
		return
	}
	agents = append(agents[:ind], a[ind+1:])
	return
}
