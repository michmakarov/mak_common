/* tags
"Сырой" пакет - это []byte, считанный из соединения с трекером процедурвми ReadRawFirstPckt или ReadRawBasePckt
"Сырой" тэг  - это значение RawTag;
Список "сырых" тегов - это []RawTag, то есть это прикладные данные, которые несет "сырой" пакет
Предполагается, что "сырые" тэги в списке идут не как попало, но в смысловой последовательности.
Например, если в списке встретился тэг 0х20 (32, время),
то он указывает время определения значений  последующих тегов до появления нового 0х20
или до конца списка.
Исходя из этого предположения выделяются пары тегов 0х20(время) и 0х30 (координаты) (см type Pair_32_48),
которые дают данные для "координатно - временной точки"
*/
package kgll

import (
	"encoding/binary"
	"fmt"

	//"kot_common/kerr"

	//"time"
	"strconv"
)

//"Сырой" тэг
type RawTag struct {
	TagNum int    // Номер тэга
	Descr  string // его описание
	Data   []byte // "сырые" данные, как они были считаны из соединения
}

//Типы Tag1, Tag2 ... представляют tags of Galileosky ptotocol как структуры Golang
//То есть, это отпарсерные RawTag.Data функциями ParseTag1, ParseTag2 ...
type (
	Tag1 struct {
		HardVersion byte //Версия железа
	}
	Tag2 struct {
		FirmVersion byte //Версия прошивки
	}
	Tag3 struct {
		IMEI string //The hell knows what is it
	}
	Tag4 struct {
		DevID uint16 //Идентификатор терминара(железки), задается в настройках железки
	}

	Tag16 struct {
		ArchiveRecID uint16 //"Номер записи в архиве"
	}

	Tag32 struct {
		//DT time.Time //Greenwich
		//190911 to simplify storing into database
		//That is to avoid specific time types of different DBMS
		DT uint32 //Число секунд от 01.01.1970 по Гринвичу
	}

	Tag48 struct { //9 bytes
		SetelliteCount     int     //Число спутников; 4 bits; >=0; <1+2+4+8=15
		Correctness_Source int     //4 bits;0 - correct from glonass/gps; 2 - correct from base station; others not correct
		Latitude           float32 //<0 - south
		Longitude          float32 //<0 - west
	}
)

//GetRawTag searches the tag list (l) for a data of member (as []byte)  with the given tag number
//It returns a first found member or nill (with error), if the tag was not found.
func GetRawTag(l []RawTag, tagNum int) (rT []byte, err error) {
	if tagNum < 0 || tagNum > 256 {
		err = fmt.Errorf("gll.GetRawTag:tag number == %v???", tagNum)
		return
	}
	for i := 0; i < len(l); i++ {
		if l[i].TagNum == tagNum {
			rT = l[byte(i)].Data
			if rT == nil {
				err = fmt.Errorf("gll.GetRawTag: Data of tag of %v not defined", tagNum)
				return
			}
			return
		}
	}
	err = fmt.Errorf("gll.GetRawTag: Tag %v not found", tagNum)
	rT = nil
	return
}

//GetTagList сканирует данные (если они есть) "сырого" пакета от первого тэга (четвертый байт).
//То есть, чтобы найти следующий тэг надо считать предыдущий
func GetTagList(rawPack []byte) (l []RawTag, err error) {
	var currTagNum int
	var currTagLen int
	var currTag RawTag
	var currTagInd = 3 //Position of tag number (index of member of rawPack

	if (len(rawPack) < 5) || (rawPack[0] != 1) {
		err = fmt.Errorf("getTagList:Bad packet=%v", rawPack)
		return
	}
	if len(rawPack) == 5 { //Data absence
		return
	}

	//currTagNum = int(currTagInd)

	for {
		currTag = RawTag{}
		currTagNum = int(rawPack[currTagInd])
		if currTagLen, currTag.Descr, err = GetTagLen(currTagNum); err != nil {
			err = fmt.Errorf("getTagList:Поз %v;Тэг %v; ошибка определения длины:%v", currTagInd, currTagNum, err.Error())
			return
		}
		currTag.TagNum = currTagNum
		currTag.Data = rawPack[currTagInd+1 : currTagInd+1+currTagLen]

		if currTagInd+1+currTagLen > len(rawPack) {
			err = fmt.Errorf("getTagList:Bad packet - tag %v at pos %v goes out packet", currTagNum, currTagInd)
			return
		}
		currTag.Data = rawPack[currTagInd+1 : currTagInd+1+currTagLen]
		currTagInd = currTagInd + currTagLen + 1 //!190909

		//currTag.Data = ReverseBytes(currTag.Data)
		l = append(l, currTag)
		if currTagInd >= (len(rawPack) - 1 - 2) { //(len(rawPack)-1) is an index the last byte of the packet
			break
		}
	}
	return
}

