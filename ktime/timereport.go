package ktime

import (
	"fmt"
	"time"
)

type TimeReport struct {
	who      string
	leaf     bool // if Leaf than Tree==nil and D!=0
	cosumers []*TimeReport
	d        time.Duration
	why_why  string // if it !="" then there are errors and it is the errors mesage
}

func CreateLeaf(who string, d time.Duration) *TimeReport {
	var tr = TimeReport{}
	tr.who = who
	tr.leaf = true
	if d >= 0 {
		tr.d = d
	} else {
		tr.why_why = fmt.Sprintf("TimeReport.CreateLeaf; who=%v;d(%v)<0?!", who, d)
	}

	return &tr
}

func CreateNotLeaf(who string) *TimeReport {
	var tr = TimeReport{}
	tr.who = who

	return &tr
}

func (tr *TimeReport) AddCosumer(c *TimeReport) {
	if tr.leaf {
		tr.why_why = tr.why_why + "\n" + fmt.Sprintf("TimeReport.AddCosumer; attempting add to a leaf; who=%v?!", tr.who)
		return
	}
	if tr.cosumers == nil {
		tr.cosumers = make([]*TimeReport, 0)
	}
	tr.cosumers = append(tr.cosumers, c)
}

func (tr *TimeReport) SetDuration(d time.Duration) {
	tr.d = d
}
func (tr *TimeReport) ToSinpleString(shift string) string {
	return shift + fmt.Sprintf("%v(%v)d=%v; why_why=%v\n", tr.who, tr.leaf, tr.d, tr.why_why)
}
func (tr *TimeReport) ToString(shift string) string {
	var s string
	s = tr.ToSinpleString(shift)
	if tr.cosumers != nil {
		for _, item := range tr.cosumers {
			s = s + item.ToString(shift+"\t")
		}
	}
	return s
}
