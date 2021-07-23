//210420 16:47 It is a template for organizing versioning an end application (a package main) as well packages that are used by that application
//It defines only the type VersionDescr and VersionList, all other are only templates and recomendation
//It assumes that an instance of VersionList will be formed manually by a application programmer;
//and that first member of it (with index 0) is describing the current version on developing stage.
//++++++++++++++++++++++++++++++++++++++++++++++
//210422 05:43 It assumes also that the Number and Git_data fields of the current vesion will be filled by a script
// with a a sed command. So strings of Number and Git_data fields of the current version (and only them)
//must have the special format. Namely:
//Number : "nv<data>env"
//Git_data : "gd<data>egd"
//So, After manually inserting the new current version Number = "nv no env" and Git_data = "gd no egd"
//All below descriptions must (!) be deprived munually those tags.
//_______12:28 What is a version of a component of the mak_common library? See /home/mich412/go/src/mak_common/sv.sh
//Now I have take that interpritation and so:
//1. The type VersionDescr has only three fields: ProgName, Number, and Descr.
//2. The Number is set to its actual value by bunch if scripts as /home/mich412/go/src/mak_common/b-r.sh and
// /home/mich412/go/src/mak_common/sv.sh
package mversion

import (
	"fmt"
)

type VersionDescr struct {
	//IndMark  string //210720 15:22 = "<index>", for example "0"//removed 210721 14:07
	ProgName string // E.g: "RELS", "mak_common.msess"; it is formed immediately (directly) by an application programmer.
	Number   string //It is set initially by on application programmer to "---nv no env---"
	// and next may be set by a script to its actual value.
	Descr string //The text description of the version; it is formed immediately (directly) by an application programmer.
}

type VersionList []VersionDescr

func (vl VersionList) GetCurrVerInfo() string {
	return fmt.Sprintf("%v:%v", vl[0].ProgName, vl[0].Number)
}

func (vl VersionList) GetCurrVerNum() string {
	return vl[0].Number
}

//versinList defines VersionDescr for each version that has been occured
//Current version has index 0, previous one - index 1, and so on.
//var versionList = mversion.VersionList{
//	{"mak_common.msess", "nv no env", blabla_210311},
//}

//Sequence next constants define textual descriptions of versions.
/*
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
Предложение к следующей версии:<br>---
`
*/
