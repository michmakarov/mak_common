// gllstruct
package kgll

//"fmt"

type Pckt struct {
	header    byte
	NSD       bool //There are data not sent
	PacketLen uint16
	Data      []byte
	CS        uint16 //Control sum
}
