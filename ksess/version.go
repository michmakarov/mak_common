//190926 About versionung
//The question: may a user manual and a version information be into the same box?
//Now I answer: no. The user manual is distinct project that may reference to a version but must not give its description.
//So, since now I am abandoning to support files with description of a version (.dot) and returning to the old idea of having a special code file (version.go) for this purpose
//191002 I confirm it
package ksess

import (
	"fmt"
	"strconv"
)

//191002 That is the first question that emerged from weak ability to foresee - how to get this?
var commit_data_1 = "No_data"

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
	{"191223", "KSESS", "No git", commit_data_1, "?????", "closed 191224 12_18", blabla_191223},
	{"191002", "KSESS", "No git", commit_data_1, "190926", "developing", blabla_191002},
}

// For backward compatibility and getting current version
func GetVesionInfo() string {
	return GetVerInfo(versionList[0].Number)
}

//Sequence such constant define textual description of a version.

const blabla_191223 = `
I alresdy do not know what was doing as well was done in previous version.
So the practice of permanent doing is vary bad!
Plan:<br>
To make an option "hurry forbidden". See KSODD 191223.
Результаты <br>
Method func (gnd *GlobalNotDone) URL_InDoing(userId string, URL string) string
SessConfigParams.HurryForbidden bool
if sessCP.HurryForbidden { ... into feeler.go


Предложение к следующей версии<br>
`

const blabla_191002 = `
1.ksess.IsAdmin<br>
<br>
Результаты <br>
Предложение к следующей версии<br>
`

func GetVerInfo(num string) string {
	for _, it := range versionList {
		if it.Number == num {
			it.Git_commit = commit_data_1
			return it.ProgName + "_" + it.GitBranch + "_" + it.Number + " : " + it.VersionState + "; commit_date=" + it.Git_commit
		}
	}
	return fmt.Sprintf("getVerInfo:No such version - %v", num)
}

func GetVersionText(num string) string {
	//var s string
	for _, it := range versionList {

		if it.Number == num {
			return it.Text
		}
	}
	return fmt.Sprintf("getVersionText:No such version - %v", num)
}

func GetVerList() string {
	var s string
	for _, it := range versionList {
		s = s + it.Number + "<br>"
	}
	return s
}

func isNum(ver string) bool {
	var err error
	if _, err = strconv.Atoi(ver); err != nil {
		return false
	}
	return true
}
