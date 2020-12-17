// smae
//That stands for "server messages about error"
//Here are functions which are producing server messages about error (ССО, см KSCEX)

package ksess

import (
	"strconv"
)

//It makes a  massage of ParsingErr
//This is the only public function of the kind for outer using
//That is that impossible to render the income message as map[string]string
func Sso1(user_id int, errMess string) (sso map[string]string) {
	sso = make(map[string]string)
	sso["action_type"] = "ParsingErr"
	sso["user_id"] = strconv.Itoa(user_id)
	sso["res_code"] = "-1"
	sso["err_msg"] = errMess
	return
}

func sso2(user_id string, errMess string) (sso map[string]string) {
	sso = make(map[string]string)
	sso["action_type"] = "NotUser_id"
	sso["user_id"] = user_id
	sso["res_code"] = "-2"
	sso["err_msg"] = errMess
	return
}

func sso3(user_id string, errMess string) (sso map[string]string) {
	sso = make(map[string]string)
	sso["action_type"] = "NotAction_name"
	sso["user_id"] = user_id
	sso["res_code"] = "-3"
	sso["err_msg"] = errMess
	return
}

func sso4(user_id string, errMess string) (sso map[string]string) {
	sso = make(map[string]string)
	sso["action_type"] = "User_idNotMatch"
	sso["user_id"] = user_id
	sso["res_code"] = "-4"
	sso["err_msg"] = errMess
	return
}

func sso5(user_id string, errMess string) (sso map[string]string) {
	sso = make(map[string]string)
	sso["action_type"] = "MarshalErr"
	sso["user_id"] = user_id
	sso["res_code"] = "-5"
	sso["err_msg"] = errMess
	return
}

func sso6(user_id string, errMess string) (sso map[string]string) {
	sso = make(map[string]string)
	sso["action_type"] = "NoSourceFiedl"
	sso["user_id"] = user_id
	sso["res_code"] = "-6"
	sso["err_msg"] = errMess
	return
}

func sso7(user_id string, errMess string) (sso map[string]string) {
	sso = make(map[string]string)
	sso["action_type"] = "NotAllowedErr_mesField"
	sso["user_id"] = user_id
	sso["res_code"] = "-7"
	sso["err_msg"] = errMess
	return
}

func sso10(user_id string, errMess string) (sso map[string]string) {
	sso = make(map[string]string)
	sso["action_type"] = "parserSocket_panic"
	sso["user_id"] = user_id
	sso["res_code"] = "-7"
	sso["err_msg"] = errMess
	return
}
