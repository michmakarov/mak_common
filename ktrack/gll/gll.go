/*
kgll дает функциональность для получения полезных данных (данных прикладного уровня)
от трекера Galileosky согласно "Протоколу обмена с сервером терминалов Galileosky" из
https://7gis.ru/assets/files/docs/manuals_ru/opisanie-protokola-obmena-s-serverom-(487740-v16).pdf
На 190921 это прежде всего получение
"кординатно-временной точки" трекера, то есть данных, которые связывают местоположение трекера со времнем
их определения
*/
package kgll

import (
	"fmt"
	"io"
	"net"
	"time"
	//"kot_common/kerr"
)

var (
	//Ошибка чтения первых трех байт пакета.
	//Устанавливается в !=nil если первая стадия чтения пакета (См. ReadRawFirstPckt) завершилась ошибкой.
	//То есть функция чтения пакета (ReadRawFirstPckt или ReadRawBasePckt) устанавливает эту переменную
	//в фактически полученную ошибку первой стадии чтения пакета
	//Таким образом, в месте вызова этих функций из (ReadingThreeByteErr==err) можно определить, на какой стадии чтения пакета произошла ошибка
	ReadingThreeByteErr error

	//firstStageLen - максимальная длительность первой стадии чтения пакета (предельное время ожидания первых трех байт пакета), микросекунды
	//минимельное значение (и значение по умолчанию) 1500 миллисекунд
	firstStageLen = time.Millisecond * 3000
	//secondStageLen - максимальная длительность второй стадии чтения пакета (предельное время ожидания остатка пакета), микросекунды
	//минимельное значение (и значение по умолчанию) 500 миллисекунд
	secondStageLen = time.Millisecond * 2000

	//Функции чтения пакета (ReadRawFirstPckt и ReadRawBasePckt) при возникновении ошибки могут возвратить (на второй стадии чтения)  не пустой массив считенных байт, но неболее этого значения
	limitReading = 66000
)

//SetFirstStageLen позволяет увеличить время ожидания первых трех байт пакета (свыше 3000 миллисекунд)
func SetFirstStageLen(d time.Duration) {
	if d < time.Millisecond*3000 {
		d = time.Millisecond * 3000
	}
	firstStageLen = d
}

//SetSecondStageLen позволяет увеличить время ожидания остатка пакета (свыше 2000 миллисекунд)
func SetSecondStageLen(d time.Duration) {
	if d < time.Millisecond*2000 {
		d = time.Millisecond * 2000
	}
	secondStageLen = d
}

//emptyConn опустошает канал.
//То есть, считывает байты в цикле, пока очередная операция чтения ничего не прочитает или не будет ошиики в соединении
//Если суммарное число считаных байт превысилю limitReading, функция паникует.
//Перед началом работы устанавливает deadline в бесконечность
func emptyConn(conn net.Conn) (read []byte) {
	var n int
	var err error
	var b []byte
	var sum int

	conn.SetDeadline(time.Time{})

	b = make([]byte, 1)
	for {
		n, err = conn.Read(b)
		if (n == 0) || (err != nil) {
			return
		} else {
			read = append(read, b[0])
		}
		sum = sum + n
		if sum > limitReading {
			panic("kgll.EmptyConn: in TCP channel there is too many bytes")
		}
	}
	return
}

