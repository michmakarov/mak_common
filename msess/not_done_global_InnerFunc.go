// GlobalNotDoneInnerFunc
//All functions  here have first-hand access to GlobalNotDone.notDone.
//So they may be invoked only by GlobalNotDone.nd_Monitor function caused by reading from a channel with the same name.
package msess

import (
	"fmt"
)

func GlobalNotDoneInnerFuncHello() {
	fmt.Println("Hello World! From GlobalNotDoneInnerFuncHello()")
}

func (gnd *GlobalNotDone) cancelAllForAgent(data interface{}) (res nd_MonitorResult) {
	var chr *Chore
	var ok bool
	var a *Agent
	if a, ok = data.(*Agent); !ok {
		res.Err = fmt.Errorf("cancelAllForAgent: Given data for registration is not converted to *Agent")
		return
	}

	for e := gnd.notDone.Front(); e != nil; e = e.Next() {
		if chr, ok = e.Value.(*Chore); !ok {
			panic("*GlobalNotDone.cancelAll:an element of the list does not represent *Chore")
		}
		if chr.ulr.tag == a.Tag {
			chr.cancel()
		}
	}
	return
}

func (gnd *GlobalNotDone) addChore(data interface{}) (res nd_MonitorResult) {
	var chr *Chore
	var ok bool

	if chr, ok = data.(*Chore); !ok {
		res.Err = fmt.Errorf("GlobalNotDone.addChore: Given data for registration is not converted to *Chore")
		return
	}

	gnd.notDone.PushBack(chr)
	gnd.count++
	return
}
