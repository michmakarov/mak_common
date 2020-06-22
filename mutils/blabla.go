package mutils

import (
	"fmt"
)

func getBlaBla(num string) string{
switch num {
case "200321": return`
It is first version that is fully as a 191029 version of kutils.
Only PrintFileContent was added
`
case "200323": return`
Here I am going to rid of kot_common/kerr
`
default: return fmt.Sprintf("mutils.getBlaBla:No such version-%v", num)

}
}//func

