package ktrack

import (
	//"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/LdDl/go-egts/crc"
)

var idleTimeout time.Duration = time.Millisecond * 500

//SetIdleTimeout sets the period for reading/writing a full incoming/outcoming packet since the ending last successful operation of reading/writing
// Default value is half of second
func SetIdleTimeout(d time.Duration) {
	idleTimeout = d
}

//EmptyConn опустошает канал. То есть, считывает байты в цикле, пока очередная операция чтения ничего не прочитает
//Если суммарное число считаных байт превысилю 65535, функция паникует.
func EmptyConn(conn net.Conn) (read []byte) {
	var n int
	var err error
	var b []byte
	var sum int

	defer conn.SetDeadline(time.Time{})
	conn.SetDeadline(time.Now().Add(idleTimeout))
	b = make([]byte, 1)
	for {
		n, err = conn.Read(b)
		if (n == 0) || (err != nil) {
			return
		} else {
			read = append(read, b[0])
		}
		sum = sum + n
		if sum > 65535 {
			panic("ktrack.EmptyConn: in TCP channel there is too many bytes")
		}
	}
	return
}

var (
	RawPcktFatalErr error //= fmt.Errorf("Falal error: Ошибка чтения первых 4 байт; rawPack==nil")
)

//GetRawPcktTL читает из соединения "сырой ([]byte)" входящий пакет транспортного уровня
//Чтение происходит в четырt этапа
//1.Из соединения считывается четыре байта, последний из киторых несет длину заголовка
//2.По длине заголовка определяется число не считанных байтов заголовка и они счтаваются
//3. Из заголовка извлекется длина данных (поле заголовка FDL (Frame Data Length))
//и, если оно не равно 0, они считываются.
//В противном случае предполагается, что поля пакета SFRD (Services Frame Data) и SFRCS (Services Frame Data Check Sum) отсутствуют
//4.Считывется SFRCS (Services Frame Data Check Sum) если поле заголовка FDL несет значение >0
//После каждого успешного считывания время ожидания очередных байтов сдвигается на idleTimeout
//При заверщении работы функция опустошает канал и устанавливает бесконечное время ожидания очередных байтов (вызовом EmptyConn)
func GetRawPcktTL(conn net.Conn) (rawPack []byte, err error) {
	var p PcktTL = PcktTL{}

	var addRestBytes = func() {
		var rest []byte
		rest = EmptyConn(conn)
		for i := 0; i < len(rest); i++ {
			rawPack = append(rawPack, rest[i])
		}
	}

	//conn.SetDeadline(time.Now().Add(idleTimeout))
	//defer conn.SetDeadline(time.Time{})
	defer EmptyConn(conn)

	fmt.Printf("GetRawPcktTL HERE! RA=%v\n", conn.RemoteAddr().String())
	{
		//1. Reading four first bytes of a packet header. Third  (fourth) byte is the length of the packet and its value may be 11 or 16
		//Possible values of the byte  from А.5.23 Блок-схема алгоритма обработки пакета данных протокола транспортного уровня при приеме представлена на рисунке А.2 (см. вкладку).
		var buff4 []byte
		buff4 = make([]byte, 4)

		if _, err = io.ReadFull(conn, buff4); err != nil { //190905 It is seemed that it is an only fatal error
			RawPcktFatalErr = fmt.Errorf("ktrack.GetPcktTL;Ошибка чтения первых 4 байт:%v", err.Error())
			err = RawPcktFatalErr
			return
		}
		//fmt.Printf("GetRawPcktTL buff4=%v\n", buff4)

		if len(buff4) != 4 { //190905 This will not occur at all! It is an impossible event.
			RawPcktFatalErr = fmt.Errorf("ktrack.GetPcktTL;первое чтение: !? нет  4 байт")
			err = RawPcktFatalErr
			return
		}
		p._1M_PRV = buff4[0]
		p._2M_SKID = buff4[1]
		p._3M_PRF_RTE_ENA_CMP_PR = buff4[2]
		p._4M_HL = buff4[3]
		for i := 0; i < 4; i++ {
			rawPack = append(rawPack, buff4[i])
		}
		//fmt.Printf("GetRawPcktTL(1) rawPack=%v\n", rawPack)
		conn.SetDeadline(time.Now().Add(idleTimeout))
	}

	if (p._4M_HL != byte(11)) && (p._4M_HL != byte(16)) {
		err = fmt.Errorf("ktrack.GetPcktTL;Bad value of header length:%v", p._4M_HL)
		addRestBytes()
		return
	}

	{
		//2. Reading rest of the header (it ends with HCS - a header check sum)
		var buffRest []byte
		var b *bytes.Buffer
		buffRest = make([]byte, p._4M_HL-4)

		if _, err = io.ReadFull(conn, buffRest); err != nil {
			err = fmt.Errorf("ktrack.GetPcktTL;Ошибка чтения остатка заголовка:%v", err.Error())
			for i := 0; i < len(buffRest); i++ {
				rawPack = append(rawPack, buffRest[i])
			}
			addRestBytes()
			return
		}

		//This is only to obtain p._6_7M_FDL (Uint16)
		b = bytes.NewBuffer(buffRest[1:3])

		if err = binary.Read(b, binary.LittleEndian, &p._6_7M_FDL); err != nil {
			err = fmt.Errorf("ktrack.GetPcktTL;Ошибка binary.Read(to &p.FDL):%v", err.Error())

			for i := 0; i < len(buffRest); i++ {
				rawPack = append(rawPack, buffRest[i])
			}
			addRestBytes()
			return
		}

		for i := 0; i < len(buffRest); i++ {
			rawPack = append(rawPack, buffRest[i])
		}

		conn.SetDeadline(time.Now().Add(idleTimeout))
	}

	//Stage 3 - чтение данных
	if p._6_7M_FDL != 0 { //3. Reading SFRD []byte // (Services Frame Data)
		var buffSFRD []byte
		//reader := bufio.NewReader(conn)
		buffSFRD = make([]byte, p._6_7M_FDL)
		if _, err = io.ReadFull(conn, buffSFRD); err != nil {
			err = fmt.Errorf("ktrack.GetPcktTL;Ошибка чтения Services Frame Data:%v", err.Error())
			for i := 0; i < len(buffSFRD); i++ { //What if there is something there?
				rawPack = append(rawPack, buffSFRD[i])
			}
			addRestBytes()
			return
		}
		for i := 0; i < len(buffSFRD); i++ {
			rawPack = append(rawPack, buffSFRD[i])
		}
		conn.SetDeadline(time.Now().Add(idleTimeout))
	}

	//Stage 4 - чтение контрольной суммы данных
	if p._6_7M_FDL != 0 { //4. Reading SFRCS (Services Frame Data Check Sum) 2 bytes
		var buffSFRCS []byte
		//reader := bufio.NewReader(conn)
		buffSFRCS = make([]byte, 2)
		if _, err = io.ReadFull(conn, buffSFRCS); err != nil {
			err = fmt.Errorf("ktrack.GetPcktTL;Ошибка чтения Services Frame Data:%v", err.Error())
			for i := 0; i < len(buffSFRCS); i++ { //What if there is something there?
				rawPack = append(rawPack, buffSFRCS[i])
			}
			addRestBytes()
			return
		}
		for i := 0; i < len(buffSFRCS); i++ {
			rawPack = append(rawPack, buffSFRCS[i])
		}
		conn.SetDeadline(time.Now().Add(idleTimeout))
	}

	return rawPack, nil
}

