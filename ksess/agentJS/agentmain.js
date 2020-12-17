//this script creates a Worker object and adjusts the window.onerror handler to send a message to it when fired

var ksessAgentWorker

if (window.Worker){
	//Let it that this script linked to index.html (or simply inserted into it) and the variable controlPassword is previously established
	var ksessAgentWorkerURL ="/get_agent_worker?CONTROL_PASSWORD="+controlPassword;
	try {
		ksessAgentWorker = new Worker(ksessAgentWorkerURL);
		window.onerror = function (message, source, lineno, colno, errror)  {
			var sentMess = message+"; "+source+": "+lineno+", "+ colno
			ksessAgentWorker.postMessage(sentMess)
			return false;
		}
	} catch (e){console.error("Кот отладка не катит : ", e)}
	
}else{
	console.error("Кот отладка не катит : нетути wingow.Worker - во!");
};

function SendToKsessAgentWorker(mess){
	var ksessUser = "Unknow_user";
	if (userID) {ksessUser=userID;};
	if (ksessAgentWorker){
		ksessAgentWorker.postMessage("User="+ksessUser+ "сказал:"+ mess);
	}else{console.error("SendToKsessAgentWorker : нетути ksessAgentWorker - во те раз!")}
}
