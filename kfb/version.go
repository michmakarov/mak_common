package kfb

import (
	"fmt"
	//"net/http"
	"strconv"
)

var commit_data_1 = "No_data"

//VersionDescr binds a version number with the version description
type VersionDescr struct {
	Number       string //Version number is a string of format "190926", that is "<Year><Month><Day>"
	ProgName     string //   = "ksodd_client"
	GitBranch    string //
	Git_commit   string
	IlnVer       string //This is similar to Number but shows a date of pulling codes those generated not dependly by some developers which do not see needs of versioning
	VersionState string //= "developing"
	Text         string
}

//versinList defines VersionDescr for each version that has been occured
//Current version has index 0, previous one - index 1, and so on.
var versionList = []VersionDescr{
	{"191013", "kot_common/kfb", "\"Git is not yet\"", commit_data_1, "No", "developing", blabla_191013},
	{"191008", "kot_common/kfb", "\"Git is not yet\"", commit_data_1, "No", "developing", blabla_191008},
}

//GetVesionInfo returns representation of the current version with index=0
//func GetVesionInfo() string {
//	versionList[0].Git_commit = commit_data_1
//	return versionList[0].ProgName + "_" + versionList[0].GitBranch + "_" + versionList[0].Number + " : " + versionList[0].VersionState + "; commit_date=" + versionList[0].Git_commit
//}
func GetVesionInfo() string {
	return GetVerInfo(versionList[0].Number)
}

//Sequence such constant define textual description of a version.
const blabla_191013 = `
<br>

Предложение к следующей версии<br>
`
const blabla_191008 = `
From Readme to 191008_FB_CLIENT: <br>

Now (191008) Ilnur have said that THE LIFE forces him to use sql.DB so, as he uses it.<br>
The conversation has been in context of presantation of 191004_PG_CLIENT<br>
- that is all the theory but the practice is something else ...<br>

So I am going to create a like client for firebird and to offer him to give a function which will crash this client<br>
Also in doing this I am going to revise kfb package with further purpose to do same with kpglo package<br>
<br>
So it is revising.<br>
1.To realize new old approach to versioning <br>
That is all. It is enough.<br>

Предложение к следующей версии<br>
There is need to revise the goal and destination of this package<br>
As I remember the main goal of the package was stated to track queries in execution state<br>
But what to say about situation when a query is done but the commit (ot rollback) is not<br>
And else: what is about chain of queries?<br>
`

func GetLastVersion() string {
	return versionList[len(versionList)-1].Number
}
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

func isNum(ver string) bool {
	var err error
	if _, err = strconv.Atoi(ver); err != nil {
		return false
	}
	return true
}
