

//This is the template file to create dynamically a JS Worker script that sends an agent error report to KOT server


//This variable contains URL to obtain the value of kotAgentSignature
//It contains among others the control password
var kotURLForGetAgentSignature = "{{.KotURLForGetAgentSignature}}"

//This contains the control password which using by function PostErrorReport
var controlPassword ="{{.ControlPassword}}"

var kotAgentSignature



function getAgentSignature(){
	console.log("getAgentSignature here!!! kotURLForGetAgentSignature="+kotURLForGetAgentSignature);
	var oReq = new XMLHttpRequest();
	
	var onLoad = function(){
		if (oReq.status == 200) {
			if (oReq.responseText){
				kotAgentSignature = oReq.responseText;
				console.log("kotAgentSignature="+kotAgentSignature);
			} else {console.error("GetAgentSignature: Status==200 but no responseText")};
		} else {
			console.error("GetAgentSignature: Status!=200 : "+oReq.responseText);
		} // else of oReq.status == 200
	};

	var onError = function(e){
			console.error("GetAgentSignature: error: "+ e);
	};


	oReq.addEventListener("load", onLoad);
	//oReq.addEventListener("error", onError);

	oReq.open("GET", kotURLForGetAgentSignature);
	oReq.send();
	return oReq;
}

function sendErrorReport(err){
	var oReq = new XMLHttpRequest();
	var URI = "/post_agent_error_report?CONTROL_PASSWORD="+controlPassword+"&AGENT_SIGNATURE="+kotAgentSignature;
	var onLoad = function(){
		if (oReq.status == 200) {
			if (oReq.responseText){
				console.log(oReq.responseText)
			} else {console.log("Рапорт об ошибке отослан, 200 - но ответа не представлено")}
		} else {
			console.error("sendErrorReport: Status!=200")
		} // else of oReq.status == 200
	}

	var onError = function(e){
			console.error("+sendErrorReport: error "+ e.message)
	}


	oReq.addEventListener("load", onLoad);
	//oReq.addEventListener("error", onError);

	oReq.open("POST", URI);
	oReq.send(err);
	return oReq;
};

var oReq = getAgentSignature();

onmessage = function (e) {
	oReq = sendErrorReport(e.data);
};



//if (kotAgentSignature){
//	GetAgentSignature();
//	
//	onmessage = function (e) {
//		sendErrorReport(e.data);
//	};
//} else {
//	console.error("kotAgentSignature not established");
//};



