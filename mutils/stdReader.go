//StdRd - 180204 ver0_4
package mutils

import (
	"fmt"
	"os"
)

//The method run of this object if it be run in goroutine would be send a read byte to CmdByte

func CreateStdRd(showDir bool, prompt string) chan []byte {
	var stdRdChan = make(chan []byte)
	go run(showDir, prompt, stdRdChan)
	return stdRdChan
}

func run(showDir bool, prompt string, stdRdChan chan []byte) {

	var b []byte
	var dir string
	var err error
	var n int
	//var ok bool

	if prompt == "" {
		prompt = ">"
	}
	if showDir {
		if dir, err = os.Getwd(); err != nil {
			dir = "?"
		}
	}

	b = make([]byte, 256)
	//fmt.Printf("%v%v", dir, prompt)
	for {
		for i := 0; i < 256; i++ {
			b[i] = 0
		}
		if n, err = os.Stdin.Read(b); err != nil {
			b[0] = 0
		}
		if b[0] != 0 {
			//fmt.Printf("%v%v%v\n", dir, prompt, b[0])
			b = b[:n]
			fmt.Printf("%v%v", dir, prompt)
		} else {
			fmt.Printf("%v%v%v\n", dir, prompt, "error")
			fmt.Printf("%v%v", dir, prompt)
		}
		stdRdChan <- b
	}
}
