package ktrack

//type RawSFRD []byte

//Таблица А.3 - Состав пакета протокола транспортного уровня
// PcktTL stands for a packet of transport level
type PcktTL struct {
	_1M_PRV                byte //M1//(Protocol Version)
	_2M_SKID               byte //M2// (Security Key ID)
	_3M_PRF_RTE_ENA_CMP_PR byte //M3
	//Параметр PRF (7-6 bits) определяет префикс заголовка протокола и содержитзначение 00.
	//Поле RTE (Route) (5 bit) определяет необходимость дальнейшей маршрутизации данного пакета на ...
	//Поле ENA (Encryption Algorithm) (4-3 bits) определяет код алгоритма,используемый для шифрования данных из поля SFRD. Если поле имеет значение 00, то данные в поле SFRD не шифруют.
	//Поле CMP (Compressed) (2 bit) определяет, используется ли сжатие данных из поля SFRD. Если поле имеет значение 1, то данные в поле SFRD считаются сжатыми.
	//Поле PR (Priority) (1-0 bits) определяет приоритет маршрутизации данного пакета и может принимать следующие значения: - 00 - наивысший; - 01 - высокий; - 10 - средний; - 11 - низкий.
	_4M_HL byte //M4 //(Header Length) А.5.12 Поле HL - длина заголовка протокола в байтах с учетом байта
	//контрольной суммы (поля HCS).
	_5M_HE    byte   //M5(Header Encoding) А.5.13 Поле НЕ определяет применяемый метод кодирования, следующей за данным параметром части заголовка протокола.
	_6_7M_FDL uint16 //M6-7//(Frame Data Length)
	_8_9M_PID uint16 //M8-9//(Packet Identifier)//А.5.15
	_10M_PT   byte   //M10//(Packet Type)
	//А.5.16 Поле РТ - тип пакета транспортного уровня. Поле РТ может принимать следующие значения:
	//- 0 - EGTS_PT_RESPONSE (подтверждение на пакет транспортного уровня);
	//- 1 - EGTS_PT_APPDATA (пакет, содержащий данные протокола уровня поддержки услуг);
	//- 2 - EGTS_PT_SIGNED_APPDATA (пакет, содержащий данные протокола уровня поддержки услуг с цифровой подписью).
	_11_12O_PRA int //O11-12// (Peer Address) <0 - it is absence into raw packet
	//А.5.17 Поле PRA - адрес аппаратно-программного комплекса, на котором
	//данный пакет сгенерирован. Данный адрес является уникальным в рамках
	//сети и используется для создания пакета-подтверждения на принимающей стороне.
	_13_14O_RCA int //O13-14//(Recipient Address)<0 - it is absence into raw packet
	//А.5.18 Поле RCA - адрес аппаратно-программного комплекса, для которого
	//данный пакет предназначен. По данному адресу производят идентификацию
	//принадлежности пакета определенного аппаратно-программного комплекса и
	//его маршрутизация при использовании промежуточных аппаратно-
	//программных комплексов.
	_15O_TTL int //O15// (Time To Live) <0 - it is absence into raw packet
	//А.5.19 Поле TTL - время жизни пакета при его маршрутизации между
	//аппаратно-программными комплексами.
	_16M_HCS byte //M11|16//(Header Check Sum)
	//А.5.20 Поле HCS - контрольная сумма заголовка протокола (начиная с поля
	//"PRV" до поля "HCS", не включая поле "HCS"). Для подсчета значения поля
	//HCS ко всем байтам указанной последовательности применяется алгоритм
	//CRC-8.

	_17O_SFRD []byte // (Services Frame Data) nil means absence
	//А.5.21 Поле SFRD - структура данных, зависящая от типа пакета и содержащая информацию протокола уровня поддержки услуг.
	//See type EGTS_PT_RESPONSE,

	_18O_SFRCS int //(Services Frame Data Check Sum) <0 - it is absence into raw packet
}

