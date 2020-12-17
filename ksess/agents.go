// commandHandlers
package ksess

import (
	"fmt"
	"io"

	//"io/ioutil"
	"html/template"
	"mak_common/kutils"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const maxAgents = 100

type Agent struct {
	RegTime       time.Time
	RemoteAddress string //r.RemoteAddr
	UserAgent     string //r.UserAgent()
	UserId        int    //cookData.UserID see getSession
}

type Agents map[string]*Agent

var agents Agents = make(map[string]*Agent)
var agentsMutex sync.Mutex

func (agents Agents) Register(r *http.Request) (sugnature string, err error) {
	var sessionData SessionData
	var signature string
	var a Agent

	if len(agents) >= maxAgents {
		err = fmt.Errorf("Agents.Register: too many agents; allowed not more %v", maxAgents)
		return
	}
	sessionData = GetSession(r)
	signature, _ = kutils.TrueRandInt()
	a.RegTime = time.Now()
	a.RemoteAddress = r.RemoteAddr
	a.UserAgent = r.UserAgent()
	a.UserId = sessionData.UserID
	agentsMutex.Lock()
	agents[signature] = &a
	agentsMutex.Unlock()
	return signature, nil
}

func (a Agents) Registered(signature string) (yes bool) {
	agentsMutex.Lock()
	if agents[signature] != nil {
		yes = true
	}
	agentsMutex.Unlock()
	return
}

func (a Agent) String() (res string) {
	res = fmt.Sprintf("%v; RA=%v; UserId=%v; %v", a.RegTime.Format(startFormat), a.RemoteAddress, a.UserId, a.UserAgent)
	return
}

func (a Agents) String(lb string) (res string) {
	agentsMutex.Lock()
	for key, value := range a {
		res = res + key + "==" + value.String() + lb
	}
	agentsMutex.Unlock()
	return
}

func GetAgentsList(lb string) string {
	return agents.String(lb)
}

func (a Agents) addInfo(agentSignature string, b io.ReadCloser) (err error) {
	if !a.Registered(agentSignature) {
		err = fmt.Errorf("Agent %v not registered", agentSignature)
		return
	}

	if sessCP.NotAgentDebugging {
		return
	}

	var f *os.File
	var fFullName string
	fFullName = sessCP.AgentFileDir + string(rune(os.PathSeparator)) + agentSignature
	if f, err = os.OpenFile(fFullName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err != nil {
		return
	}
	f.WriteString(time.Now().Format(startFormat) + " : ")
	io.Copy(f, b)
	f.WriteString("\n")
	return
}

func postAgentErrorReportHandler(w http.ResponseWriter, r *http.Request) {
	var err error
	var agentSignature string
	if sessCP.NotAgentDebugging {
		w.WriteHeader(403)
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("/debug: debugging kot agent forbidden - sessCP.NotAgentDebugging==true"))
		return
	}
	if agentSignature = strings.TrimSpace(r.Form.Get("AGENT_SIGNATURE")); agentSignature == "" {
		if agentSignature, err = agents.Register(r); err != nil {
			w.WriteHeader(403)
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.Write([]byte("/debug: error: " + err.Error()))
			return
		}
		w.WriteHeader(200)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Write([]byte(agentSignature))
		return
	}

	//Now we know that there is an agent signature and are going to add the info from request body to the file of this agent
	if err = agents.addInfo(agentSignature, r.Body); err != nil {
		w.WriteHeader(403)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Write([]byte("/debug: error of adding info : " + err.Error()))
		return
	}
	w.WriteHeader(200)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write([]byte("Debugging info was successfully added to " + agentSignature))
	return
}

//It creates or empties agents directory
func setEmptyAgents() (err error) {
	var files []string
	var agentsDir string
	var agentsDirStat os.FileInfo
	if agentsDir, err = filepath.Abs(sessCP.AgentFileDir); err != nil {
		return
	}

	//It checks the file if such is
	//If it is not the function creats a directory
	//If it is but no directiory the function returns erorr
	if agentsDirStat, err = os.Stat(agentsDir); err != nil {
		if os.IsNotExist(err) {
			if err = os.Mkdir(agentsDir, 0777); err != nil {
				return
			}
		} else {
			return
		}
	} else {
		if !agentsDirStat.IsDir() {
			err = fmt.Errorf("File %v is exist but it is not directory", agentsDir)
		}
	}

	//Now we are sure that the agents directory exists in a right place and we clean it not regard that it may be empty
	if files, err = filepath.Glob(filepath.Join(agentsDir, "*")); err != nil {
		return
	}
	for _, file := range files {
		if err = os.RemoveAll(file); err != nil {
			return
		}
	}
	return
} //setEmptyAgents

func getAgentWorkerHandler(w http.ResponseWriter, r *http.Request) {
	type pageData struct {
		ControlPassword            string
		KotURLForGetAgentSignature string
	}

	var data pageData
	var page *template.Template
	data.ControlPassword = sessCP.ControlPassword
	data.KotURLForGetAgentSignature = fmt.Sprintf("/post_agent_error_report?CONTROL_PASSWORD=%s", sessCP.ControlPassword)

	if page, err = template.ParseFiles(sessCP.AgentWorkerDir + "agentworker.js"); err != nil {

	}

	if err = page.Execute(w, data); err != nil {
		//		kerr.SysErrPrintf(" indexHandler: err == %s", err.Error())
	}

}
