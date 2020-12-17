package kerr

import (
	"fmt"
	"log"
	"path/filepath"
	"reflect"
	"runtime"
	//"sync"
)

func ProcessingError(user_id int, err_text string) (text string) {
	var (
		filename string
		line     int
	)

	_, filename, line, _ = runtime.Caller(1)

	text = fmt.Sprintf("panic: %s:%d: user_id = %d: error = %s",
		filepath.Base(filename), line, user_id, err_text)

	log.Print(text)
	//SendDeveloper("error", text)

	return
}

//if the rec has type error it returns Error(),  otherwise
//if the rec has type string it returns that string,  otherwise
//it returns "kerr.GetRecoverErrorText: rec(%v) is not error or string"
func GetRecoverErrorText(rec interface{}) (text string) {
	var (
		ok  bool
		err error
		s   string
	)

	if err, ok = rec.(error); ok {
		text = err.Error()
	} else {
		if s, ok = rec.(string); ok {
			text = s
		} else {
			text = fmt.Sprintf("kerr.GetRecoverErrorText: rec(%v) is not error or string", rec)
		}
	}
	return
}

func GetRecoverError(rec interface{}) (err error) {
	var (
		s string

		filename string
		line     int
		recType  string
	)
	recType = reflect.TypeOf(rec).Name()
	if recType == "" {
		recType = "Type not defined"
	}
	//fmt.Printf("--M-- GetRecoverError recType: %v\n", recType)
	_, filename, line, _ = runtime.Caller(1)

	switch recType {
	case "error":
		err, _ = rec.(error)
		err = fmt.Errorf("source=%v(line=%v): %v", filename, line, err.Error())
	case "string":
		s, _ = rec.(string)
		err = fmt.Errorf("source=%v(line=%v): %v", filename, line, s)
	default:
		err = fmt.Errorf("source=%v(line=%v): rec = %v ", filename, line, rec)
	}
	return
}