//Таблица В.1 - Формат отдельной записи протокола уровня поддержки услуг
// PcktSL_Rec stands for a record of packet of service level
type PcktSL_Rec struct {
	RL  uint16 //M2(Record Length)- определяет размер данных из поля RD
	RN  uint16 //M2(Record Number)
	RFL byte   //M1(Record Flags)
	//SSOD(1)_RSOD(1)_GRP(1)_RPP(2)_TMFE(1)_EVFE(1)_OBFE(1)
	//SSOD 1 (Source Service On Device) - битовый флаг, определяющий расположение сервиса-отправителя:
	//1 - сервис-отправитель расположен на стороне АСН (авторизуемой телематической платформой (ТП)),
	//0 - сервис-отправитель расположен на авторизующей ТП.
	//RSOD 1 (Recipient Service On Device) - битовый флаг, определяющий расположение Сервиса-получателя:
	//1 - сервис-получатель расположен на стороне АСН (авторизуемой ТП),
	//0 - сервис-получатель расположен на авторизующей ТП.
	//GRP - (Group) - битовый флаг, определяющий принадлежность передаваемых данных определенной группе, идентификатор которой указан в поле OID:
	//1 - данные предназначены для группы,
	//0 - принадлежность группе отсутствует.
	//RPP (Record Processing Priority) - битовое поле, определяющее приоритет обработки данной записи сервисом:
	//00 - наивысший,
	//01 - высокий,
	//10 - средний,
	//11 - низкий.
	//TMFE (Time Field Exists) - определяющее наличие в данном пакете поля ТМ:
	//1 - поле ТМ присутствует,
	//0 - поле ТМ отсутствует.
	//EVFE (Event ID Field Exists) - определяющее наличие в данном пакете поля EVID:
	//1 - поле EVID присутствует,
	//0 - поле EVID отсутствует.
	//OBFE (Object ID Field Exists) -  определяющее наличие в данном пакете поля OID:
	//1 - поле OID присутствует;
	//0 - поле OID отсутствует.
	OID  uint32 //O4(Object Identifier)
	EVID uint32 //O4(Event Identifier)
	TM   uint32 //O4(Time)
	SST  byte   //M1(Source Service Type)
	RST  byte   //M1(Recipient Service Type)
	RD   []byte //(Record Data) поле, содержащее информацию, присущую
	//определенному типу сервиса (одну или несколько подзаписей сервиса типа,
	//указанного в поле SST или RST, в зависимости от вида предаваемой
	//информации).
}

//Таблица В.2 - Формат отдельной подзаписи протокола уровня поддержки услуг
// PcktSL_SubRec stands for a subrecord of packet of service level
type PcktSL_SubRec struct {
	SRT byte //M1(Subrecord Туре) тип подзаписи (подтип передаваемых данных в
	//рамках общего набора типов одного сервиса). Тип 0 - специальный,
	//зарезервирован за подзаписью подтверждения данных для каждого сервиса.
	//Конкретные значения номеров типов подзаписей определяются логикой
	//самого сервиса. Протокол указывает лишь то, что этот номер должен
	//присутствовать, а нулевой идентификатор зарезервирован;
	SRL uint16 //M2(Subrecord Length) - длина данных в байтах подзаписи в поле SRD;
	SRD []byte //(Subrecord Data)
}

//А.6.2 Структура данных пакета EGTS_PT_RESPONSE
//Содержит информацию о результате обработки данных протокола
//транспортного уровня, полученного ранее. В таблице А.5 представлен формат
//поля SFRD для пакета типа EGTS_PT_RESPONSE.
//Таблица A.5 - Формат поля SFRD для пакета типа EGTS_PT_RESPONSE

type EGTS_PT_RESPONSE struct {
	RPID   uint16   //RPID - идентификатор пакета транспортного уровня, подтверждение на который сформировано.
	PR     byte     //PR - код результата обработки части пакета, относящейся к транспортному уровню.
	SDRArr [][]byte //SDR 0, SDR 1, ... SDR n содержат информацию уровня поддержки услуг.
}

