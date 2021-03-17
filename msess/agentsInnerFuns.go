// agentsInnerFuns
//210306 10:53 The place for functions using immediately by function agentMonitor for modification agests slice.
//That is the functions must not be called from other places.
package msess

import (
	"fmt"
)

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
	var newAgents Agents
	var ok bool
	var ind int
	var item *Agent
	if a, ok = data.(*Agent); !ok {
		res.Err = fmt.Errorf("unregister: Given data for registration is not converted to *Agent")
		return
	}
	for ind, item = range agents {
		if item.Tag == a.Tag {
			break
		} else {
			newAgents = append(newAgents, agents[ind])
		}
	}
	if ind >= len(agents)-1 {
		res.Err = fmt.Errorf("unregister: Agent %v is not registered", a.String())
		return
	}
	agents = newAgents
	return
}

func is_registered(data interface{}) (res MonitorResult) {
	var cd *SessCookieData
	var ok bool
	var a Agent
	if cd, ok = data.(*SessCookieData); !ok {
		panic("is_registered: Given data is not converted to *SessCookieData")
	}
	for _, item := range agents {
		if item.Tag == cd.Tag {
			a = Agent{}
			a.RegTime = item.RegTime
			a.RemoteAddress = item.RemoteAddress
			a.UserAgent = item.UserAgent
			a.Tag = item.Tag
			a.UserId = item.UserId
			a.conn = item.conn
			res.Data = &a
			return
		}
	}
	res.Err = fmt.Errorf("Agent with a tag of %v is not registered", cd.Tag)
	return
}
