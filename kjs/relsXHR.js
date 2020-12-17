
"use strict";

//201210 05:57
//This is a library for the RELS project.
//_______14:01 cookies: https://stackoverflow.com/questions/42260885/send-a-cookie-with-xmlhttprequest-tvmljs
//201212 08:41 How to maintain versions of the library as well ones of it functions.

function CheckLibraryLoading(){
	alert("qqq");
	var testEl = document.getElementById("CheckLibraryLoading");
	testEl.innerHTML="The  library was loaded successfully. You may use it.";
}

//201210 06:13 ______13:18
//This is an object represented a result of a http request.
function UrlResult(){
	this.method = undefined;
	this.url = undefined;
	this.code = undefined;//if Code>0 it is a http status code (200,400 and such). Otherwise, an answer was not got and an error have took place.
	this.type = undefined;// A type of the body. It have sense if code > 0
	this.body = undefined;//A result representetion. if code>0 it is http ancawer body, otherwise error event
}

//201210 06:20 _13:00 _14:25 _14:55; 201211 08:35 _09:45
//urlAnswer does only a Get request with the empty body.
//onGetResult is a callback function that should take an UrlResult object.
//This function will be called when a process of obtaining a result is ended by receiving an ancwer or an any kind error.
function urlAnswer(url, answerBodyType, tymeout, onGetResult){
		var xhr = new XMLHttpRequest();
		var urlResult = new UrlResult();
		urlResult.method = method;
		urlResult.url = url;
		
		var onLoadFun = function(e){//success of waiting - response have come
			//console.log("execRequest :onLoadFun: e=="+e);
			urlResult.code=xhr.status;
			urlResult.type = answerBodyType
			urlResult.body=xhr.responce;
			onGetting(urlResult)
		};//var onLoad
		
		var onErrorFun = function(e){
			//console.log("execRequest :onErrorFun: e=="+e);
			urlResult.code=-1;
			urlResult.type = "";
			urlResult.body=e;
			onGetting(urlResult)
		};
		var onProgressFun = function(e){//success of waiting - response have come
			//console.log("execRequest :onProgressFun: e=="+e);
			doOnProgress(e);
		};//var onLoad

		xhr.onload = onLoadFun;
		xhr.onprogress=onProgressFun;
		xhr.onerror = onErrorFun;

		xhr.open("GET", url);
		xhr.responseType = bodyType;
		xhr.timeout = timeout;
		xhr.send();
		xhrResult.xhr=xhr;
		return xhr;
}