//ReadRawFirstPckt считывает из соединения "сырой (rawPckt []byte)" входящий пакет в предположении, что это первый пакет
//после установления соединения, а также его контрольную сумму (cs uint16).
// В общем, функция не интересуется содержанием считаных байт, за двумя
//исключениями: (1) проверяет, что первый байт несет 1; (2) проверяет, что контрольная сумма пакета за исключеним
//последних двух байт равна значению uint16 в этих последних.
//То есть, если нет ошибки, то первый пакет гарантировано принят.
//Чтение происходит в два этапа:
//1.Считываются первые три байта в блокирующем режиме. То есть функция может ожидать поступления трех байт бесконечно
//Если ожидание завершается ошибкой, то возвращается nil и  err =  ReadingRawPcktFatalErr (фатальная ошибка). Это
//единственный случай ошибки (равно для первого и второго этапа), когда rawFirstPckt==nil. В остальных случаях возвлащается три байта плюс то
//что удалось прочитать из соединения на момент обнаружения ощибки. Например,
// первый байт не 1, это ошибка - пакет не отвечает требованиям к первому пакету. Но все равно возвращаются первые
// три байта и все то, нашлось в соединении на момент обнаружения факта ошибки.
//
//2.Считывается остаток пакета, длина которого определяется из трех байт. считаных на первом этапе
//В начале этапа для соединения устанавливается deadline. См. также функцию SetIdleTimeout.
//То есть, ожидание оставшихся байт может продолжаться не далее deadline
//
//Перед возвратом, посредством defer, deadline соединения устанавливается в бесконечность
func ReadRawFirstPckt(conn net.Conn) (rawPckt []byte, cs []byte, err error) {
	var dataLen uint16 // Длина данных. То есть число байт между третьим и предпоследним

	// read - уже считанные байты на момент вызова функции (может быть nil)
	//Далее функция полагает rawPckt = read, считывает остаток rest и присоединяет его к rawPckt
	var addRestBytes = func(read []byte) {
		var rest []byte
		conn.SetDeadline(time.Time{})
		rawPckt = read
		rest = emptyConn(conn)
		for i := 0; i < len(rest); i++ {
			rawPckt = append(rawPckt, rest[i])
		}
	}

	defer conn.SetDeadline(time.Time{})

	{ //1.Перва стадия: чтение первых трех байт
		var rp []byte      //raw packet buffer
		var copyLen []byte //копия 2 и 3 байта для определения длины данных

		rp = make([]byte, 3)

		conn.SetDeadline(time.Now().Add(firstStageLen))

		if _, err = io.ReadFull(conn, rp); err != nil { //Единственный случай, когда функция возвращает rawPckt==nil - из соединения не пришло трех байт по исчерпанию времени олидания, или ошибки сети, или ...
			err = fmt.Errorf("kgll.ReadRawFirstPckt:(deadline = %v)Ошибка чтения первых трех байт =%v", firstStageLen, err.Error())
			ReadingThreeByteErr = err
			rawPckt = nil
			return
		}

		//kerr.PrintDebugMsg(false, "stage1", fmt.Sprintf("F1;rp(3 байта)=%v", rp))

		//Проверка значения первого байта
		if rp[0] != 1 {
			err = fmt.Errorf("kgll.ReadFirstPckt:В первом пакете первый байт должен быть 1 но не %v", rp[0])
			addRestBytes(rp)
			return
		}

		//Копирование второго и третьего байта
		copyLen = make([]byte, 2)
		//lenBuff = bytes.NewBuffer(rp[1:3])
		copyLen[0] = rp[1]
		copyLen[1] = rp[2]

		//установка в копии бита признака наличия данных в архиве в 0
		copyLen[1] = copyLen[1] & 127

		//Определение длины данных
		dataLen = uint16(copyLen[0]) + uint16(copyLen[1])*256

		rawPckt = rp
	} //first stage end

	{ //Second stage
		var restLen uint16 //длина остатка пакета
		var rest []byte    //остаток пакета
		var calcCS uint16  //Вычесленная контрольная сумма

		//Считывание остатка пакета
		restLen = dataLen + uint16(2) //минимальное значение 2, когда длина данных равняется 0, то есть осталось считать только контрольную сумму
		rest = make([]byte, restLen)

		//kerr.PrintDebugMsg(false, "stage1", fmt.Sprintf("F2;restLen=%v", restLen))

		conn.SetDeadline(time.Now().Add(secondStageLen))
		if _, err = io.ReadFull(conn, rest); err != nil {
			err = fmt.Errorf("kgll.ReadFirstPckt:Ошибка чтения остатка пакета %v", err.Error())
			addRestBytes(rawPckt)
			return
		}

		//Присоединение остатка к считанным на первой стадии трем байтам
		for i := 0; i < len(rest); i++ {
			rawPckt = append(rawPckt, rest[i])
		}

		//kerr.PrintDebugMsg(false, "stage1", fmt.Sprintf("F2;rest=%v", rest))

		//Вычисление контрольной суммы пакета
		calcCS = CRC16(rawPckt[0 : len(rawPckt)-2]) //То есть пакета за исключением последних двух байт

		//Выделение из остатка байт, несущих контрольную сумму, то есть последних двух байт
		cs = rest[len(rest)-2 : len(rest)]

		//kerr.PrintDebugMsg(false, "crc", fmt.Sprintf("calcCS=%x; cs=%v", calcCS, cs))

		//Сравнение контрольных сумм
		if calcCS != (uint16(cs[0]) + uint16(cs[1])*256) {
			err = fmt.Errorf("kgll.ReadFirstPckt:Вычисленная контрольная сумма не равна переданой")
			addRestBytes(rawPckt)
			return
		}
	} //second stage end

	return
}

