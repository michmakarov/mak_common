package mutils

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"reflect"
	"regexp"
	"runtime"
	"strconv"
	"sync/atomic"
	"time"

	//"kot_common/kerr"
	"strings"

	"gopkg.in/gomail.v2"
)

type DevMailSettings struct {
	MAIL_USER_EMAIL string
	MAIL_USER_PASS  string
	MAIL_SMTP       string
	MAIL_SMTP_PORT  string
	MAIL_DEVELOPERS string
	AppData         string //additional data that describes an application (name, version and so on)
}

//const Version = "190801" //"190403_closed190416" //"190202" //"190124" //"181005" //"180829"
//const VersionState = "closed190801"

//0 - not printing
var FmtFlag int

var (
	devMailSettings *DevMailSettings
	lettersNumber   int32 //number of sent letters
	lettersTotalDur int64 //total time that was speared for sending all letters that was counted by ettersNumber
	letterMaxDur    int64 // max time that was speared for sending letter
)

func SetDevMailSettings(
	MAIL_USER_EMAIL string,
	MAIL_USER_PASS string,
	MAIL_SMTP string,
	MAIL_SMTP_PORT string,
	MAIL_DEVELOPERS string,
	AppData string,
) {
	devMailSettings = &DevMailSettings{
		MAIL_USER_EMAIL,
		MAIL_USER_PASS,
		MAIL_SMTP,
		MAIL_SMTP_PORT,
		MAIL_DEVELOPERS,
		AppData,
	}
}

func GetMailStatistic() string {
	return fmt.Sprintf("lettersNumber=%v, letterMaxDur=%v, lettersTotalDur=%v", atomic.LoadInt32(&lettersNumber), time.Duration(atomic.LoadInt64(&letterMaxDur)), time.Duration(atomic.LoadInt64(&lettersTotalDur)))
}

//Errror of setting rand.Seed for TrueRand (it is prefix) functions
//That is these functions will return it if it is not equal nil
var err error

func init() {
	var b time.Time

	b, err = time.Parse("02.01.2006", "07.11.1917")
	if err != nil {
		panic("kutils. init: time.Parse returns error")
	}
	rand.Seed(int64(b.Sub(time.Now())))
}

//StrToLen aligns a string to length by truncating it or padding with blank character
func StrToLen(s string, l int) string {
	if len(s) == l {
		return s
	}
	if len(s) > l {
		return s[:l-1]
	}
	for i := 0; i < l-len(s); i++ {
		s = s + "_"
	}
	return s
}

//NULL functions are functions with prefix “Null”.
//Each of function with prefix "Null" presumes that its first parameter carries a corresponding to its name value and returns its string representation.
//E.g. the function “NullTimeToString” presumes that the first parameter is of type “time”.
//If a first parameter is nil the second one (i.e. "ifNil") is returned.
//If the first is not nil and the presumption is not true, then the third (i.e. "ifWhatIs") is returned.
func NullTimeToString(nt interface{}, ifNil string, ifWhatIs string) string {
	if nt == nil {
		return ifNil
	}
	if reflect.TypeOf(nt) == reflect.TypeOf(time.Now()) {
		return (nt.(time.Time)).Format("02.01.2006")
	} else {
		return ifWhatIs
	}
}

func NullIntToString(nt interface{}, ifNil string, ifWhatIs string) string {
	if nt == nil {
		return ifNil
	}
	if reflect.TypeOf(nt) == reflect.TypeOf(int(0)) {
		return strconv.Itoa(nt.(int))
	}
	if reflect.TypeOf(nt) == reflect.TypeOf(int32(0)) {
		return strconv.Itoa(int(nt.(int32)))
	}
	return ifWhatIs
	//return reflect.TypeOf(nt).Name()

}

func NullFloatToString(nt interface{}, ifNil string, ifWhatIs string) string {
	var (
		//f32 float32
		f64 float64
	)
	if nt == nil {
		return ifNil
	}
	if reflect.TypeOf(nt) == reflect.TypeOf(float64(3.14)) {
		f64 = nt.(float64)
		return strconv.FormatFloat(f64, 'f', -1, 32)
	}
	if reflect.TypeOf(nt) == reflect.TypeOf(float32(3.14)) {
		f64 = nt.(float64)
		return strconv.FormatFloat(f64, 'f', -1, 32)
	}
	return ifWhatIs
}

