//190926 About versionung
//The question: may a user manual and a version information be into the same box?
//Now I answer: no. The user manual is distinct project that may reference to a version but must not give its description.
//So, since now I am abandoning to support files with description of a version (.dot) and returning to the old idea of having a special code file (version.go) for this purpose
//191002 About this versioning
//This file is a copy of the such from the KSODD. But KSODD is a Git's repository immediatelly, whereas this package is a directory into KOT_COMMON repesitory.
//It is significant that a package has not its own commit but of an including library
//So here there is need of function that obtains a library commit
package kpglo

import (
	"fmt"
	"net/http"
	"strconv"
)

//var commit_data_1 = "No_data_yet"
var commit_data_1 = "No_data_still"

//VersionDescr binds a version number with the version description
type VersionDescr struct {
	Number       string //Version number is a string of format "190926", that is "<Year><Month><Day>"
	ProgName     string //   = "ksodd"
	GitBranch    string // "pgf_with_ksess"
	Git_commit   string
	IlnVer       string //This is similar to Number but shows a date of pulling codes those generated not dependly by some developers which do not see needs of versioning
	VersionState string //= "developing"
	Text         string
}

//versinList defines VersionDescr for each version that has been occured
//Current version has index 0, previous one - index 1, and so on.
var versionList = []VersionDescr{
	{"191102", "kpglo", "master", commit_data_1, "", "closed", blabla_191102},
	{"191107", "kpglo", "master", commit_data_1, "", "developing", blabla_191107},
}

func GetVesionInfo() string {
	return getVerInfo(versionList[0].Number)
}

func GetVerNum() string { return versionList[0].Number }

//Constants with names of patten blabla_<version number> define textual description of a version.

const blabla_191102 = `
This is first new old version of kpglo.<br>
As now (191102) it has been said above<br>
//191002 About this versioning
//This file is a copy of the such from the KSODD. But KSODD is a Git's repository immediatelly, whereas this package is a directory into KOT_COMMON repesitory.
//It is significant that a package has not its own commit but of an including library
//So here there is need of function that obtains a library commit
План:<br>
The main idea is to develop writing and reading asynchronously with informing about progress of these processes.
Beside this, the notion of closing a large object must be developed and realized.
And all those must be described into the Programmer manual with describing all existing functionality.
Результаты<br>
191103 asynch.go was established. Yesterday I asked for more clear tasks. <br>
Now (191105) I have decided to do something asynchronous and close the version.<br>
191107 Yesterday I had ended func AsynchReadLo(loid int, chunkSize int) (chunkChan chan ReadChunkRep)<br>
and tested it by /madm/save_lo from server ksodd;Ver=ksodd(pgf_2)191101 : developing; commit_date=e864f86f_07.11.2019<br>
At this I decide to close this version.<br>

Предложение к следующей версии:<br>
`

const blabla_191107 = `
<br>
<br>

Предложение к следующей версии:<br>
`

func getVerInfo(num string) string {
	for _, it := range versionList {
		if it.Number == num {
			it.Git_commit = commit_data_1
			return it.ProgName + "(" + it.GitBranch + ")" + it.Number + " : " + it.VersionState + "; commit_date=" + it.Git_commit
		}
	}
	return fmt.Sprintf("getVerInfo:No such version - %v", num)
}

func getVersionText(num string) string {
	//var s string
	for _, it := range versionList {

		if it.Number == num {
			return it.Text
		}
	}
	return fmt.Sprintf("getVersionText:No such version - %v", num)
}

func getVerList() string {
	var s string
	for _, it := range versionList {
		s = s + it.Number + "<br>"
	}
	return s
}

func verHandler(w http.ResponseWriter, r *http.Request) {
	var (
		ver = r.URL.Query().Get("ver")
	)
	defer func() {
		if rec := recover(); rec != nil {
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.WriteHeader(500)
			w.Write([]byte(fmt.Sprintln("verHandler: panic occured with message = ", rec)))
		}
	}()
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	switch {
	case ver == "list":
		w.Write([]byte(getVerList()))
	case ver == "last":
		w.Write([]byte(GetVesionInfo()))
	case isNum(ver):
		w.Write([]byte(getVersionText(ver)))
	default:
		w.Write([]byte("usage: /version?ver=INTEGER || /version?ver=last || /version?ver=list"))
	}

}

func isNum(ver string) bool {
	var err error
	if _, err = strconv.Atoi(ver); err != nil {
		return false
	}
	return true
}