//GetRawPcktTL_Stupid reads from connection all that there is, but panics if there is more than 65535
//It subsequently reads the connection while gives no bytes.
//First reading it does "as is". That is it does not set any deadline.
//For next reading idleTimeout is set (see SetIdleTimeout, default half of second)
//Before returning it sets the deadline to infinite.
func GetRawPcktTL_Stupid(conn net.Conn) (rawPack []byte, err error) {
	var n int
	var b []byte
	var sum int

	//defer conn.SetDeadline(time.Time{})
	b = make([]byte, 1)
	for {
		if n, err = conn.Read(b); err != nil {
			err = fmt.Errorf("ktrack.GetRawPcktTL_Stupid caught err =%v", err.Error())
			return
		}
		if n == 0 {
			return
		}
		rawPack = append(rawPack, b[0])

		sum = sum + n
		if sum > 65535 {
			panic("GetRawPcktTL_Stupid: in TCP channel there is too many bytes (more than 65535")
		}
		conn.SetDeadline(time.Now().Add(idleTimeout))
	}
	//conn.SetDeadline(time.Time{})
	return
}

//SendAnswer writes to connection
//It panics if len(answer)>65535
func SendAnswer(conn net.Conn, answer []byte) (err error) {
	var n int
	var sent int

	if len(answer) > 65535 {
		panic("An answer must not be more than 65535")
	}

	defer conn.SetDeadline(time.Time{})
	conn.SetDeadline(time.Now().Add(idleTimeout))

	for {
		if n, err = conn.Write(answer); err != nil {
			return
		}
		sent = sent + n
		if sent < len(answer) {
			continue
		} else {
			break
		}
	}
	return

}

