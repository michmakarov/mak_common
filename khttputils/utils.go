package khttputils

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
)

//copies headers of "Set-Cookie" from given http.Responce to new http.Responce
//It returns values of found cookies if any
func CopySetCookies(sR *http.Response, dW http.ResponseWriter) (cookies []string) {
	cookies = make([]string, 0)
	for _, v := range sR.Cookies() {
		if v.Name == "Set-Cookie" {
			http.SetCookie(dW, v)
			cookies = append(cookies, v.String())
		}
	}
	return
}
func CopyRequestBody(r *http.Request) (newBody io.Reader) {
	newBody = bufio.NewReader(r.Body)
	return
}

func ReqDescr(tag string, r *http.Request) (descr string) {
	descr = fmt.Sprintf("(%v)--%v:%v TO %v(%v)", tag, r.Method, r.RequestURI, r.Host, r.RemoteAddr)
	return
}

func ReqLabel(r *http.Request) (l string) {
	var IP string
	//var port string
	var fields int
	var fieldsSlice []string

	fieldsSlice = strings.Split(r.RemoteAddr, ":")
	//fmt.Printf("--M--ReqLabel RemoteAddr=%v; fieldsSlice=%v", r.RemoteAddr, fieldsSlice)

	fields = len(fieldsSlice)
	if fields < 2 {
		l = fmt.Sprintf("%s;%s;?IP", r.Method, r.RequestURI)
		return
	}
	//port = fieldsSlice[fields-1]

	fieldsSlice = fieldsSlice[0 : fields-1]
	IP = strings.Join(fieldsSlice, "")
	l = fmt.Sprintf("%s;%s;%s", r.Method, r.RequestURI, IP)
	return
}

//String represantation of an incoming HTTP request
//Fully similar to ReqLabel exept that it is showing remote port
func ReqLabel_2(r *http.Request) (l string) {
	var IP string
	var PORT string
	var sArr []string
	sArr = strings.Split(r.RemoteAddr, ":")
	IP = sArr[0]
	if len(sArr) > 1 {
		PORT = sArr[1]
	} else {
		PORT = "?"
	}
	l = fmt.Sprintf("%s;%s;%s:%s", r.Method, r.RequestURI, IP, PORT)
	return
}

func Grt_IP_Port(r *http.Request) (ip, port string) {
	//var IP string
	//var port string
	var fields int
	var fieldsSlice []string

	fieldsSlice = strings.Split(r.RemoteAddr, ":")
	//fmt.Printf("--M--ReqLabel RemoteAddr=%v; fieldsSlice=%v", r.RemoteAddr, fieldsSlice)

	fields = len(fieldsSlice)
	if fields < 2 {
		ip = r.RemoteAddr
		port = "?"
		return
	}
	port = fieldsSlice[fields-1]

	fieldsSlice = fieldsSlice[0 : fields-1]
	ip = strings.Join(fieldsSlice, "")
	return
}

func RawRequest(r *http.Request) (rslt string) {
	var (
		err error
		buf bytes.Buffer
	)
	err = r.Write(&buf)
	if err == nil {
		rslt = buf.String()
	} else {
		rslt = "khttputils.RawRequest error : " + err.Error()
	}
	return
}

func RawResponse(r *http.Response) (rslt string) {
	var (
		err error
		buf bytes.Buffer
	)
	err = r.Write(&buf)
	if err == nil {
		rslt = buf.String()
	} else {
		rslt = "khttputils.RawResponse error : " + err.Error()
	}
	return
}

func PrintHeader(h http.Header) {
	fmt.Println("Headers _____________________________")
	for k, v := range h {
		fmt.Println(k, " : ", v)
	}
	fmt.Println("____________________________________")
}

func PrintHeaderWithTitle(h http.Header, t string) {
	fmt.Println(t)
	for k, v := range h {
		fmt.Println(k, " : ", v)
	}
	fmt.Println("____________________________________")
}

func PrintRequest(r *http.Request) {
	fmt.Println(r.Host, r.RequestURI, "__", r.Method)
	PrintHeader(r.Header)
}

func PrintResponce(res *http.Response) {
	var (
		i int = -1
		c *http.Cookie
	)
	fmt.Println(res.Status)
	for i, c = range res.Cookies() {
		fmt.Println("c=", c.String())
	}
	if i == -1 {
		fmt.Println("-------No coookies---------")
	}
	PrintHeader(res.Header)
}

func Headers(h http.Header, nl string) (headers string) {
	var line string
	for k, v := range h {
		line = fmt.Sprint(k, " : ", v, nl)
		headers = headers + line
	}
	return
}

//210214 12:24 Potensial_Error: Why ct[0] end ctype do not cast to
func CheckCT(r *http.Request, ctype string) (err error) {
	var ct []string = r.Header.Values(http.CanonicalHeaderKey("Content-Type"))
	if len(ct) == 0 {
		err = fmt.Errorf("Content-Type header not set")
		return
	}
	if len(ct) > 1 {
		err = fmt.Errorf("More than one value of Content-Type")
		return
	}
	if ct[0] != ctype {
		err = fmt.Errorf("Content-Type: expecting %v; is %v", ctype, ct[0])
		return
	}
	return
}

func validateID(ID string) (err error) {
	var fingers = make(map[string]bool)
	fingers["0"] = true
	fingers["1"] = true
	fingers["2"] = true
	fingers["3"] = true
	fingers["4"] = true
	fingers["5"] = true
	fingers["6"] = true
	fingers["7"] = true
	fingers["8"] = true
	fingers["9"] = true
	if len(ID) == 0 {
		err = fmt.Errorf("Empty identifier")
		return
	}
	if ID == "0" {
		return
	}
	for true {
		if string(ID[0]) == "0" {
			ID = strings.TrimPrefix(ID, "0")
		}
		if string(ID[0]) != "0" {
			return
		}
	}
	for i := 0; i < len(ID); i++ {
		if fingers[string(ID[i])] == false {
			err = fmt.Errorf("Illegal (%v) for ID letter", ID[i])
			return
		}
	}
	return
}
