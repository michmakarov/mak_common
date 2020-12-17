
"use strict";

//(190228)This comment is for checking how "Cache-Control:no-store" works
//And it works as if.
//Under Chromium with no checking flag "Disable cache" these coments comes
//But what will be if we have not send the header?
//Oh! We have not received fourth line. Now we switch the "PRODUCT" off again.
//And we have all five lines that are above


//For what does it serve? I just have liked it, it's all
function makeid(len) {
  var text = "";
  var possible = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789";

  for (var i = 0; i < len-1; i++)
    text += possible.charAt(Math.floor(Math.random() * possible.length));

  return text;
}
//It takes XMLHttpReuest and returns a file name or throws exceptions
function getFileNameFromResponse(xhr){
	if (typeof(xhr)!="object"){
		throw "kjs.xhr.getFileNameFromResponse: the xhr is not object; it=="+xhr;
	}
	if (xhr.constructor.name!=="XMLHttpRequest"){
		throw "kjs.xhr.getFileNameFromResponse: the xhr is not XMLHttpReuest; it=="+xhr.constructor.name+"==";
	}
	if (xhr.readyState!=4){
		throw "kjs.xhr.getFileNameFromResponse: the xhr has not readyState of 4; it=="+xhr.readyState;
	}

    var filename = "";
	var contentDisposition = xhr.getResponseHeader("Content-Disposition")
    if (contentDisposition && contentDisposition.indexOf('attachment') !== -1) {
        var filenameRegex = /filename[^;=\n]*=((['"]).*?\2|[^;\n]*)/;
        var matches = filenameRegex.exec(contentDisposition);
        if (matches != null && matches[1]) { 
          filename = matches[1].replace(/['"]/g, '');
        }
    }
    if (filename===""){filename="unknownFile"};
    return filename;	
}

//returns the amount of millisecond that was spended for performing of saving
function saveBlodAsFileWithMeasureTime(blb, fileName){
	var start = performance.now();
    var a = document.createElement("a");
    //var blb = new Blob(body)
    document.body.appendChild(a);
    a.style = "display: none";
    var url = window.URL.createObjectURL(blb);
    //var url = "data:,"+body;
    a.href = url;
    a.download = fileName;
    a.click();
    window.URL.revokeObjectURL(url);
    return ((performance.now() - start)+"(mls)");
}


//190305 For what this constructor is needed?
//I have to admit that in fact it is consequence of weak understanding of javascript as language
//If throw off all more dim thoughts one clear need remains: to define a list of properties and methods!
//To say more definitely I want a class in classical sense; so let it be.
//How is the class name? Let it be "XhrResult" as it is said, for historical reason
//This class have single setter of "setResult", that establishs the result of some process.
//And this have three getters, namely "status", "proc" and "err".
function XhrResult(){
	var panicPrefix = "XhrResult:"
	var proc = null;
	var err = null;
	var status = -1// -1 the object does not carry information about any processes
		//0 The process has been fulfilled and successfully ended
		//1 The process has been fulfilled (carried out) but not successfully in its internal sense
		//2 The process has not been carried out

	this.setResult=function(sts, prc, error){
		switch (sts){
			case 0: if (typeof(proc)!=="object")
				{throw panicPrfix+"setResult (sts == 0): not allowed value of second parameter -"+proc};
				proc=prc; status=sts; return;
			case 1: if (typeof(proc)!=="object")
				{throw panicPrfix+"setResult (sts == 0): not allowed value of second parameter -"+proc};
				err=error; proc=prc; status=sts; return;
			case 2:if (!error)
				{throw panicPrfix+"setResult (sts == 2): no value of third parameter 'error'"};
				err=error; proc=prc; status=sts; return;
			default: throw panicPrfix+"setResult: illegal status =="+sts;
		}
	}

	this.status=function(){return status;}
		
	this.err=function(){
		switch (status){
			case -1: return "No process";
			case 0: return "No error; status === 0";
			case 1: case 2: return err;
			default:throw panicPrfix+"internal problem: status="+status;
		}
	}

	this.proc = function(){return proc};
}

//The function disguise an error event as an execution event.
//That is on load and on error the same callback function will be called 
//The function doOnExec is called when the request will have been ended with success or without
//On calling it will be given as parameter an object xhrResult

//The function create XMLHttpRequest (the xhr variable) and XhrResult (the xhrResurt variable) objects
//The function returns "xhr" after the corresponded request will be sent.
//If some function will have obtained the "xhr", it may show  changing  of object through (e. g.) inctance of CreateXhrReport.
//An instance of XhrResult is accessible through a callback of "doOnExec" that is the function will give the obtect as parameter of th callback.
function execRequest(reqTag, doOnExec, doOnProgress, method, uri, reqBody, responseType){
		var xhr = new XMLHttpRequest();
		
		var xhrResult= new XhrResult();
		xhrResult.tag=reqTag;
		xhrResult.uri=uri;
		
		var onLoadFun = function(e){//success of waiting - response have come
			console.log("execRequest :onLoadFun: e=="+e);
			if (xhr.status==200){
				xhrResult.setResult(0, xhr, null);
			}else{
				var s="no response text";
				if (xhr.responseText){s=xhr.responseText.substring(0,150)};
				xhrResult.setResult(1, xhr, xhr.status+"("+xhr.statusText+")-- " + s);
			}
			doOnExec(xhrResult);
		};//var onLoad
		var onErrorFun = function(e){
			//console.log("execRequest :onErrorFun: e=="+e);
			xhrResult.setResult(2, xhr, "execRequest :onErrorFun: e=="+e);
			doOnExec(xhrResult);
		};
		var onProgressFun = function(e){//success of waiting - response have come
			//console.log("execRequest :onProgressFun: e=="+e);
			doOnProgress(e);
		};//var onLoad

		xhr.onload = onLoadFun;
		xhr.onprogress=onProgressFun;
		xhr.onerror = onErrorFun;
		
		//xhr.oncancel = function (){
		//	console.log("execRequest :onProgressFun: e=="+e);
		//}

		xhr.open(method, uri);
		switch (responseType){
			case "":
			case "text":
				xhr.responseType = "text";
				break;
			case "blob":
				xhr.responseType = "blob";
				break;
			default :
				throw "kjs.xhr.execRequest: Not allowed response type ="+responseType;
		}
		if (!reqBody) {xhr.send();} else {xhr.send(body);};
		xhrResult.xhr=xhr;
		return xhr;
}



//This function create an object through which some process can show its flowing by the object's methods
//For each object it create some report element that appearance of which will be affected by those methods.
//The reportContainerId (the first parameter of this function, string)
// is an identifier of an element inside that element (a report elemnt) will be created.
//The process can be canceled via second parameter; it must be an object with method "abort".
//Third parameter must be string that represents the process
//The object has inner flag "finished"(it is false in beginning) that is ruled and rules behavior next methods.
//Method "progressText"; if finished==false it does nothing, otherwise the text is displayed for represent of flow of the process.
//Method "resultText"; It establishs finished=true and displays the text as represantation of the process result.
function CreateXhrReport(reportContainerId, proc, title){
	var finished = false;
	var cancelErr="canceled successfully";
	var progressCouter=0
	//var rep = {};
	if (typeof(reportContainerId)!=="string")
		{throw "createXhrReport: reportContainerId parametre must be string, but is " + reportContainerId}
	
	if (typeof(proc.abort)!=="function")
		{throw "createXhrReport: proc must have the abort method, but has " + proc.abort}
	
	if (typeof(title)!=="string")
		{throw "createXhrReport: title parametre must be string, but is " + title}

	var reportContainer = document.getElementById(reportContainerId);
	if (!reportContainer){throw "createXhrReport: there is not such "+ reportContainerId}
	
	var report = document.createElement("p");
	report.style.border="dotted";
	
	var span= document.createElement("span");
	
	var titleSpan= document.createElement("span");
	titleSpan.innerHTML = title +"<br>";
	report.appendChild(titleSpan);

	var button = document.createElement("button");
	button.type="button";
	button.innerHTML="_CANCEL_"
	button.onclick=function(){
		try {
			//console.log("CreateXhrReport :_CANCEL_.onclick"+proc+"--<cancel>--");
			//cancel();
			proc.abort();
		}
		catch (e) {cancelErr="cancel throw exeption:"+e}
		button.innerHTML="_DELETE_REPORT_"+cancelErr
		button.onclick=function(){
			reportContainer.removeChild(report)
		}
	};
	report.appendChild(button);
	report.appendChild(span);
	reportContainer.appendChild(report);
	
	//rep.abort = function(){
	//	xhr.abort;
	//}
	this.progressText = function(text){
		if (finished) {return};
		progressCouter++;
		span.innerText=text;
	}
	this.resultText = function(text){
		finished = true;
		span.innerHTML=text;
		button.innerHTML="_DELETE_REPORT_progressCouter="+progressCouter;
		button.onclick=function(){
			reportContainer.removeChild(report);
		};
	}
	
	//return rep;
}

//This function makes  asynchronously a HTTP request and regards the response body as blob that must be save on disk.
//That is, if the response code is 200 it calls "saveBlodAsFileWithMeasureTime" (with "getFileNameFromResponse").
//1. It binds a report object with an XMLHttpRequest object through callback functions "doOnExec" and 'doOnProgress"
//2. It creates the XMLHhttpRequest object by call of execRequest
//3. It creates the report object by call of createXhrReport.
//4. It defines the response body as blob
function makeRequestForBlob( methodId, uriId, reqBody, reportContainerId){
	var start = performance.now();
	var xhrResult;
	var xhrReport;
	var count=0;
	var progress="";

	var methodInpEl = document.getElementById(methodId);
	var uriInpEl = document.getElementById(uriId);


	var doOnExec = function(xhrResult){
		var totalDur = performance.now() - start;
		switch(xhrResult.status()){
			case 0:
				var saveDur = saveBlodAsFileWithMeasureTime(xhrResult.xhr.response,
					getFileNameFromResponse(xhrResult.xhr));
				xhrReport.resultText("-Exec result : RESP TYPE="+xhrResult.xhr.responseType + ": Headers-------------<br>"
				+ xhrResult.xhr.getAllResponseHeaders()+"totalDur=\n"+totalDur+"saveDur=\n"+saveDur);
				break;
			case 1:case 2:
				xhrReport.resultText("-Exec result : RESP ERROR,xhrResult.status()"+xhrResult.status()) + ": Headers-------------<br>"
				break;
			default:
			throw "makeRequestForBlob : Let the hell takes such programmers;xhrResult.status()="+xhrResult.status();
			
		}
	}
	
	var doOnProgress = function(e){
		count++
		progress="-"+e.loaded+" of ("+e.total+")";
		xhrReport.progressText(progress);
	}
	
	var xhr=execRequest("noTag", doOnExec, doOnProgress, methodInpEl.value, uriInpEl.value, reqBody, "blob")
	xhrReport = new CreateXhrReport(reportContainerId, xhr, "no Tag;"+uriInpEl.value)
	
}//qqq2 definition	


//Unlike the makeRequestForBlob (qqq2) this function does not care about  showing progress and cancelation.
//So on obtaining some result it creates an element of <p> and inserts the element into the container in two cases
//The case one takes place 
function qqq3(tag, methodId, uriId, reqBody, reportContainerId,elementForTagsId){
	var xhr;

	var elementForTags = document.getElementById(elementForTagsId);

	var methodInpEl = document.getElementById(methodId);
	var uriInpEl = document.getElementById(uriId);
	var  reportContainer =  document.getElementById(reportContainerId);

	var reportParagraph = document.createElement("p");
	//reportContainer.appendChild(reportParagraph)

	var doOnExec = function(xhrResult){
		switch (xhrResult.status()){
			case 0:
				elementForTags.innerHTML=elementForTags.innerHTML+xhrResult.tag+"_";
				break;
			case 1: case 2:
				reportParagraph.innerHTML=xhrResult.err();
				reportContainer.appendChild(reportParagraph);
				elementForTags.innerHTML=elementForTags.innerHTML+xhrResult.tag+"-err_";
				break;
			default:throw "internal problem: xhrResult.statuss="+xhrResult.status();
			
		}
	}
	
	var doOnProgress = function(e){
		//count++
		//progress=progress+"-"+count;
		//xhrReport.progressText(progress);
		reportParagraph.innerHTML= "Ждем ...";

	}
	
	xhr=execRequest(tag, doOnExec, doOnProgress, methodInpEl.value, uriInpEl.value, reqBody,"")
	//xhrReport = createXhrReport(reportContainerId, xhr, uriInpEl.value)
	
}//qqq3 definition	


