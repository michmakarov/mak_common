// agentsInnerFuns
//210306 10:53 The place for functions using immediately by function agentMonitor for modification agests slice.
//That is the functions must not be called from other places.
package msess

import (
	"fmt"
	"strings"
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

//210323 05:52 The question: if the cd is not beared a valid tag
//Then it retuns an empty agent! 14:25 -- nil! As it was initially!
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

//210325 05:28
func where_user(data interface{}) (res MonitorResult) {
	var userId string
	var ok bool
	var a Agent
	if userId, ok = data.(string); !ok {
		panic("user is_registered: Given data is not converted to string (tag)")
	}
	for _, item := range agents {
		if item.UserId == userId {
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
	return
}

//210326 06:11
func assign_user(data interface{}) (res MonitorResult) {
	//var userId string
	var ok bool
	var a *Agent
	var item *Agent
	if a, ok = data.(*Agent); !ok {
		panic("assign_user: Given data is not converted to *Agent")
	}
	for _, item = range agents {
		if item == a {
			panic("assign_user: item==a")
		}
		if item.Tag == a.Tag {
			if item.UserId != "" {
				panic("assign_user: item.UserId !=\"\"")
			}
			item.UserId = a.UserId
			return
		}
	}
	if item == nil {
		panic("assign_user: no such agent; Tag="+a.Tag)
	}
	return
}

//FindAgentByTag returns nil if the searching is not successful.
func findAgentByTag(tag string) (agent *Agent) {
	for ind, item := range agents {
		if item.Tag == tag {
			return agents[ind]
		}
	}
	return
}

func findAgentByUser(user_id string) (agent *Agent) {
	for ind, item := range agents {
		if item.UserId == user_id {
			return agents[ind]
		}
	}
	return
}

//210319 16:53
// Always res.Data==nil
// data : map of WsMess
//210321 15:52 Is or is not there any sense in makeCopyAndCheck function?
//In other words, is the data converted to a pointer or to a copy of the value?
//To a pointer! See Questions interface{}. So the sense is and is very!
func sendToWs(data interface{}) (res MonitorResult) {
	var mess, messCopy WsMess
	var ok bool
	var a *Agent
	var addr string
	if mess, ok = data.(WsMess); !ok {
		panic("sendToWs: Given data is not converted to WsMess")
	}
	if messCopy, err = makeCopyAndCheck(mess); err != nil {
		res.Err = fmt.Errorf("makeCopyAndCheck err=%v", err.Error())
		return
	}
	addr = messCopy["to"]
	switch strings.Split(addr, ":")[0] {
	case "tag":
		a = findAgentByTag(strings.Split(addr, ":")[1])
	case "user":
		a = findAgentByUser(strings.Split(addr, ":")[1])
	}
	if a == nil {
		res.Err = fmt.Errorf("=%v", err.Error())
		return
	}
	a.WsOut <- messCopy
	return
}

func doInWsMess(wsInMess WsMess) {
	panic("doInWsMess is not realized yet.")
}
