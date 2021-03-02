package kconfig

/*
Кой-какие соображения по поводу конфигурации котового сервера.
1. Котовых серверов может быть много, со своими отдельными задачями. Что же их объединяет?
1.1 Один ... so what is the first?
In my opinion the first is that the natural type for configuration data is map[string]string
180517 Let's return here in context of KOT server framework (com180417 task)
Yes, the KOT servers may be many, and each has its own file of configuration!
For what then is this package needed?
The package is needed for giving common functionality for each case!
Let's it be so. Now some questions are.
1. May a configuration file not be ? We answer: not, it always must be.
2. May it be empty? We answer: yes.
3. What is about structure of the file? Let it be a right JOSON object. As here
{
"name1":<any JSON value>,
...
"//":"it is a comment, it will not be read",
"nameN":<any JSON value>
}
And members with "//" name are comments which must not be read by the functionality
*/
import (
	"encoding/json"
	"errors"
	"fmt"

	"mak_common/klog"
	"os"
	"strconv"
)

const (
	ProgName     = "kconfig"
	Version      = "---201216_rels:51d5bee--*main--210222_2108---" //"190612" //"181003" //"180823"
	VersionState = "developing"
)

func init() {
	fmt.Println(GetVesionInfo())
}

func GetVesionInfo() string {
	return ProgName + "_" + Version + " : " + VersionState
}

var configFileName = "config.json" //190612

type Configuration map[string]interface{}

type CheckConfiguration func(map[string]interface{}) error

//it supposes existing config file in working directory
func ReadConfig(cf CheckConfiguration) (c Configuration, err error) {
	c = make(map[string]interface{})
	var decoded interface{}
	var d map[string]interface{}
	var ok bool

	if _, err = os.Stat(configFileName); os.IsNotExist(err) {
		err = nil
		c = nil
		return
	}

	f, err := os.Open(configFileName)

	if err != nil {
		c = nil
		return
	} else {
		defer f.Close()
	}

	decoder := json.NewDecoder(f)

	if err = decoder.Decode(&decoded); err != nil {
		c = nil
		return
	}
	if d, ok = decoded.(map[string]interface{}); !ok {
		c = nil
		err = errors.New("In config.json must be a valid json object")
		return
	}

	//fmt.Printf("--M-- d=%v\n", d)

	for k, v := range d {
		if k != "//" {
			c[k] = v
		}
	}

	if cf != nil {
		if err = cf(c); err != nil {
			c = nil
			return
		}
	}

	return
}

func (c Configuration) Print(l *klog.Klogger) {
	if l != nil {
		l.Printf("===Configuration(ver=%v);  len=%v", Version, len(c))
		for k, v := range c {
			l.Printf("%v=%v", k, v)
		}
		l.Printf("===============")
	} else {
		fmt.Printf("===Configuration(ver=%v);  len=%v\n", Version, len(c))
		for k, v := range c {
			fmt.Printf("%v=%v\n", k, v)
		}
		fmt.Printf("===============\n")
	}
}

func (c Configuration) GetAsString(key string) (value string) {
	var val interface{}
	var ok bool

	val = c[key]
	if value, ok = val.(string); !ok {
		value = ""
		//err = errors.New(fmt.Sprintf("kconfig.GetAsString: %v of %v can not be converted to string", val, key))
	}
	return
}

func (c Configuration) GetAsInt(key string) (value int, err error) {
	var val interface{}
	var ok bool
	var valueFloat float64

	//210104 07:02 How might such nonsense be occured?
	//It is not nonsense at all! 210104 07:45

	val = c[key]
	//fmt.Printf("--M--val==%v;key==%v\n", val, key)
	if valueFloat, ok = val.(float64); !ok {
		err = fmt.Errorf("kconfig.GetAsInt: %v(type %T) of %v can not be converted to float64", val, val, key)
	} else {
		value = int(valueFloat)
	}
	return
}

//210219 10:20 for 201216_rels
func (c Configuration) GetAsUint(key string) (value uint, err error) {
	var val interface{}
	var ok bool
	var valueFloat float64
	var intValue int

	val = c[key]
	if valueFloat, ok = val.(float64); !ok {
		err = fmt.Errorf("kconfig.GetAsUint: %v(type %T) of %v can not be converted to float64\n", val, val, key)
	} else {
		intValue = int(valueFloat)
	}
	if intValue < 0 {
		err = fmt.Errorf("kconfig.GetAsUint: for %v given negative value %v\n", key, intValue)
	} else {
		value = uint(intValue)
	}
	return
}

//210104 07:45 the initial value must be in form of "XXXXXXXX" (eight or less symbols), where X is 1 or 0
func (c Configuration) GetAsByte(key string) (value byte, err error) {
	var val interface{}
	var valStr string
	var valUint64 uint64
	var ok bool

	val = c[key]
	if valStr, ok = val.(string); !ok {
		err = fmt.Errorf("kconfig.GetAsByte: %v of %v does not hold a string", val, key)
		return
	}
	if len(valStr) > 8 {
		err = fmt.Errorf("kconfig.GetAsByte: %v of %v holds more than 8 symbols and cannot rendered as byte", val, key)
		return
	}
	if len(valStr) == 0 {
		valStr = "00000000"
	}
	if valUint64, err = strconv.ParseUint(valStr, 2, 8); err != nil {
		err = fmt.Errorf("kconfig.GetAsByte: %v of %v cannot be parsed into a byte", val, key)
		return
	}

	return byte(valUint64), nil
}

func (c Configuration) GetAsStringArr(key string) (value []string, err error) {
	var (
		val          interface{}
		valAsArr     []interface{}
		item         interface{}
		itemAsString string
		ok           bool
	)

	val = c[key]
	if valAsArr, ok = val.([]interface{}); !ok {
		err = fmt.Errorf("kconfig.GetAsStringArr: %v(type %T) of %v can not be converted to []inteface{}", val, val, key)
		value = nil
		return
	}
	for _, item = range valAsArr {
		if itemAsString, ok = item.(string); !ok {
			err = fmt.Errorf("kconfig.GetAsStringArr: %v(type %T) of %v can not be converted to string", item, item, valAsArr)
			value = nil
			return
		}
		value = append(value, itemAsString)
	} //for
	return
}

func (c Configuration) GetAsBool(key string) (value bool, err error) {
	var val interface{}
	var ok bool
	var valueBool bool

	val = c[key]
	if valueBool, ok = val.(bool); !ok {
		err = fmt.Errorf("kconfig.GetAsint: %v(type %T) of %v can not be converted to bool", val, val, key)
	} else {
		value = valueBool
	}
	return
}

func SetConfigFileName(cfn string) (err error) { //190612
	if cfn == "" {
		err = fmt.Errorf("kconfig.SetConfigFileName: name of file must not be empty")
		return
	}
	configFileName = cfn
	return
}