func Table_A_14(code int) string {
	switch code {
	case 0:
		{
			return "(0  )EGTS_PC_OK Успешно обработано"
		}
	case 1:
		{
			return "(1  )EGTS_PC_IN_PROGRESS	В процессе обработки"
		}
	case 128:
		{
			return "(128)EGTS_PC_UNS_PROTOCOL Неподдерживаемый протокол"
		}
	case 129:
		{
			return "(129)EGTS_PC_DECRYPT_ERROR Ошибка декодирования"
		}
	case 130:
		{
			return "(130)EGTS_PC_PROC_DENIED	Обработка запрещена"
		}
	case 131:
		{
			return "(131)EGTS_PC_INC_HEADERFORM Неверный формат заголовка"
		}
	case 132:
		{
			return "(132)EGTS_PC_INC_DATAFORM Неверный формат данных"
		}
	case 133:
		{
			return "(133)EGTS_PC_UNS_TYPE Неподдерживаемый тип"
		}
	case 134:
		{
			return "(134)EGTS_PC_NOTEN_PARAMS Неверное число параметров"
		}
	case 135:
		{
			return "(135)EGTS_PC_DBL_PROC Попытка повторной обработки"
		}
	case 136:
		{
			return "(136)EGTS_PC_PROC_SRC_DENIED Обработка данных от источника запрещена"
		}
	case 137:
		{
			return "(137)EGTS_PC_HEADERCRC_ERROR Ошибка контрольной суммы заголовка"
		}
	case 138:
		{
			return "(138)EGTS_PC_DATACRC_ERROR Ошибка контрольной суммы данных"
		}
	case 139:
		{
			return "(139)EGTS_PC_INVDATALEN Некорректная длина данных"
		}
	case 140:
		{
			return "(140)EGTS_PC_ROUTE_NFOUND Маршрут не найден"
		}
	case 141:
		{
			return "(141)EGTS_PC_ROUTE_CLOSED Маршрут закрыт"
		}
	case 142:
		{
			return "(142)EGTS_PC_ROUTE_DENIED Маршрутизация запрещена"
		}
	case 143:
		{
			return "(143)EGTS_PC_INVADDR Неверный адрес"
		}
	case 144:
		{
			return "(144)EGTS_PC_TTLEXPIRED Превышено число ретрансляции данных"
		}
	case 145:
		{
			return "(145)EGTS_PC_NO_ACK Нет подтверждения"
		}
	case 146:
		{
			return "(146)EGTS_PC_OBJ_NFOUND Объект не найден"
		}
	case 147:
		{
			return "(147)EGTS_PC_EVNT_NFOUND	Событие не найдено"
		}
	case 148:
		{
			return "(148)EGTS_PC_SRVC_NFOUND	Сервис не найден"
		}
	case 149:
		{
			return "(149)EGTS_PC_SRVC_DENIED	Сервис запрещен"
		}
	case 150:
		{
			return "(150)EGTS_PC_SRVC_UNKN Неизвестный тип сервиса"
		}
	case 151:
		{
			return "(151)EGTS_PC_AUTH_DENIED	Авторизация запрещена"
		}
	case 152:
		{
			return "(152)EGTS_PC_ALREADY_EXISTS Объект уже существует"
		}
	case 153:
		{
			return "(153)EGTS_PC_ID_NFOUND Идентификатор не найден"
		}
	case 154:
		{
			return "(154)EGTS_PC_INC_DATETIME Неправильная дата и время"
		}
	case 155:
		{
			return "(155)EGTS_PC_IO_ERROR Ошибка ввода/вывода"
		}
	case 156:
		{
			return "(156)EGTS_PC_NO_RES_AVAIL Недостаточно ресурсов"
		}
	case 157:
		{
			return "(157)EGTS_PC_MODULE_FAULT Внутренний сбой модуля"
		}
	case 158:
		{
			return "(158)EGTS_PC_MODULE_PWR_FLT Сбой в работе цепи питания модуля"
		}
	case 159:
		{
			return "(159)EGTS_PC_MODULE_PROC_FLT Сбой в работе микроконтроллера модуля"
		}
	case 160:
		{
			return "(160)EGTS_PC_MODULE_SW_FLT Сбой в работе программы модуля"
		}
	case 161:
		{
			return "(161)EGTS_PC_MODULE_FW_FLT Сбой в работе внутреннего ПО модуля"
		}
	case 162:
		{
			return "E(162)GTS_PC_MODULE_IO_FLT Сбой в работе блока ввода/вывода модуля"
		}
	case 163:
		{
			return "E(163)GTS_PC_MODULE_MEM_FLT Сбой в работе внутренней памяти модуля"
		}
	case 164:
		{
			return "E(164)GTS_PC_TEST_FAILED Тест не пройден	"
		}
	default:
		{
			return fmt.Sprintf("Unknow code:%v", code)
		}
	}
}