//GetTagNums возвращает перечень номеров тэгов из массива "сырых" тэгов
func GetTagNums(l []RawTag) string {
	var s string
	for i := 0; i < len(l); i++ {
		s = s + strconv.Itoa(l[i].TagNum)
		if i != len(l)-1 {
			s = s + ", "
		}
	}
	return s
}

//GetTagLen возвращает длину тела тэга (длину данных), его описание и ошибку, если тэг не известен
//В совокупности с набором функций ParseTag1, ParseTag2, ... и типов Tag1, Tag2, ..., эта функция реализует
//таблицу "Тэги протокола Galileosky" из "Протокол обмена с сервером терминалов Galileosky"
func GetTagLen(tagNum int) (int, string, error) {
	var (
		ln    int
		descr string
		err   error
	)
	switch tagNum {
	case 1:
		ln = 1
		descr = "Версия железа"
	case 2:
		ln = 1
		descr = "Версия прошивки"
	case 3:
		ln = 15
		descr = "IMEI"
	case 4:
		ln = 2
		descr = "Идентификатор устройства"
	case 16:
		ln = 2
		descr = "Номер записи в архиве"
	case 32:
		ln = 4
		descr = "Дата и время"
	case 48:
		ln = 9
		descr = "Координаты в градусах, число спутников, признак корректности, источник координат"
	default:
		err = fmt.Errorf("kgll.GetTagLen: неизвестный тэг %v", tagNum)
	}
	return ln, descr, err

}

//Функции ParseTag1, ParseTag2, ... пытаются парсить полученный
// байтовый массив (rt -  raw tag, см. GetRawTag) в значение соответствующего типа
//Если это удается, то возвращается указатель на значение
// Если rt==nil, то возвращается nil, nil - то есть отсутствие массива на входе не считается за ошибку.
//Набор ошибок специфичен для каждой функции.
//ParseTag1 возвращает err!=nil если длина массива > 1
func ParseTag1(rt []byte) (*Tag1, error) { //Версия железа
	var err error
	var tg Tag1
	if rt == nil {
		return nil, nil
	}
	if len(rt) != 1 {
		err = fmt.Errorf("ParseTag1:Длина первого тега (версия железа) должна быть 1 байт.")
	}
	tg.HardVersion = rt[0]
	return &tg, err
}

//Общее описание см. ParseTag1
//ParseTag2 возвращает err!=nil если длина массива != 1
func ParseTag2(rt []byte) (*Tag2, error) { //Версия прошивки
	var err error
	var tg Tag2
	if rt == nil {
		return nil, nil
	}
	if len(rt) != 1 {
		err = fmt.Errorf("ParseTag2:Длина второго тега (версия прошивки) должна быть 1  байт.")
	}
	tg.FirmVersion = rt[0]
	return &tg, err
}

//Общее описание см. ParseTag1
//ParseTag3 возвращает err!=nil если длина массива != 15
func ParseTag3(rt []byte) (*Tag3, error) { //IMEI
	var err error
	var tg Tag3
	if rt == nil {
		return nil, nil
	}
	if len(rt) != 15 {
		err = fmt.Errorf("ParseTag3:Длина  тега (IMEI) должна быть 15  байт.")
	}
	tg.IMEI = string(rt)
	return &tg, err
}

