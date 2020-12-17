// debug
package kerr

import (
	"fmt"
	"runtime"
	"sync"
)

type DebugFlag struct {
	Name  string
	Value bool
}

type DebugFlags struct {
	All_switch_off bool
	Flags          []DebugFlag
}

var (
	dFs      = DebugFlags{false, make([]DebugFlag, 0)}
	setMtx   = &sync.Mutex{}
	printMtx = &sync.Mutex{}
	onMtx    = &sync.Mutex{}
	offMtx   = &sync.Mutex{}
)

func findDebugFlag(name string) (p *DebugFlag) {
	for _, v := range dFs.Flags {
		if v.Name == name {
			p = &v
			return
		}
	}
	return
}

func SetDebugFlag(flagName string) {
	setMtx.Lock()
	defer setMtx.Unlock()
	if findDebugFlag(flagName) != nil {
		return
	}
	dFs.Flags = append(dFs.Flags, DebugFlag{flagName, true})
	fmt.Println("--- Was Set debug flag", flagName)
}

func SwitchFlagOn(debugFlagName string) {
	onMtx.Lock()
	defer onMtx.Unlock()
	if df := findDebugFlag(debugFlagName); df != nil {
		df.Value = true
	}
}
func SwitchFlagOff(debugFlagName string) {
	offMtx.Lock()
	defer offMtx.Unlock()
	if df := findDebugFlag(debugFlagName); df != nil {
		df.Value = false
	}
}

func PrintDebugMsg(showLocation bool, debugFlagName string, msg string) {
	var (
		filename string
		line     int
		df       *DebugFlag
	)
	printMtx.Lock()
	defer printMtx.Unlock()
	if dFs.All_switch_off {
		return
	}
	//fmt.Printf("PrintDebugMsg: debugFlagName=%v", debugFlagName)
	df = findDebugFlag(debugFlagName)
	if df == nil {
		return
	}
	//fmt.Printf("PrintDebugMsg: df=%v\n", df)
	if !df.Value {
		return
	}

	_, filename, line, _ = runtime.Caller(1)
	if showLocation {
		fmt.Printf("-!-File %v(%v):%v / %v\n", filename, line, msg, debugFlagName)
	} else {
		fmt.Printf("-!-%v / %v \n", msg, debugFlagName)
	}
}
