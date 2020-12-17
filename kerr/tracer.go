// tracer
package kerr

//import (
//	"fmt"
//)

type TracePoint struct {
	Label string
	Phase string
}

func (tp TracePoint) htmlIl() string {
	return tp.Label + " : " + tp.Phase
}

type Trace struct {
	Before []TracePoint
	Next   *Trace
	After  []TracePoint
}

func (trc *Trace) AddBefore(l, p string) {
	var tp TracePoint
	if trc == nil {
		trc = &Trace{}
	}
	tp = TracePoint{l, p}
	trc.Before = append(trc.Before, tp)
}

func (trc *Trace) AddAfter(l, p string) {
	var tp TracePoint
	if trc == nil {
		trc = &Trace{}
	}
	tp = TracePoint{l, p}
	trc.After = append(trc.Before, tp)
}

func (trc *Trace) GetNext() *Trace {
	if trc == nil {
		panic("kerr - (&Trace).GetNext : &Trace==nil")
	}
	if trc.Next == nil {
		trc.Next = &Trace{}
	}
	return trc.Next
}

func (trc *Trace) htmlUl() (res string) {
	if trc == nil {
		return ""
	}
	res = "<ul>"
	for _, mb := range trc.Before {
		res = res + mb.htmlIl()
	}

	if trc.Next != nil {
		res = res + trc.Next.htmlUl()
	}

	for _, mb := range trc.After {
		res = res + mb.htmlIl()
	}

	res = res + "</ul>"
	return res
}
