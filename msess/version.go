//210311 16:53 It is a new attempt to wrought a common approach to versioning of my golang projects
//As it seems to me maintaining history files (or diarys) is well.
//But it will be better to combine versioning ang diarying.
//210422 06:32 Under
package msess

import (
	//"fmt"
	"mak_common/mversion"
	//"strconv"
)

//versinList defines VersionDescr for each version that has been occured
//Current version has index 0, previous one - index 1, and so on.
var versionList = mversion.VersionList{
	{"mak_common.msess", "---210224_rels:7d9714d--*main--210723_1012---", blabla_210609},
	{"mak_common.msess", "---nv no env---", blabla_210604},
	{"mak_common.msess", "---nv no env---", blabla_210311},
}

func GetCurrVerInfo() string {
	return versionList.GetCurrVerInfo()
}

func GetCurrVer() string {
	return versionList[0].Number
}

//Sequence next constants define textual descriptions of versions.

const blabla_210609 = `
Plan:
1. to test logging<br>
2. ind problem<br>
<br>
<br>---
Developer_notes:<br>
210609 15:07 embed package!<br>
210610 03:49 Yeasterday an idea hovered over: to have resources.<br>
Yes, to have learnd the structure of ELF files is very, very good!<br>
But to have the agent files in work directory of the server is no less good!<br>

<br>
<br>---
Результаты:<br>---
Предложение к следующей версии: <br>
<br>---
`

const blabla_210604 = `
Plan:
1.Now a request of "/" cause "feeler panic err = runtime error: invalid memory address or nil pointer dereference". Why?
2.(210607 04:10) To convert the "rules&terms.html" for going to distinct definition files
<br>---
<br>
<br>
Developer_notes:<br>
210604 18:03 About "Plan:1". If you have some awkwardness in implementation of your conception it further obligatory will emerge as rough error<br>
210607 03:22 About "Plan:1". Logic of func (f *feeler) ServeHTTP(w http.ResponseWriter, r *http.Request) must be overwrought and described.<br>
<br>---
Результаты:<br>---
1. There was so many foolishness and I fear that remains no less ...
2. It was done in principle
Предложение к следующей версии: <br>
"/" wears an agent but where to search files that enbodies an agent? Let it be "ind problem"
<br>---
`

const blabla_210311 = `
Plan:<br>---
<br>
<br>
Developer_notes:<br>
This is the first version of msess; see history.txt, record 210311 16:30
<br>210311 17:36 Current problem.
The index request is a intercepted one, so the response to it must be wholy controlled by this package.
In other hand, an application programmer should have ability to define his own HTML and JS.
What to do? To send only js!!??
<br> 210312 14:17
As all last pondering show the idea of AGENT have sense only for single page application.
Is it so?
<br> 210312 20:04 The question: how dynamically load additional script into the current page? 
As it is shown by /home/mich412/Progects/http_srv_210312 there is an acceptable way.
<br> 210316 12:35 A urgent want have arised to rid of the web socket for the simple periodic polling.
But after https://learn.javascript.ru/long-polling and https://developer.mozilla.org/en-US/docs/Web/API/EventSource
the want was diminished.<br>
210322 16:23 Golang http.Server
1. Is there the way to find out the response code that Server.Handler returns? As if there is not.
2. Server.WriteTimeout = 0; It seems well as there is the info about not done requests.
3. Server.ReadTimeout = 0; It does not seems well at all as maybe errors in an agent realization, as well wrong work of the net.
4. If waiting of a complete request is aborted by Server.ReadTimeout may in what way the info about it be obtained?
May be Server.ErrorLog gives it.
<br>210324 11:24
The truth that I have understanded very recently: mutexes gard code but not data.
So if you desire to gard data you must guarantee that only one function has access to it (for changing!).
And so the not_done_global_storage must will be remade.
<br>!!! 210326 16:44
I decide that the agent cookie should contain only a tag.
It simplifies all not only but more corresponds the matter!<br>!!!
210331 06:31
I have delayed the msess and the 210224_rels because I now in some concept quagmire with the msess.
Or, maybe, the openVPN does not give calm and peace but itch. So I will plunge thoroughly into the openVPN  <br>
210419 I am here again. <br>
<br>---
Результаты:<br>---
210604 03:54 Resalts! About what you are? Now the affairs are too mean, broad, and complicated to talk about such high matter!<br>
Предложение к следующей версии:<br>---
To tune program for accomplishing elementary right behaviour
`