//ReadRawBasePckt считывает из соединения "сырой (rawPckt []byte)" входящий пакет в предположении, что это "основной" пакет
//То есть не первый
//Эта функция повторяет ReadRawFirstPckt за исключением текстов сообщений об ошибках
//The repeat contradicts the main rule of programming : do not repeat code
//But there is the super main rule : there are not obligatory rules. Is not it so?
func ReadRawBasePckt(conn net.Conn) (rawPckt []byte, cs []byte, err error) {
	var dataLen uint16 // Длина данных. То есть число байт между третьим и предпоследним

	// read - уже считанные байты на момент вызова функции (может быть nil)
	//Далее функция полагает rawPckt = read, считывает остаток rest и присоединяет его к rawPckt
	var addRestBytes = func(read []byte) {
		var rest []byte
		conn.SetDeadline(time.Time{})
		rawPckt = read
		rest = emptyConn(conn)
		for i := 0; i < len(rest); i++ {
			rawPckt = append(rawPckt, rest[i])
		}
	}

	defer conn.SetDeadline(time.Time{})

	{ //1.Перва стадия: чтение первых трех байт
		var rp []byte      //raw packet buffer
		var copyLen []byte //копия 2 и 3 байта для определения длины данных

		rp = make([]byte, 3)

		conn.SetDeadline(time.Now().Add(firstStageLen))
		if _, err = io.ReadFull(conn, rp); err != nil { //Единственный случай, когда функция возвращает rawPckt==nil - из соединения не пришло трех байт по исчерпанию времени олидания, или ошибки сети, или ...
			err = fmt.Errorf("kgll.ReadRawBasePckt:(deadline = %v)Ошибка чтения первых трех байт =%v", firstStageLen, err.Error())
			ReadingThreeByteErr = err
			rawPckt = nil
			return
		}

		//kerr.PrintDebugMsg(false, "stage1", fmt.Sprintf("F1;rp(3 байта)=%v", rp))

		//Проверка значения первого байта
		if rp[0] != 1 {
			err = fmt.Errorf("kgll.ReadRawBasePckt:В первом пакете первый байт должен быть 1 но не %v", rp[0])
			addRestBytes(rp)
			return
		}

		//Копирование второго и третьего байта
		copyLen = make([]byte, 2)
		//lenBuff = bytes.NewBuffer(rp[1:3])
		copyLen[0] = rp[1]
		copyLen[1] = rp[2]

		//установка в копии бита признака наличия данных в архиве в 0
		copyLen[1] = copyLen[1] & 127

		//Определение длины данных
		dataLen = uint16(copyLen[0]) + uint16(copyLen[1])*256

		rawPckt = rp
	} //first stage end

	{ //Second stage
		var restLen uint16 //длина остатка пакета
		var rest []byte    //остаток пакета
		var calcCS uint16  //Вычесленная контрольная сумма

		//Считывание остатка пакета
		restLen = dataLen + uint16(2) //минимальное значение 2, когда длина данных равняется 0, то есть осталось считать только контрольную сумму
		rest = make([]byte, restLen)

		//kerr.PrintDebugMsg(false, "stage1", fmt.Sprintf("F2;restLen=%v", restLen))

		conn.SetDeadline(time.Now().Add(secondStageLen))
		if _, err = io.ReadFull(conn, rest); err != nil {
			err = fmt.Errorf("kgll.ReadRawBasePckt:Ошибка чтения остатка пакета %v", err.Error())
			addRestBytes(rawPckt)
			return
		}

		//Присоединение остатка к считанным на первой стадии трем байтам
		for i := 0; i < len(rest); i++ {
			rawPckt = append(rawPckt, rest[i])
		}

		//kerr.PrintDebugMsg(false, "stage1", fmt.Sprintf("F2;rest=%v", rest))

		//Вычисление контрольной суммы пакета
		calcCS = CRC16(rawPckt[0 : len(rawPckt)-2]) //То есть пакета за исключением последних двух байт

		//Выделение из остатка байт, несущих контрольную сумму, то есть последних двух байт
		cs = rest[len(rest)-2 : len(rest)]

		//kerr.PrintDebugMsg(false, "crc", fmt.Sprintf("calcCS=%x; cs=%v", calcCS, cs))

		//Сравнение контрольных сумм
		if calcCS != (uint16(cs[0]) + uint16(cs[1])*256) {
			err = fmt.Errorf("kgll.ReadRawBasePckt:Вычисленная контрольная сумма не равна переданой")
			addRestBytes(rawPckt)
			return
		}
	} //second stage end

	return
}

