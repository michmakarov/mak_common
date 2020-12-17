//190926 About versionung
//191029 - this file
package kutils

import (
	"fmt"
	"net/http"
	"strconv"
)

var commit_data_1 = "No_git_data"

//VersionDescr binds a version number with the version description
type VersionDescr struct {
	Number       string //Version number is a string of format "190926", that is "<Year><Month><Day>"
	ProgName     string //   = "kitils"
	GitBranch    string // Branch of kot_common
	Git_commit   string
	IlnVer       string //
	VersionState string //= "developing"
	Text         string
}

//versinList defines VersionDescr for each version that has been occured
//Current version has index 0, previous one - index 1, and so on.
var versionList = []VersionDescr{
	{"191225", "kutils", "branch?", commit_data_1, "IlnVer?", "developing", blabla_191225},
	{"191029", "ksodd", "pgf_2", commit_data_1, "191029", "developing", blabla_191029},
}

func GetVesionInfo() string {
	return getVerInfo(versionList[0].Number)
}

func GetVerNum() string { return versionList[0].Number }

//Constants with names of patten blabla_<version number> define textual description of a version.

const blabla_191225 = `
This was inspired by working Ver=ksodd(new-cs-mak2)191223,
namely by func h_Recover_New(rec interface{}, r *http.Request) (res []byte).
In more wide approach: how to convert some panic recovery result to message for some user?
План:<br>
Получить функцию, которая конвентирует сообщение паники в читабельный текст и
(опционально) отправляет email разработчику
<br>
191225 11:44 There is func messFronPanic(rec interface{}, contextUser int, send bool) (mess []byte)
But how to test it?
<br>
Результаты<br>
<br>
<br>
Предложение к следующей версии:<br>
`

const blabla_191029 = `
План:<br>
Это первая версия в новом старом формате.<br>
То есть, отныне работа над версией будет описыватся этими константами (blabla_...)<br>
<br>
<br>
Результаты<br>
1.Заложен programer_manual.odt, куда переписано с некоторым переосмыслением содержимое kutils_190801.odt<br>
2.Реализованы (и описаны в programer_manual.odt) функции для отсылки сообщений SetDevMailSettings, SetDevMailSettings, и SendDeveloper(subject, text string)<br>

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