func StrWithoutNull(str string, rep string) (res string) {
	var buff []byte
	if (len(rep) > 1) || (len(rep) == 0) {
		rep = "!"
	} else if rep[0] < 10 {
		rep = "!"
	}
	buff = make([]byte, len(str))
	for i := 0; i < len(str); i++ {
		buff[i] = str[i]
		if buff[i] < 9 {
			buff[i] = rep[0]
		}
	}
	res = string(buff)
	return

}

//it returns string without null characters
func EraseNulls(str string) (res string) {
	var buff []byte
	var i, ii int
	//if len(str) == 0 {
	//	return
	//}
	buff = make([]byte, len(str))
	for i = 0; i < len(str); i++ {
		if str[i] != 0 {
			buff[ii] = str[i]
			ii++
		}
	}
	//if ii == 0 {
	//	res = str
	//	return
	//}
	res = string(buff[:ii])
	return

}

//if function with prefix "TrueRand" returns error it also returning a random value but from default seed
func TrueRandInt() (string, error) {
	return strconv.Itoa(rand.Int()), nil
}

//if function with prefix "TrueRand" returns error it also returning a random value but from default seed
func TrueRandIntAsInt() (int, error) {
	return rand.Int(), nil
}

//180911
//It extracts bytes from beginning up to encounting a byte outside interval of [48 - 57]  or of the end of slice.
//That is it reads digits 0 ... 9. If the such there are not it retuns "-1".
//Otherwise it renders the bytes having been read as a record of decimal number and returns a string presentation of this
//E.g., the {49,48,48,13,0,0,0} it renders as "100", the {48,48,13,0,0,0} it renders as "0"
func ExtractCommandPar(rawPar []byte) string {
	var res = "-1"
	var byteToDigit = func(bt byte) string {
		switch bt {
		case 48:
			return "0"
		case 49:
			return "1"
		case 50:
			return "2"
		case 51:
			return "3"
		case 52:
			return "4"
		case 53:
			return "5"
		case 54:
			return "6"
		case 55:
			return "7"
		case 56:
			return "8"
		case 57:
			return "9"
		}
		return "-1"
	}

	if rawPar == nil {
		return res
	}
	res = ""
	for i := 0; i < len(rawPar); i++ {
		if rawPar[i] >= 48 && rawPar[i] <= 57 {
			res = res + byteToDigit(rawPar[i])
		} else {
			break
		}
	}
	if res == "" {
		res = "-1"
	}
	return res
}

func ExtractCommandParInt(rawPar []byte) int {
	var res = ExtractCommandPar(rawPar)
	var resInt = -1
	resInt, _ = strconv.Atoi(res)
	return resInt
}

func GetApplName() string {
	return os.Args[0]
}

//190124_2 - fnt wraps
func Printf(format string, args ...interface{}) {
	if FmtFlag == 0 {
		return
	}
	fmt.Printf("--M--"+format, args...)
}

//TraceStack returns a goroutin stack presenting as a sequence of function names
//E. g. "f - f1 - ...", where f is the name of a function that have been called by TraceStack itself
//Since 190403  (wrote 190409)
func TraceStack() (stack string) {
	var f *runtime.Func
	pc := make([]uintptr, 10) // at least 1 entry needed
	runtime.Callers(2, pc)
	for _, item := range pc {
		f = runtime.FuncForPC(item)
		stack = stack + f.Name() + " - "
	}
	return
}

//IntfToInt returns interface{} as int if it is possible. Otherwise it returns false
func IntfToInt(v interface{}) (int, bool) {
	if v == nil {
		return -11, false
	}
	var ok = true
	switch v.(type) {
	case uint:
		{
			return int(v.(uint)), ok
		}
	case int:
		{
			return int(v.(int)), ok
		}
	case int32:
		{
			return int(v.(int32)), ok
		}
	case int64:
		{
			return int(v.(int64)), ok
		}
	}
	return -1, false
}

//IntfToString returns v as string if it is possible
func IntfToString(v interface{}) (string, bool) {
	if v == nil {
		return "no string but nil", false
	}
	var ok = true
	switch v.(type) {
	case string:
		{
			return v.(string), ok
		}
	}
	return "no string", false
}

//190703 ------------------------------------

