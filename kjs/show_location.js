"use strict";
//This was supposed (190120 11:24) as thing showing a location object of current window or document.
//Let us be looking how that destination formula will change further

//What do I want to do at this begin moment?
//It is a function that will report properties of js object "location"
//It take the "id" of some DOM element in which the report will be placed.
//

function showLocation_MachMak(id){
	var makeRep=function(){
		return "Click this to hide<br>"
		+ "window.location.href=" + window.location.href+"<br>"
		+ "window.location.hostName="+ window.location.hostname+"<br>"
		+ "window.location.pathname="+ window.location.pathname+"<br>"
		+ "window.location.port="+ window.location.port+"<br>"
		+ "window.location.protocol="+ window.location.protocol;
	}	

	var showLocEl = document.getElementById(id);
	//var qqq = function(){showLocation_MachMak(id);return;};
	if (!showLocEl) {
		throw "showLocation ERROR! there is not element to place a report";
	}
	if (showLocEl.tagName!="P") {
		//alert(showLocEl.tagName);
		throw "showLocation ERROR! The tag of report element must be '<p>'";
	}
	if (typeof showLocEl.showElStatus === 'undefined'){
		showLocEl.showElStatus=0;
		showLocEl.innerHTML="It can show to you window.location properties. Click it to show ...";
		//alert(showLocEl.showElStatus);
		//showLocEl.onClick="showLocation_MachMak("+id+")";
		//alert(showLocEl.onClick);
		showLocEl.addEventListener("click", function(){showLocation_MachMak(id);return;}, false);
		return
	}
	switch(showLocEl.showElStatus){
		case 0:
		showLocEl.showElStatus=1;
		//showLocEl.innerHTML="I am glad but do not know how to show location properties.<br> Click me to get back";
		showLocEl.innerHTML=makeRep();
		break
		case 1:
		showLocEl.showElStatus=0;
		showLocEl.innerHTML="I can show to you window.location properties. Click me to get ...";
		break
		default:
		throw "showLocation ERROR! Illegal status of report element: "+showLocEl.showElStatus;
	}
}
function www(){return "www";}

function hide_showContent(elementId, contentForHidenstate){
	var elem=document.getElementById(elementId);
	
	if (!contentForHidenstate){contentForHidenstate="Показать ..."}
	if (!elem.visibilityState){
		elem.visibilityState="notVisible";
		elem.oldContent=elem.innerHTML
		elem.innerHTML=contentForHidenstate;
	}else{
		elem.visibilityState=undefined;
		elem.innerHTML=elem.oldContent;
	}
} 