//Таблица А.14 - Коды результатов обработки PR (Processing Result)
const (
	EGTS_PC_OK          byte = 0 //Успешно обработано
	EGTS_PC_IN_PROGRESS byte = 1 //В процессе обработки

	EGTS_PC_UNS_PROTOCOL    byte = 128 //Неподдерживаемый протокол
	EGTS_PC_DECRYPT_ERROR   byte = 129 // Ошибка декодирования
	EGTS_PC_PROC_DENIED     byte = 130 // Обработка запрещена
	EGTS_PC_INC_HEADERFORM  byte = 131 // Неверный формат заголовка
	EGTS_PC_INC_DATAFORM    byte = 132 // Неверный формат данных
	EGTS_PC_UNS_TYPE        byte = 133 // Неподдерживаемый тип
	EGTS_PC_NOTEN_PARAMS    byte = 134 // Неверное число параметров
	EGTS_PC_DBL_PROC        byte = 135 // Попытка повторной обработки
	EGTS_PC_PROC_SRC_DENIED byte = 136 // Обработка данных от источника запрещена
	EGTS_PC_HEADERCRC_ERROR byte = 137 // Ошибка контрольной суммы заголовка
	EGTS_PC_DATACRC_ERROR   byte = 138 // Ошибка контрольной суммы данных
	EGTS_PC_INVDATALEN      byte = 139 // Некорректная длина данных
	EGTS_PC_ROUTE_NFOUND    byte = 140 // Маршрут не найден
	EGTS_PC_ROUTE_CLOSED    byte = 141 // Маршрут закрыт
	EGTS_PC_ROUTE_DENIED    byte = 142 // Маршрутизация запрещена
	EGTS_PC_INVADDR         byte = 143 // Неверный адрес
	EGTS_PC_TTLEXPIRED      byte = 144 // Превышено число ретрансляции данных
	EGTS_PC_NO_ACK          byte = 145 // Нет подтверждения
	EGTS_PC_OBJ_NFOUND      byte = 146 // Объект не найден
	EGTS_PC_EVNT_NFOUND     byte = 147 // Событие не найдено
	EGTS_PC_SRVC_NFOUND     byte = 148 // Сервис не найден
	EGTS_PC_SRVC_DENIED     byte = 149 // Сервис запрещен
	EGTS_PC_SRVC_UNKN       byte = 150 // Неизвестный тип сервиса
	EGTS_PC_AUTH_DENIED     byte = 151 // Авторизация запрещена
	EGTS_PC_ALREADY_EXISTS  byte = 152 // Объект уже существует
	EGTS_PC_ID_NFOUND       byte = 153 // Идентификатор не найден
	EGTS_PC_INC_DATETIME    byte = 154 // Неправильная дата и время
	EGTS_PC_IO_ERROR        byte = 155 // Ошибка ввода/вывода
	EGTS_PC_NO_RES_AVAIL    byte = 156 // Недостаточно ресурсов
	EGTS_PC_MODULE_FAULT    byte = 157 // Внутренний сбой модуля
	EGTS_PC_MODULE_PWR_FLT  byte = 158 // Сбой в работе цепи питания модуля
	EGTS_PC_MODULE_PROC_FLT byte = 159 // Сбой в работе микроконтроллера модуля
	EGTS_PC_MODULE_SW_FLT   byte = 160 // Сбой в работе программы модуля
	EGTS_PC_MODULE_FW_FLT   byte = 161 // Сбой в работе внутреннего ПО модуля
	EGTS_PC_MODULE_IO_FLT   byte = 162 // Сбой в работе блока ввода/вывода модуля
	EGTS_PC_MODULE_MEM_FLT  byte = 163 // Сбой в работе внутренней памяти модуля
	EGTS_PC_TEST_FAILED     byte = 164 // Тест не пройден
)
