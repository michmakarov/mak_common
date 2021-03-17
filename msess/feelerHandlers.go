// feelerHandlers
package msess

import (
	"fmt"
	"net/http"
)

func qqmain() {
	fmt.Println("Hello World!")
}

//This func invokes by func (f *feeler) ServeHTTP if call of getCookieData gives error
//and the request is index request
//That is it is helper function that may be called only in above pointed place.
func indexHandler(w http.ResponseWriter, r *http.Request, cD SessCookieData) {
	//getCookieData(r) gave error
	agentRegistered(cD, r)
}