func ExtractHeader(rawPckt []byte) []byte {
	if len(rawPckt) < 4 {
		panic("Length < 4")
	}
	//return rawPckt[0:rawPckt[4]]
	switch rawPckt[3] {
	case 11:
		return rawPckt[0:11]
	case 16:
		return rawPckt[0:16]
	default:
		panic("package length != 11 or 16")
	}
}

func ParseRawPckt(rawPckt []byte) (p PcktTL, err error) {
	var fullPcktLength int
	p._1M_PRV = rawPckt[0]
	p._2M_SKID = rawPckt[1]               //             byte //M2// (Security Key ID)
	p._3M_PRF_RTE_ENA_CMP_PR = rawPckt[2] //byte //M3
	//Параметр PRF (7-6 bits) определяет префикс заголовка протокола и содержитзначение 00.
	//Поле RTE (Route) (5 bit) определяет необходимость дальнейшей маршрутизации данного пакета на ...
	//Поле ENA (Encryption Algorithm) (4-3 bits) определяет код алгоритма,используемый для шифрования данных из поля SFRD. Если поле имеет значение 00, то данные в поле SFRD не шифруют.
	//Поле CMP (Compressed) (2 bit) определяет, используется ли сжатие данных из поля SFRD. Если поле имеет значение 1, то данные в поле SFRD считаются сжатыми.
	//Поле PR (Priority) (1-0 bits) определяет приоритет маршрутизации данного пакета и может принимать следующие значения: - 00 - наивысший; - 01 - высокий; - 10 - средний; - 11 - низкий.
	p._4M_HL = rawPckt[3]                                     //byte //M4 //(Header Length) А.5.12 Поле HL - длина заголовка протокола в байтах с учетом байта	контрольной суммы (поля HCS).
	p._5M_HE = rawPckt[4]                                     // byte   //M5(Header Encoding) А.5.13 Поле НЕ определяет применяемый метод кодирования, следующей за данным параметром части заголовка протокола.
	p._6_7M_FDL = uint16(rawPckt[5]) + 256*uint16(rawPckt[6]) //uint16 //M6-7//(Frame Data Length)
	p._8_9M_PID = uint16(rawPckt[7]) + 256*uint16(rawPckt[8]) // uint16 //M8-9//(Packet Identifier)//А.5.15
	p._10M_PT = rawPckt[9]                                    // byte   //M10//(Packet Type)
	//А.5.16 Поле РТ - тип пакета транспортного уровня. Поле РТ может принимать следующие значения:
	//- 0 - EGTS_PT_RESPONSE (подтверждение на пакет транспортного уровня);
	//- 1 - EGTS_PT_APPDATA (пакет, содержащий данные протокола уровня поддержки услуг);
	//- 2 - EGTS_PT_SIGNED_APPDATA (пакет, содержащий данные протокола уровня поддержки услуг с цифровой подписью).

	switch p._4M_HL {
	case 11:
		p._11_12O_PRA = -1
		p._13_14O_RCA = -1
		p._15O_TTL = -1
		p._16M_HCS = rawPckt[10]
	case 16:
		p._11_12O_PRA = int(rawPckt[10]) + 256*int(rawPckt[11])
		p._13_14O_RCA = int(rawPckt[12]) + 256*int(rawPckt[13])
		p._15O_TTL = int(rawPckt[14])
		p._16M_HCS = rawPckt[15]
	default:
		panic("ktract.ParseRawPckt: HL packet header length != 11 or 16")
	}
	//PRA int //O11-12// (Peer Address) <0 - it is absence into raw packet
	//А.5.17 Поле PRA - адрес аппаратно-программного комплекса, на котором
	//данный пакет сгенерирован. Данный адрес является уникальным в рамках
	//сети и используется для создания пакета-подтверждения на принимающей стороне.
	//RCA int //O13-14//(Recipient Address)<0 - it is absence into raw packet
	//А.5.18 Поле RCA - адрес аппаратно-программного комплекса, для которого
	//данный пакет предназначен. По данному адресу производят идентификацию
	//принадлежности пакета определенного аппаратно-программного комплекса и
	//его маршрутизация при использовании промежуточных аппаратно-
	//программных комплексов.
	//TTL int //O15// (Time To Live) <0 - it is absence into raw packet
	//А.5.19 Поле TTL - время жизни пакета при его маршрутизации между
	//аппаратно-программными комплексами.
	//HCS byte //M11|16//(Header Check Sum)
	//А.5.20 Поле HCS - контрольная сумма заголовка протокола (начиная с поля
	//"PRV" до поля "HCS", не включая поле "HCS"). Для подсчета значения поля
	//HCS ко всем байтам указанной последовательности применяется алгоритм
	//CRC-8.

	fullPcktLength = int(p._4M_HL) + int(p._6_7M_FDL) + 2
	if fullPcktLength != len(rawPckt) {
		panic("ktract.ParseRawPckt: fullPcktLength!=len(rawPckt)")
	}

	if p._6_7M_FDL == 0 {
		p._17O_SFRD = nil
		p._18O_SFRCS = -1 //absence
	} else {
		if p._4M_HL == 11 {
			p._17O_SFRD = rawPckt[11 : 12+p._6_7M_FDL]
		} else { //==16
			p._17O_SFRD = rawPckt[16 : 17+p._6_7M_FDL]
		}
		p._18O_SFRCS = int(uint16(rawPckt[fullPcktLength-2]) + uint16(rawPckt[fullPcktLength-1])*256)
		fmt.Printf(">>>>>>>>>>>>>>>>>%v!%v!%v\n", fullPcktLength, rawPckt[fullPcktLength-2], rawPckt[fullPcktLength-1])
	}

	//SFRD []byte // (Services Frame Data) nil means absence
	//А.5.21 Поле SFRD - структура данных, зависящая от типа пакета и содержащая информацию протокола уровня поддержки услуг.
	//See type EGTS_PT_RESPONSE,

	//SFRCS int //(Services Frame Data Check Sum) <0 - it is absence into raw packet

	return
}

