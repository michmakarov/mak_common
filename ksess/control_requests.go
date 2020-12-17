// control_requests
package ksess

type writeToFrontLog func(string)

/* 201203 06:46
func doControlRequest(w http.ResponseWriter, r *http.Request, wtfl writeToFrontLog) (yes bool) {
	yes = true
	switch r.URL.Path {
	case "/post_agent_error_report":
		if ok := flr.checkCommandRequst(w, r); ok {
			wtfl("accepted")
		} else {
			wtfl("refused")
			return
		}
		postAgentErrorReportHandler(w, r)
		yes = true
		return
	case "/get_agent_worker":
		if ok := flr.checkCommandRequst(w, r); ok {
			wtfl("accepted")
		} else {
			wtfl("refused")
			return
		}
		getAgentWorkerHandler(w, r)
		yes = true
		return
	default:
		yes = false
	}
	return
}
*/
