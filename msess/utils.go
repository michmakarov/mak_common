// utils
//210604 17:36 Here are utility that do not have special demands for their using
//That is them go not have any side effects and supposing about environment of using
//Of course, each of them was created for special case but may be using anywhere
package msess

import (
	"fmt"
	"os"
	"strings"
)

//210101 for func (fl *feelerLogger) getFlrlogMess
func byteSet(value byte, bitNum int) bool {
	var mask byte
	if bitNum < 1 || bitNum > 7 {
		panic(fmt.Sprintf("byteSet: illegal bit number=%v", bitNum))
	}
	switch bitNum {
	case 1:
		mask = 0b00000001
	case 2:
		mask = 0b00000010
	case 3:
		mask = 0b00000100
	case 4:
		mask = 0b00001000
	case 5:
		mask = 0b00010000
	case 6:
		mask = 0b00100000
	case 7:
		mask = 0b01000000
	case 8:
		mask = 0b10000000
	}
	return (value & mask) != 0
}

//210603 07:41 This is an analog of the func byteSet
//That is it answers whether is in the value a one-length substring char
//The length of the value must not be more than 8 , and the length of the char must be equal 1. If not the case the function panics.
func stringSet(value string, char string) bool {
	if len(value) > 8 {
		panic("msess.stringSet: too long value (>8)")
	}
	if len(char) != 1 {
		panic("msess.stringSet: bad char parameter (len != 1")
	}
	if strings.Index(value, char) != -1 {
		return true
	} else {
		return false
	}
}

//210603 09:09
func checkLogDirs() error {
	var err error
	if sessCP.Loggers == "" {
		return nil
	}
	if stringSet(sessCP.Loggers, "h") {
		if _, err = os.Stat("logs/h"); err != nil {
			return fmt.Errorf("Absence of logs/h directory")
		}
	} //h
	if stringSet(sessCP.Loggers, "f") {
		if _, err = os.Stat("logs/f"); err != nil {
			return fmt.Errorf("Absence of logs/f directory")
		}
	} //f
	if stringSet(sessCP.Loggers, "u") {
		if _, err = os.Stat("logs/u"); err != nil {
			return fmt.Errorf("Absence of logs/u directory")
		}
	} //u
	if stringSet(sessCP.Loggers, "g") {
		if _, err = os.Stat("logs/g"); err != nil {
			return fmt.Errorf("Absence of logs/g directory")
		}
	} //g
	return nil
}
