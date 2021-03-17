//200311 16:53 It is a new attempt to wrought a common approach to versioning of my golang projects
//As it seems to me maintaining history files (or diarys) is well.
//But it will be better to combine versioning ang diarying.
package msess

//"fmt"
//"strconv"

var git_data = "yet no git data"

type VersionDescr struct {
	Number             string //Version number is a string of format "190926", that is "<Year><Month><Day>"
	ProgName           string //
	Commit_branch_time string // The time is last time of launch the tgh.sh
	VersionState       string //= "developing" How to do that it be filled automaticlly?
	Text               string
}

//versinList defines VersionDescr for each version that has been occured
//Current version has index 0, previous one - index 1, and so on.
var versionList = []VersionDescr{
	{"210311", "msess", git_data, "developing", blabla_210311},
}

func GetCurrVerInfo() string {
	return versionList[0].Text
}

func GetCurrVer() string {
	return versionList[0].Number
}

//Sequence next constants define textual descriptions of versions.

const blabla_210311 = `
Plan:<br>
<br>
<br>
Developer_notes:<br>
This is the first version of msess; see history.txt, record 210311 16:30
<br>210311 17:36 Current problem.
The index request is a intercepted one, so the answer to it must be wholy controlled by this packet.
In other hand, an application programmer should have ability to define his own html and js.
What to do? To send only js!!??
<br> 210312 14:17
As all last pondering show the idea of AGENT have sense only for single page application.
Is it so?
<br> 210312 20:04 The question: how dynamically load additional script into the current page? 
As it is shown by /home/mich412/Progects/http_srv_210312 there is an acceptable way.
<br> 210316 12:35 A urgent want have arised to rid of the web socket for the simple periodic polling.
But after https://learn.javascript.ru/long-polling and https://developer.mozilla.org/en-US/docs/Web/API/EventSource
the want was diminished.
Результаты:<br>
Предложение к следующей версии:<br>
`