func getBit(b byte, pos int) int {
	if pos < 0 || pos > 7 {
		panic(fmt.Sprintf("Into a byte there is only eight places, from 0 to 7, but you have given %v", pos))
	}
	return int(b & (1 << uint(pos)))
}

func _3M_PRF_RTE_ENA_CMP_PR_detals(flags byte) string {
	var PRF, RTE, ENA, CMP, PR string

	PRF = fmt.Sprintf("PRF=%v;", getBit(flags, 7)*1+getBit(flags, 6))
	RTE = fmt.Sprintf("RTE=%v;", getBit(flags, 5))
	ENA = fmt.Sprintf("ENA=%v;", getBit(flags, 4)*1+getBit(flags, 3))
	CMP = fmt.Sprintf("CMP=%v;", getBit(flags, 2))
	PRF = fmt.Sprintf("PRF=%v;", getBit(flags, 1)*1+getBit(flags, 0))
	return PRF + RTE + ENA + CMP + PR
}

func (p PcktTL) String(lb string) string {
	var s string
	s = fmt.Sprintf("Protocol Version (1M_PRV):%v%v", p._1M_PRV, lb)
	s = s + fmt.Sprintf("Security Key ID (2M_SKID):%v%v", p._2M_SKID, lb)
	s = s + fmt.Sprintf("Flags (%v):%v%v", p._3M_PRF_RTE_ENA_CMP_PR, _3M_PRF_RTE_ENA_CMP_PR_detals(p._3M_PRF_RTE_ENA_CMP_PR), lb)
	s = s + fmt.Sprintf("Header Length (_4M_HL):%v%v", p._4M_HL, lb)
	s = s + fmt.Sprintf("Header Encoding (_5M_HE):%v%v", p._5M_HE, lb)
	s = s + fmt.Sprintf("Frame Data Length(_6_7M_FDL):%v%v", p._6_7M_FDL, lb)
	s = s + fmt.Sprintf("Packet Identifier(_8_9M_PID):%v%v", p._8_9M_PID, lb)
	s = s + fmt.Sprintf("Packet Type (_10M_PT):%v%v", p._10M_PT, lb)
	s = s + fmt.Sprintf("Peer Address (_11_12O_PRA):%v%v", p._11_12O_PRA, lb)
	s = s + fmt.Sprintf("Recipient Address (_13_14O_RCA):%v%v", p._13_14O_RCA, lb)
	s = s + fmt.Sprintf("Time To Live (_15O_TTL):%v%v", p._15O_TTL, lb)
	s = s + fmt.Sprintf("Header Check Sum (_16M_HCS):%v%v", p._16M_HCS, lb)
	if p._17O_SFRD == nil {
		s = s + fmt.Sprintf("Services Frame Data (_17O_SFRD):%v%v", p._17O_SFRD, lb)
	} else {
		s = s + fmt.Sprintf("Services Frame Data длина (_17O_SFRD):%v%v", len(p._17O_SFRD), lb)
	}
	s = s + fmt.Sprintf("Services Frame Data Check Sum (_18O_SFRCS):%v%v", p._18O_SFRCS, lb)

	return s
}