//Общее описание см. ParseTag1
//ParseTag4 возвращает err!=nil если длина массива != GetTagLen(4)
//Если GetTagLen(4) возращает ошибку, то ParseTag4 паникует
func ParseTag4(rt []byte) (*Tag4, error) { //IMEI
	var err error
	var tg Tag4
	var dueLen int
	if rt == nil {
		return nil, nil
	}
	if dueLen, _, err = GetTagLen(4); err != nil {
		panic(fmt.Sprintf("GetTagLen(4) (в ParseTag4) вернула ошибку: %v", err.Error()))
	}
	if len(rt) != dueLen {
		err = fmt.Errorf("ParseTag3:Длина  тега 4 должна быть %v байт, но имеется %v .", dueLen, len(rt))
	}

	tg.DevID = uint16(rt[0]) + uint16(rt[1])*256
	return &tg, err
}

//Общее описание см. ParseTag1
func ParseTag16(rt []byte) (*Tag16, error) { //Archive Record identifier
	var tg16 Tag16
	var err error
	var dueLen int

	if rt == nil {
		return nil, nil
	}

	if dueLen, _, err = GetTagLen(16); err != nil {
		err = fmt.Errorf("GetTagLen(16) (в ParseTag16) вернула ошибку: %v", err.Error())
	}
	if dueLen != 2 {
		err = fmt.Errorf("ParseTag16:Длина  тега 16 должна быть 4 байта, но GetTagLen(16) вернула  %v .", dueLen)

	}
	if len(rt) != dueLen {
		err = fmt.Errorf("ParseTag16:Длина  тега 4 должна быть %v байт, но имеется %v .", dueLen, len(rt))
	}

	if err != nil {
		return nil, err
	}

	tg16.ArchiveRecID = uint16(rt[0]) + uint16(rt[1])*256

	return &tg16, err
}

//Общее описание см. ParseTag1
//"Дата и время"
func ParseTag32(rt []byte) (*Tag32, error) {
	var err error
	var tg32 Tag32
	var dueLen int
	if rt == nil {
		return nil, nil
	}

	if dueLen, _, err = GetTagLen(32); err != nil {
		err = fmt.Errorf("GetTagLen(4) (в ParseTag32) вернула ошибку: %v", err.Error())
	}
	if dueLen != 4 {
		err = fmt.Errorf("ParseTag32:Длина  тега 32 должна быть 4 байта, но GetTagLen(32) вернула  %v .", dueLen)
	}
	if len(rt) != dueLen {
		err = fmt.Errorf("ParseTag32:Длина  тега 42 должна быть %v байт, но имеется %v .", dueLen, len(rt))
	}

	if err != nil {
		return nil, err
	}

	//rt 4 bytes - amount of second since 1970/01/01 according Greenwich
	//var t0 time.Time
	var seconds uint32
	//if t0, err = time.Parse("2006/01/02", "1970/01/01"); err != nil {
	//	err = fmt.Errorf("ParseTag32: parsing 1970/01/01 err = %v .", dueLen)
	//	return nil, err
	//}
	//
	//seconds = uint32(rt[0]) + uint32(rt[1])*256 + uint32(rt[2])*256*265 + uint32(rt[3])*256*256*256
	seconds = binary.LittleEndian.Uint32(rt)

	//tg32.DT = t0.Add(time.Second * time.Duration(seconds))
	tg32.DT = seconds

	return &tg32, err
}

