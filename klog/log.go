//klog
//What for is it if there is Golang package Log
//That  is some reasons, for example the next
//Once Ilnur said that the console on the working server is not accessible for him
//It appears to be true for me too
//but it so wants by a one movement to print a message to the console and to a log file
//at that where it is need
//And yet - I have noted that sometimes to understand existing thing in its deeps more difficult than to create a new one
//In any case to create is more merry than to understand
package klog

import (
	"fmt"
	"os"
)

type Klogger struct {
	working bool //if true the message does not print to the console
	f       *os.File
}

func NewKlog(fFulName string, mode bool /*working or not*/) (*Klogger, error) {
	var (
		f   *os.File
		err error
	)
	if f, err = os.Create(fFulName); err != nil {
		return nil, err
	}
	return &Klogger{mode, f}, nil
}

func (kl *Klogger) Printf(format string, args ...interface{}) {
	var mess = fmt.Sprintf(format+"\n", args...)
	kl.f.Write([]byte(mess))
	if !kl.working {
		fmt.Println(mess)
	}
}

func (kl *Klogger) SetMode(working bool) {
	kl.working = working
}