func (p PcktTL) ShotString(lb string) string {
	var s string
	s = fmt.Sprintf("Protocol Version (1M_PRV):%v%v", p._1M_PRV, lb)
	s = s + fmt.Sprintf("SKID:%v%v", p._2M_SKID, lb)
	s = s + fmt.Sprintf("Flags :%v%v", p._3M_PRF_RTE_ENA_CMP_PR, lb)
	s = s + fmt.Sprintf("HL:%v%v", p._4M_HL, lb)
	s = s + fmt.Sprintf("HE:%v%v", p._5M_HE, lb)
	s = s + fmt.Sprintf("FDL:%v%v", p._6_7M_FDL, lb)
	s = s + fmt.Sprintf("PID:%v%v", p._8_9M_PID, lb)
	s = s + fmt.Sprintf("PT:%v%v", p._10M_PT, lb)
	s = s + fmt.Sprintf("PRA:%v%v", p._11_12O_PRA, lb)
	s = s + fmt.Sprintf("RCA:%v%v", p._13_14O_RCA, lb)
	s = s + fmt.Sprintf("TTL:%v%v", p._15O_TTL, lb)
	s = s + fmt.Sprintf("HCS:%v%v", p._16M_HCS, lb)
	if p._17O_SFRD == nil {
		s = s + fmt.Sprintf("SFRD:%v%v", p._17O_SFRD, lb)
	} else {
		s = s + fmt.Sprintf("len(SFRD):%v%v", len(p._17O_SFRD), lb)
	}
	s = s + fmt.Sprintf("SFRCS:%v%v", p._18O_SFRCS, lb)

	return s
}