//Общее описание см. ParseTag1
// "Координаты в градусах, число спутников, признак корректности, источник координат"
func ParseTag48(rt []byte) (*Tag48, error) {
	var err error
	var tg48 Tag48
	var dueLen int
	//var latInt, latFrac int
	//var lonInt, lonFrac int

	//kerr.PrintDebugMsg(false, "t48", fmt.Sprintf("ParseTag48:%v", rt))

	if rt == nil {
		return nil, nil
	}

	if dueLen, _, err = GetTagLen(48); err != nil {
		err = fmt.Errorf("GetTagLen(4) (в ParseTag48) вернула ошибку: %v", err.Error())
	}
	if dueLen != 9 {
		err = fmt.Errorf("ParseTag48:Длина  тега 48 должна быть 9 байт, но GetTagLen(48) вернула  %v .", dueLen)
	}
	if len(rt) != dueLen {
		err = fmt.Errorf("ParseTag48:Длина  тега 48 должна быть %v байт, но имеется %v .", dueLen, len(rt))
	}

	if err != nil {
		return nil, err
	}

	tg48.SetelliteCount = int(rt[0] & 15)      //1+2+4+8
	tg48.Correctness_Source = int(rt[0] & 240) //16+32+64+128

	/*
		latInt = int(uint(rt[1])+uint(rt[2])*256+uint(rt[3])*256*256+uint(rt[4])*256*256*256) / 1000000
		lonInt = int(uint(rt[5])+uint(rt[6])*256+uint(rt[7])*256*256+uint(rt[8])*256*256*256) / 1000000
		latFrac = int(uint(rt[1])+uint(rt[2])*256+uint(rt[3])*256*256+uint(rt[4])*256*256*256) % 1000000
		lonFrac = int(uint(rt[5])+uint(rt[6])*256+uint(rt[7])*256*256+uint(rt[8])*256*256*256) % 1000000
		tg48.Latitude = float32(latInt) + float32(latFrac/1000000)
		tg48.Longitude = float32(lonInt) + float32(lonFrac/1000000)
	*/
	tg48.Latitude = float32(int32(binary.LittleEndian.Uint32(rt[1:5]))) / 1000000
	tg48.Longitude = float32(int32(binary.LittleEndian.Uint32(rt[5:9]))) / 1000000

	return &tg48, err
}

//Методы  (pt *Tag1)String(),  (pt *Tag2)String() ... возвращают строковае представления тэга для целей логгирования
func (pt *Tag1) String() string {
	var descr string
	var err error
	if _, descr, err = GetTagLen(1); err != nil {
		descr = err.Error()
	}

	return fmt.Sprintf("%v;%v", descr, pt.HardVersion)
}

func (pt *Tag2) String() string {
	var descr string
	var err error
	if _, descr, err = GetTagLen(2); err != nil {
		descr = err.Error()
	}

	return fmt.Sprintf("%v;%v", descr, pt.FirmVersion)
}

func (pt *Tag3) String() string {
	var descr string
	var err error
	if _, descr, err = GetTagLen(3); err != nil {
		descr = err.Error()
	}

	return fmt.Sprintf("%v;%v", descr, pt.IMEI)
}

func (pt *Tag4) String() string {
	var descr string
	var err error
	if _, descr, err = GetTagLen(4); err != nil {
		descr = err.Error()
	}

	return fmt.Sprintf("%v;%v", descr, pt.DevID)
}

func (pt *Tag16) String() string {
	var descr string
	var err error
	if _, descr, err = GetTagLen(16); err != nil {
		descr = err.Error()
	}

	return fmt.Sprintf("%v;%v", descr, pt.ArchiveRecID)

}

func (pt *Tag32) String() string {
	var descr string
	var err error
	if _, descr, err = GetTagLen(32); err != nil {
		descr = err.Error()
	}

	return fmt.Sprintf("%v;%v", descr, pt.DT)
}

func (pt *Tag48) String() string {
	var descr string
	var err error
	if _, descr, err = GetTagLen(48); err != nil {
		descr = err.Error()
	}

	return fmt.Sprintf("%v;(%v , %v)(Lat%v ,Lon%v)", descr, pt.SetelliteCount, pt.Correctness_Source, pt.Latitude, pt.Longitude)
}

