// GlobalNotDoneInnerFunc
//All functions  here have first-hand access to GlobalNotDone.notDone.
//So they may be invoked only by GlobalNotDone.Run function caused by reading from a channel with the same name.
package msess

import (
	"fmt"
)

func GlobalNotDoneInnerFuncHello() {
	fmt.Println("Hello World! From GlobalNotDoneInnerFuncHello()")
}

func (gnd *GlobalNotDone) cancelAll(a *Agent) {
	var chr *Chore
	var ok bool
	for e := gnd.notDone.Front(); e != nil; e = e.Next() {
		if chr, ok = e.Value.(*Chore); !ok {
			panic("*GlobalNotDone.cancelAll:an element of the list does not represent *Chore")
		}
		if chr.ulr.tag == a.Tag {
			chr.cancel()
		}
	}
}
