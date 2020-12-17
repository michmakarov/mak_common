// crc
package kgll

//	"bufio"
//"fmt"
//"kot_common/kerr"
//	"os"

func getLoByte(crc uint16) (loB byte) {
	return byte(crc % 0xFF)
}

//CRC16 вычисляет сонтрольную сумму согласно MODBUS over serial line specification and implementation guide V1.02; 6.2 Appendix B - LRC/CRC Generation; 6.2.2 CRC Generation
func CRC16(msg []byte) (crc uint16) {
	var q uint16 = 0x0001  // 0000000000000001
	var pv uint16 = 0xA001 //1010 0000 0000 0001
	var LSB, n int

	//1. Load a 16–bit register with FFFF hex (all 1’s). Call this the CRC register.
	crc = 0xFFFF
	for i := 0; i < len(msg); i++ {
		n = 0
		//2. Exclusive OR the first 8–bit byte of the message with the low–order byte of the 16–bit CRC register, putting the result in theCRC register.
		crc = crc ^ uint16(msg[i])
		//kerr.PrintDebugMsg(false, "crc", fmt.Sprintf("XOR %v BYTE (n=%v)=%016b; ", i, n, crc))

		for {
			//5. Repeat Steps 3 and 4 until 8 shifts have been performed. When this is done, a complete 8–bit byte will have been processed.
			if n > 7 {
				break
			}

			//3. Shift the CRC register one bit to the right (toward the LSB), zero–filling the MSB. Extract and examine the LSB.
			LSB = int(crc & q)
			crc = crc >> 1
			n++
			//kerr.PrintDebugMsg(false, "crc", fmt.Sprintf("MOVE %v byte (n=%v)=%016b|%v ", i, n, crc, LSB))
			//4. (If the LSB was 0): Repeat Step 3 (another shift).
			//(If the LSB was 1): Exclusive OR the CRC register with the polynomial value 0xA001 (1010 0000 0000 0001).
			if LSB == 0 {
				//crc = crc >> 1
				//LSB = int(crc & q)
				//n++
				//kerr.PrintDebugMsg(false, "crc", fmt.Sprintf("MOVE %v byte (n=%v)=%016b|%v ", i, n, crc, LSB))
			} else {
				crc = crc ^ pv
				//kerr.PrintDebugMsg(false, "crc", fmt.Sprintf("XOR PV (byte %v)(n %v)=%016b;", i, n, crc))
			}

		}

	}
	return

}