//Get_PID возвращает (копирует) PID - иденификатор пакета трансплортного уровня //M8-9//(Packet Identifier)//А.5.15
//Если CheckStructure, то предварительно проверяет корректность сырого пакета вызовом ParseRawPckt
func Get_PID(rawPckt []byte, CheckStructure bool) (PID uint16, err error) {
	if CheckStructure {
		if _, err = ParseRawPckt(rawPckt); err != nil {
			return
		}
	}
	PID = uint16(rawPckt[8]) + uint16(rawPckt[9])*256
	return
}

func make_ok_empty_EGTS_PT_RESPONSE(RPID uint16) []byte {
	var buf bytes.Buffer
	var PR byte
	binary.Write(&buf, binary.LittleEndian, RPID)
	return append(buf.Bytes(), PR)
}

func MakeAnswer11(PID uint16, to_PID uint16) (rawPckt []byte) {
	var SFRD = make_ok_empty_EGTS_PT_RESPONSE(to_PID)
	var PID_buf bytes.Buffer
	var FDL_buf bytes.Buffer
	var HCS_val int //Header check sum value

	var SFRCS_val int //Data check sum value
	var SFRCS_buf bytes.Buffer

	binary.Write(&PID_buf, binary.LittleEndian, PID)
	binary.Write(&FDL_buf, binary.LittleEndian, uint16(len(SFRD)))
	binary.Write(&SFRCS_buf, binary.LittleEndian, len(SFRD))

	rawPckt = make([]byte, 11) //Header with HCS, but without PRA, RCA, TTL (these are absent)

	rawPckt[0] = 1  //PRV
	rawPckt[1] = 0  //SKID
	rawPckt[2] = 2  //flags PRF_RTE_ENA_CMP_PR
	rawPckt[3] = 11 //HL
	rawPckt[4] = 0  //HE
	rawPckt[5] = FDL_buf.Bytes()[0]
	rawPckt[6] = FDL_buf.Bytes()[1]
	rawPckt[7] = PID_buf.Bytes()[0]
	rawPckt[8] = PID_buf.Bytes()[1]
	rawPckt[9] = 0 //PT
	HCS_val = crc.Crc(8, rawPckt[0:10])
	rawPckt[9] = byte(HCS_val)

	for i := 0; i < len(SFRD); i++ {
		rawPckt = append(rawPckt, SFRD[i])
	}

	SFRCS_val = crc.Crc(16, SFRD)
	binary.Write(&SFRCS_buf, binary.LittleEndian, uint16(SFRCS_val))
	for i := 0; i < 2; i++ {
		rawPckt = append(rawPckt, SFRCS_buf.Bytes()[i])
	}

	return
}
