
"use strict";


//210218 12:19 from ksodd. Does it work? That is good and enough!
//returns the amount of millisecond that was spended for performing of saving
function saveBlodAsFileWithMeasureTime(blb, fileName){
	var start = performance.now();
    var a = document.createElement("a");
    //var blb = new Blob(body)
    document.body.appendChild(a);
    a.style = "display: none";
    var url = window.URL.createObjectURL(blb);
    a.href = url;
    a.download = fileName;
    a.click();
    window.URL.revokeObjectURL(url);
    return ((performance.now() - start)+"(mls)");
}




//201210 05:57
//This is a library for the RELS project.
//_______14:01 cookies: https://stackoverflow.com/questions/42260885/send-a-cookie-with-xmlhttprequest-tvmljs
//201212 08:41 How to maintain versions of the library as well ones of it functions.

//201218 04:46 It is the result pondering about the "riddle of alert"
//How a script can identify itself?
var What_am_I = "---51d5bee--*main--210222_2108---" //"201218_05:49; 41; mak_common/kjs/relsXHR.js"

function CheckLibraryLoading(){
	//alert("qqq");
	var testEl = document.getElementById("CheckLibraryLoading");
	testEl.innerHTML="The  agent library ("+What_am_I+") was loaded.";
}

//201210 06:13 ______13:18
//This is an object represented a result of a http request.
function UrlResult(){
	this.method = undefined;
	this.url = undefined;
	this.code = undefined;//if Code>0 it is a http status code (200,400 and such). Otherwise, an answer was not got and an error have took place.
	this.type = undefined;// A type of the body. It have sense if code > 0
	this.body = undefined;//A result representetion. if code>0 it is http ancawer body, otherwise error event
	this.dur = undefined;//a duration from sending a query to obtaining the result
}

//201210 06:20 _13:00 _14:25 _14:55; 201211 08:35 _09:45
//urlAnswer does only a Get request with the empty body.
//onGetResult is a callback function that should take an UrlResult object.
//This function will be called when a process of obtaining a result is ended by receiving an ancwer or an any kind error.
function urlAnswer(url, answerBodyType, timeout, onGetResult, progRepId){
		var xhr = new XMLHttpRequest();
		var urlResult = new UrlResult();
		urlResult.method = "get";
		urlResult.url = url;
		
		var onLoadFun = function(e){//success of waiting - response have come
			//alert("onLoadFun: e=="+e);
			urlResult.code=xhr.status;
			urlResult.type = answerBodyType;
			urlResult.body = xhr.response;
			//xhr.response.text().then(text => alert( "RESPONSE="+text));
			urlResult.dur= (performance.now() - start)+"(mls)";

			onGetResult(urlResult)
		};//var onLoad
		
		var onErrorFun = function(e){
			//console.log("execRequest :onErrorFun: e=="+e);
			urlResult.code=-1;
			urlResult.type = "";
			urlResult.body=e;
			onGetResult(urlResult)
		};
		var onProgressFun = function(e){//success of waiting - response have come
			var rep = document.getElementById(progRepId)
			var count = 0;
			count++;
			if (e.lengthComputable) {
				var percentComplete = e.loaded / e.total * 100;
				rep.innerHTML = percentComplete
			} else {
				// Unable to compute progress information since the total size is unknown
				rep.innerHTML = count
			}		
		}//onProgressFun

		
		xhr.onload = onLoadFun;
		xhr.onprogress=onProgressFun;
		xhr.onerror = onErrorFun;

		xhr.open("GET", url);
		xhr.responseType = answerBodyType;
		xhr.timeout = timeout;
		var start = performance.now();
		xhr.send();
		urlResult.xhr=xhr;
		return xhr;
}


//210105 15:28  
//urlPosAnswer does only a Post request with the given body.
//onGetResult is a callback function that should take an UrlResult object.
//This function will be called when a process of obtaining a result is ended by receiving an ancwer or an any kind error.
function urlPostAnswer(url, answerBodyType , body, content_type, timeout, onGetResult){
		var xhr = new XMLHttpRequest();
		var urlResult = new UrlResult();
		urlResult.method = "post";
		urlResult.url = url;
		
		var onLoadFun = function(e){//success of waiting - response have come
			//alert("onLoadFun: e=="+e);
			urlResult.code=xhr.status;
			urlResult.type = answerBodyType
			urlResult.body=xhr.response;
			onGetResult(urlResult)
		};//var onLoad
		
		var onErrorFun = function(e){
			//console.log("execRequest :onErrorFun: e=="+e);
			urlResult.code=-1;
			urlResult.type = "";
			urlResult.body=e;
			onGetResult(urlResult)
		};
		var onProgressFun = function(e){//success of waiting - response have come
			//console.log("execRequest :onProgressFun: e=="+e);
			//doOnProgress(e);
		};//var onLoad

		xhr.onload = onLoadFun;
		xhr.onprogress=onProgressFun;
		xhr.onerror = onErrorFun;

		xhr.open("POST", url);
		xhr.responseType = answerBodyType;
		xhr.timeout = timeout;
		if (!body) {body="foo=bar&lorem=ipsum"};
		if (!content_type) {content_type="application/x-www-form-urlencoded"}
		xhr.setRequestHeader("Content-type", content_type);
		xhr.send(body);
		urlResult.xhr=xhr;
		return xhr;
}

function exitSession(){
	alert("exitSession() is here!")
}

function downloalFile(fileName, timeOut, progRepId){
	var url="/downfife?a_p_p_n=qqq&fname="+fileName;
	var onGetResult=function(urlResult){
				if (urlResult.code == 200) {
					var saveDur = saveBlodAsFileWithMeasureTime(urlResult.body, fileName);
					var transDur = urlResult.dur;
					var rep = document.getElementById("report")
					rep.innerHTML="transDUr="+transDur+";saveDur="+saveDur;
				}else{
					urlResult.body.text().then(text => alert( "("+urlResult.code+")RESPONSE="+text));
				}
			};
	urlAnswer(url, "blob", timeOut, onGetResult, progRepId
	)		
}

//function ceateFoo(){
//var file = new File(["foo"], "foo.txt", {
//  type: "text/plain",
//});
//}


