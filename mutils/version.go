package mutils

import (
	"fmt"
	//"net/http"
	"strconv"
)

var commit_data_1 = "No_git_commit_data"
var branch_data_1 = "No_git_branch_data"

type VersionDescr struct {
	Number       string //Version number is a string of format "190926", that is "<Year><Month><Day>"
	ProgName     string //   = "kitils"
}

//versinList defines VersionDescr for each version that has been occured
//Current version has index 0, previous one - index 1, and so on.
var versionList = []VersionDescr{
	{"200321", "mutils"}, //, branch_data_1, commit_data_1}//, "191029", "developing", blabla_200321},
	{"200323", "mutils"}, //The question: for what is the ProgName field needed?
}

func GetVesionInfo() string {
	return getVerInfo(versionList[0].Number)
}

func GetVerNum() string { return versionList[0].Number }


func getVerInfo(num string) string {
	for _, it := range versionList {
		if it.Number == num {
			return getBlaBla(num)
		}
	}
	return fmt.Sprintf("mutils.getVerInfo:No such version - %v", num)
}

func getVersionText(num string) string {
	return getVerInfo(num)
}

func getVerList() string {
	var s string
	for _, it := range versionList {
		s = s + it.Number + "<br>"
	}
	return s
}
func isNum(ver string) bool {
	var err error
	if _, err = strconv.Atoi(ver); err != nil {
		return false
	}
	return true
}
