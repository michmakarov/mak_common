"use strict";




/*

function congratulation(user){
	var c = {};
	c.action_name="hello";
	c.user_id=user;
	c.timestamp=new Date();
	return JSON.stringify(c);
}

function checkCongrAnswer(rawCongrAnswer){
	var cR={};//checking result; An object with properties err (bool), errMess(string) and dur (integer) - time (millisekomd) of obtaining answer
	var cA;//An answer to congratulation as object (not raw string)
	try{
		cA=JSON.Parse(rawCongrMsg);
		if(cA.action_name!=="hello"){cR.err=true;cR.errMess="Answer: action-name is undefined or not 'hellow'"; return cR;};
		if(cA.user_id!=UsedId){cR.err=true;cR.errMess="Answer: user_id is undefined or not "+UserId;return cR;};
		cR.err=false; return cR;
	}catch(e){
		cR.err=true;cR.errMess="checkCongrAnswer: JSON.Parse error = "+e;return cR;
	}
}

*/


//checkInCome attempts to applay JSON.Parse to rawIncome parameter.
//It returns an object (cR) with attributes 'err' (type of boolean), 'errMess' (type of string), and 'data' (type of object)
//If cR.err!==true the attempt is failed and cR.errMess comprises an error massage.
//If cR.err!==false cR.data comprises an object that represents incoming message.
//This object may comprise any atributes but with one obligated named 'action_name' (type of string, non empty)
function checkIncome(rawIncome){
	var cR={};
	var income;
	try{
		income=JSON.pParse(rawIncome);
		if(typeof(income.action_name) !== "string") {cR.err=true;cR.errMess="checkIncome: action-name is undefined or not 'string'";
			 return cR;};
		if(income.action_name.trim().length == 0) {cR.err=true;cR.errMess="checkIncome: action-name is empty";
			 return cR;};
	}catch(e){
		cR.err=true;cR.errMess="checkInCome: JSON.Parse error = "+e;return cR;
	}
	cR.err=false;cR.errMess="";cR.data=income;return cR; 
}

//checkOutcome is like checkIncome but applaying JSON.stringify to the object outcomeObj and returning the result into cR.data
function checkOutcome(outcomeObj){
	var cR={};
	var outcome;
	if(typeof(outcomeObj.action_name) !="string") {cR.err=true;cR.errMess="checkOutcome: action-name is undefined or not 'string'";
			 return cR;};
	if(outcomeObj.action_name.trim().length == 0) {cR.err=true;cR.errMess="checkOutcome: action-name is empty";
			 return cR;};
	try{
		outcome=JSON.stringify(outcomeObj);
	}catch(e){
		cR.err=true;cR.errMess="checkOutcome: JSON.stringify error = "+e;return cR;
	}
	cR.err=false;cR.errMess="";cR.data=outcome;return cR;
}





//Kjs_socket returns object that represent Websocket connection to a server pointed by the url parameter.
//An object may be in two main states - working state and disabled state (or error state).
//These two main states are expresed by error value, that is false or true respectively. if error value is true the object is disabled forever.
//Those values are returned by method getError; if the value is true the method getErrorMessage returns the reason of error, otherwise it returns empty string.
//More precisely state of an object is expressed by integer value named "status". This value is returned by method "getStatus" and may by as...
//-1 - the constructor WebSocket gives exeption;
//-2 - the socket fired an error; the object is disabled forever; error state.
//-3 - the socket was closed by server; the object is disabled forever; error state.
//-4 - checkInCome gave error.
//-5 - socket gave message to object with status!==1.

//0 - The initial state; if it is then the constructor has not been working properly.
//1 - The event 'open' has been fired;
//
//"onreceive" function returns nothing and expects an object with three properties:
//status - it is the current status of the object
//msg - Some message string that explains the returning result.
//inObj - it is null if status is not one, otherwise it carries an incoming WS message rendered as valid object
//That is the function says what is occured: (1) connection closed, (2)some error occured, (3) a message from the server is come
//The function envoking only when some event is fired by enclosed WebSocket object.
function Kjs_socket(url, onreceive){
	var socket;
	var status=0;
	var error = true;
	var errorMessage="";
	
	if ((!url) || (!onreceive)){
		throw "Kjs_socket: parameters are obligatory.";
	}

	this.getStatus=function(){
		return status;
	}
	this.getError=function(){
		return error;
	}
	this.getErrorMessage=function(){
		return errorMessage;
	}


	try{
		socket = new WebSocket(url);
	}catch(e){
		status=-1;
		error=true;
		errorMessage="Kjs_socket: exception of new WebSocket =="+e;
		onreceive({status:status, msg:errorMesssage, inobj:null});
	};




	socket.addEventListener('open', function (event) {
		error=false;errorMessage="Kjs_socket: socket was opened";
		status=1;
		onreceive({status:status, msg:errorMessage, inobj:null});
	}
	);

	//alert('Before socket.addEventListener("error",');
	socket.addEventListener('error', function (event) {
	//alert('Into socket.addEventListener("error",...event='+JSON.stringify(event));
		status=-2;
		error=true;
		errorMessage="Kjs_socket: socket err event == "+JSON.stringify(event);
		onreceive({status:status, msg:errorMessage, inobj:null});
	}
	);

	socket.addEventListener('close', function (event) {
		status=-3;
		error=true;
		errorMessage="Kjs_socket: socket closed =="+event;
		onreceive({status:status, msg:errorMessage, inobj:null});
	});




	socket.addEventListener('message', function (event) {
		var cR;//The result of checking the congratulation answer; see "checkCongrAnswer" function
			//var recObj ={status:0, msg:"", inObj:null}
		if (status===1){
			cR=checkInCome(event.data);
			if (cR.err){
				status=-4;
				error=true;
				errorMessage=cR.errMess;
				onreceive({status:status, msg:errorMesssage, inobj:null});
			}else{
				status=1; error=false;errorMessage="";
				onreceive({status:status, msg:errorMesssage, inobj:cR.data});
			}
		};
		error=true;errorMessage="Kjs_socket onmessage: not allowed status="+status;
		status=-5;
		onreceive({status:status, msg:errorMesssage, inobj:null});
	});	
	
	//Method send returns an empty string if success and a non-empty string if fault
	this.send=function(outcomeObj){
		var cR=checkOutcome(outcomeObj);
		if (cR.err) {return " Kjs_socket.send error = " + cR.errMess}
		try{socket.send(cR.data)}catch(e){return " Kjs_socket.send error = "+e}
	};
	
	this.close = function(){
		if (status==1) {socket.close()};
	};
	return {status:status, msg:errorMessage, inobj:null}
}
