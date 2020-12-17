package ktime

import (
	"errors"
	"fmt"
	"kot_common/kerr"
	"strconv"
	"strings"
)

//it takes hours (e.g. H="4.25") and return minutes (in the case 4*60 + 15)
//it takes "," or "." as floating-point
//if these characters are not single it returns error
func HoursToMinutes(H string) (min string, err error) {
	var (
		flStrComps             []string
		fPointDot, fPointComma bool
		fPoint                 string
		intPart, fracPart      int
		M                      int
	)

	defer func() {
		if rec := recover(); rec != nil {
			err = errors.New(kerr.GetRecoverErrorText(rec))
		}
	}()

	if strings.ContainsRune(H, '-') {
		panic(fmt.Sprintf("Плохие часы, со знаком - : %v", H))
	}

	fPointDot = strings.ContainsRune(H, '.')
	fPointComma = strings.ContainsRune(H, ',')

	if fPointComma && fPointDot {
		panic(fmt.Sprintf("Плохие часы, и с точкой и с зяпятой: %v", H))
	}

	if fPointDot {
		fPoint = "."
	} else {
		fPoint = ","
	}

	flStrComps = strings.Split(H, fPoint)

	if len(flStrComps) > 2 {
		panic(fmt.Sprintf("Плохие часы, слишком много отделителей дробной части: %v", H))
	}

	if intPart, err = strconv.Atoi(flStrComps[0]); err != nil {
		panic(fmt.Sprintf("Беда с целой частью часов: %v", H))
	}

	//fmt.Printf("flStrComps=%v;intPart=%v\n", flStrComps, intPart)

	if len(flStrComps) == 1 { // no a fractional part
		M = intPart * 60
		min = strconv.Itoa(M)
		return
	}

	if fracPart, err = strconv.Atoi(flStrComps[1]); err != nil {
		panic(fmt.Sprintf("Беда с дробной частью часов: %v", H))
	}

	switch {
	case fracPart < 10:
		fracPart = fracPart * 10
	case fracPart > 10:
		fracPart = int(fracPart / 10)
	}

	//fmt.Printf("intPart=%v, fracPart=%v, M=%v\n", intPart, fracPart, M)

	M = intPart*60 + (fracPart*60)/100

	min = strconv.Itoa(M)
	return

}