func GetTagAsStr(rt RawTag) string {
	var err error
	switch rt.TagNum {
	case 1:
		var t *Tag1
		if t, err = ParseTag1(rt.Data); err != nil {
			return err.Error()
		} else {
			return t.String()
		}
	case 2:
		var t *Tag2
		if t, err = ParseTag2(rt.Data); err != nil {
			return err.Error()
		} else {
			return t.String()
		}
	case 3:
		var t *Tag3
		if t, err = ParseTag3(rt.Data); err != nil {
			return err.Error()
		} else {
			return t.String()
		}
	case 4:
		var t *Tag4
		if t, err = ParseTag4(rt.Data); err != nil {
			return err.Error()
		} else {
			return t.String()
		}
	case 16:
		var t *Tag16
		if t, err = ParseTag16(rt.Data); err != nil {
			return err.Error()
		} else {
			return t.String()
		}
	case 32:
		var t *Tag32
		if t, err = ParseTag32(rt.Data); err != nil {
			return err.Error()
		} else {
			return t.String()
		}
	case 48:
		var t *Tag48
		if t, err = ParseTag48(rt.Data); err != nil {
			return err.Error()
		} else {
			return t.String()
		}
	default:
		return fmt.Sprintf("kgll.GetTagAsStr:unknow tag=%v", rt.TagNum)
	}
}

func ReverseBytes(bs []byte) []byte {
	var rbs []byte = make([]byte, len(bs))
	for i := 0; i < len(bs); i++ {
		rbs[i] = bs[len(bs)-1-i]
	}
	return rbs
}

//Pair_32_48 содержит "сырые" тэги 32(время) и 48(кооддинаты)
type Pair_32_48 [2]RawTag

//Seek_32_48 просматривает переданный массив "сырых тэгов" (l []RawTag)
// и если встречает последовательно (возможно, с другими тэгами межту ними) идущие тэги 32 и 48 (время и координаты)
//То есть, встретив тэг 32, функция далее сканирует массив в поисках тэга 48.
//Если он найден, то пара запоминается и продолжается поиск следующей пары
//Если ни одной пары не найдено, возвращается nil
func Seek_32_48(l []RawTag) []Pair_32_48 {
	var pair Pair_32_48
	var isPair bool
	var pairs []Pair_32_48
	var i, j int

	for {
		isPair = false
		if l[i].TagNum == 32 {
			pair[0] = l[i]
			j = i
			for {
				if l[j].TagNum == 48 {
					pair[1] = l[j]
					isPair = true
					i = j
					break
				}
				j++
				if j >= len(l) {
					i = j
					break
				}
			}
		}
		if isPair {
			pairs = append(pairs, pair)
		}
		i++
		if i >= len(l) {
			break
		}
		//kerr.PrintDebugMsg(false, "save", fmt.Sprintf("Seek_32_48:i=%v; j=%v", i, j))
	}
	//kerr.PrintDebugMsg(false, "save", fmt.Sprintf("Seek_32_48:pairs=%v;", len(pairs)))

	return pairs
}

//GetDevID просматривает переданный массив "сырых тэгов" (l []RawTag)
//и, если находит 4 (идентификатор устройства) то парсит его
func GetDevID(l []RawTag) (id uint16, err error) {
	var rt4Data []byte
	var t4 *Tag4
	if rt4Data, err = GetRawTag(l, 4); err != nil {
		err = fmt.Errorf("GetDevID: getting tag raw data error =%v", err.Error())
		return
	}
	if t4, err = ParseTag4(rt4Data); err != nil {
		err = fmt.Errorf("GetDevID: parding tag error =%v", err.Error())
		return
	}
	id = t4.DevID
	return
}

//GetIMEI просматривает переданный массив "сырых тэгов" (l []RawTag)
//и, если находит 3 (IMEI) то парсит его
func GetIMEI(l []RawTag) (imei string, err error) {
	var rt3Data []byte
	var t3 *Tag3
	if rt3Data, err = GetRawTag(l, 3); err != nil {
		err = fmt.Errorf("GetDevID: getting tag raw data error =%v", err.Error())
		return
	}
	if t3, err = ParseTag3(rt3Data); err != nil {
		err = fmt.Errorf("GetDevID: parding tag error =%v", err.Error())
		return
	}
	imei = t3.IMEI
	return
}