//SendConfirm формирует и посылает подтверждающий пакет
//cs - контрольная сумма того пакета, для которого шлется подтверждение
//Перед отправкой устанавливает deadline в secondStageLen
//Перед возвратом (через defer)  устанавливает deadline в бесконечносить
func SendConfirm(conn net.Conn, cs []byte) (cnfpack []byte, err error) {
	var n int
	var answer []byte
	var sent int //Число отправленых байт

	defer conn.SetDeadline(time.Time{})
	conn.SetDeadline(time.Now().Add(secondStageLen))

	//Таблица1 "Структура пакета подтверждения приема"
	answer = make([]byte, 3)
	answer[0] = 2
	//answer[1] = byte(cs % 256)
	//answer[2] = byte((cs - (cs % 256)) / 256)
	answer[1] = cs[0]
	answer[2] = cs[1]

	for {
		if n, err = conn.Write(answer); err != nil {
			return
		}
		sent = sent + n
		if sent < 3 {
			answer = answer[sent:3]
		} else {
			break
		}
		time.Sleep(time.Millisecond * 50)

	}
	cnfpack = answer
	return

}

//IsPacksEqual возвращает true только если p1 и p2 не nil, имеют одинаковую длину и
//одинаковые по значению элементы
func IsPacksEqual(p1, p2 []byte) bool {
	if (p1 == nil) || (p2 == nil) {
		return false
	}
	if len(p1) != len(p2) {
		return false
	}
	for i := 0; i < len(p1); i++ {
		if p1[i] != p2[i] {
			return false
		}
	}
	return true
}

//copyBytes return a copy of sl
//That is it allocates memory for a retuning slice and copyes content of sl into there
func CopyBytes(sl []byte) []byte {
	var cp []byte
	if len(sl) == 0 {
		return nil
	}
	cp = make([]byte, len(sl))
	return cp
}