//UpdateStruct updates components of first (old) structure by values of not zero values of second (updStr) structure.
//That is both parameters must be structures of the same type
//It has not rolled. As it turns out I do not know the reflection enough
func UpdateStruct(oldStr, updStr interface{}) (err error) {
	//var oldFields, updFields []reflect.StructField
	if reflect.TypeOf(oldStr).Kind() != reflect.Struct {
		err = fmt.Errorf("UpdateStruct: argument is not struct")
		return
	}
	if reflect.TypeOf(oldStr).Name() != reflect.TypeOf(updStr).Name() {
		err = fmt.Errorf("UpdateStruct: arguments are not the same type")
		return
	}
	//oldFields=reflect.TypeOf(oldStr).
	for i := 0; i < reflect.TypeOf(oldStr).NumField(); i++ {
		if reflect.ValueOf(updStr).String() != "" {
			reflect.ValueOf(oldStr).Field(i).Set(reflect.ValueOf(updStr).Field(i))
		}
	}
	return
}

//190704 to get wanted lines from a text file
func GetWantedLines(r io.Reader, filter string) (lns []string, err error) {
	var lineCount int
	var matched bool
	//if filter == "" {
	//	err = fmt.Errorf("kutils.GetWantedLines: the filter must not be empty.")
	//	lns = nil
	//	return
	//}
	scanner := bufio.NewScanner(r)
	pattern := ".*" + filter + ".*"
	for scanner.Scan() {
		lineCount++
		if matched, err = regexp.Match(pattern, scanner.Bytes()); err != nil {
			err = fmt.Errorf("kutils.GetWantedLines: matching, line=%v; err=%v", lineCount, err.Error())
			lns = nil
			return
		}
		if matched {
			lns = append(lns, scanner.Text())
		}
	}
	if err = scanner.Err(); err != nil {
		err = fmt.Errorf("kutils.GetWantedLines:scanning, line=%v; err=%v", lineCount, err.Error())
		lns = nil
		return
	}
	return
}

//191028struct tags
func SendDeveloper(subject, text string) {
	var (
		arr_to []string
		m      *gomail.Message
		d      *gomail.Dialer
		err    error
		port   int
		start  time.Time
		dur    int64
		maxDur int64
	)
	start = time.Now()
	atomic.AddInt32(&lettersNumber, 1)
	defer func() {
		dur = int64(time.Since(start))
		atomic.AddInt64(&lettersTotalDur, dur)
		maxDur = atomic.LoadInt64(&letterMaxDur)
		if dur > maxDur { //191029 between this and next line may be a time gap but I think the probability of that that in the gap more value has come is very small
			atomic.StoreInt64(&letterMaxDur, dur)
		}
		if rec := recover(); rec != nil {
			//log.Print(GetRecoverErrorText(rec))
			//kerr.SysErrPrintf("Panic of kutils.SendDeveloper; rec=%v", rec)
			panic(fmt.Sprintf("Panic of mutils.SendDeveloper; rec=%v", rec))
		}
	}()

	if devMailSettings == nil {
		//kerr.SysErrPrintf("kutils.SendDeveloper: no mail settings")
		//return
		panic(fmt.Sprint("mutils.SendDeveloper: no mail settings"))

	}

	arr_to = strings.Split(devMailSettings.MAIL_DEVELOPERS, ",")

	m = gomail.NewMessage()
	m.SetHeader("From", devMailSettings.MAIL_USER_EMAIL)
	m.SetHeader("To", arr_to...)

	if subject != "" {
		m.SetHeader("Subject", subject)
	} else {
		m.SetHeader("Subject", "Prog problem")
	}

	text = "(" + devMailSettings.AppData + ")" + text
	m.SetBody("text/html", text)

	port, err = strconv.Atoi(devMailSettings.MAIL_SMTP_PORT)
	if err != nil {
		panic(err)
	} else {
		d = gomail.NewDialer(devMailSettings.MAIL_SMTP, port, devMailSettings.MAIL_USER_EMAIL, devMailSettings.MAIL_USER_PASS)
		if err = d.DialAndSend(m); err != nil {
			panic(err)
		}
	}

}


func PrintFileContent(fName string){
	var err error
	var f *os.File
	var buff []byte
	if f, err = os.Open(fName); err!=nil{
		fmt.Printf("File:%v:Open err=%v\n",fName, err.Error())
		return
	}
	if buff, err = ioutil.ReadAll(f); err!=nil{
		fmt.Printf("Reading of:%v: err=%v\n",fName, err.Error())
		return
	}
	fmt.Println(string(buff))

}

