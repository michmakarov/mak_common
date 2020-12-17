// handlers
package handlers

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"runtime"
	"strconv"
)

func ShowHTTPMessage(w http.ResponseWriter, r *http.Request, mess string) {
	var dataPage struct {
		Mess       string
		MemState   string
		Goroutines string
	}

	var (
		err              error
		page             *template.Template
		templateFileName string
		ps               string

		mStats = runtime.MemStats{}
	)

	defer func() {
		if rec := recover(); rec != nil {
			w.Write([]byte(fmt.Sprintf("ShowHTTPMessage takes the cake: %v\n", rec)))
		}
	}()
	ps = string(os.PathSeparator)
	templateFileName = os.Getenv("GOPATH") +
		ps + "src" + ps + "kot_common" + ps + "" + ps +
		"handlers" + ps + "html" + ps + "mess.html"
	dataPage.Mess = mess
	dataPage.MemState = fmt.Sprintf("<p>GR HeapAlloc=%v;HeapSys=%v</p>", mStats.HeapAlloc, mStats.HeapSys)
	dataPage.Goroutines = fmt.Sprintf("<p>GR number=%v,HeapSys=%v</p>", strconv.Itoa(runtime.NumGoroutine()))

	page, err = template.ParseFiles(templateFileName)
	if err != nil {
		panic(err.Error())
	}

	var buf bytes.Buffer

	err = page.Execute(&buf, dataPage)
	if err != nil {
		panic(err.Error())
	}

	err = page.Execute(w, dataPage)
	if err != nil {
		panic(err.Error())
	}

}
