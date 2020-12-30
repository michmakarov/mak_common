//190926 About versionung
//The question: may a user manual and a version information be into the same box?
//Now I answer: no. The user manual is distinct project that may reference to a version but must not give its description.
//So, since now I am abandoning to support files with description of a version (.dot) and returning to the old idea of having a special code file (version.go) for this purpose
//191002 I confirm it
//201225 04:41 Let's to agree it for the present time. In it are sober thoufhts:
//a commit only fixes some state and says about the past
//But what is now? In what is our will?
//Yesterday a question arose: what is a library (or a packet) version. As the answer see mak_common/sv.sh
package ksess

import (
	"fmt"
	"strconv"
)

//191002 That is the first question that emerged from weak ability to foresee - how to get this?
//201224 15:14 the tgh.sh does it.
var commit_data_1 = "---201216_rels:da21c61--*main--201230_0720---"

//VersionDescr binds a version number with the version description
type VersionDescr struct {
	Number             string //Version number is a string of format "190926", that is "<Year><Month><Day>"
	ProgName           string //   = "makcommon.ksess"
	Commit_branch_time string // The time is last time of launch the tgh.sh
	VersionState       string //= "developing" How to do that it be filled automaticlly?
	Text               string
}

//versinList defines VersionDescr for each version that has been occured
//Current version has index 0, previous one - index 1, and so on.
var versionList = []VersionDescr{
	{"201223", "makcommon.ksess", commit_data_1, "developing", blabla_201224},
	{"191223", "KSESS", commit_data_1, "closed 191224 12_18", blabla_191223},
	{"191002", "KSESS", commit_data_1, "developing", blabla_191002},
}

func GetCurrVerInfo() string {
	return GetVerInfo(versionList[0].Number)
}

//Sequence such constant define textual description of a version.

const blabla_201224 = `
Plan:<br>
To get goods that is enough for the rels project
Результаты <br>
Предложение к следующей версии<br>
`

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
			return it.ProgName + "_" + it.Number + " : " + it.VersionState + ";" + it.Commit_branch_time
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
