package formaterror

import "strings"

var errMsg = make(map[string]string)
var err error

func FormatError(errString string) map[string]string {
	if strings.Contains(errString, "username") {
		errMsg["Taken_username"] = "username already taken"
	}
	if strings.Contains(errString, "email") {
		errMsg["Taken_email"] = "email already taken"
	}
	if strings.Contains(errString, "title") {
		errMsg["Taken_title"] = "title already taken"
	}
	if strings.Contains(errString, "hashedPassword") {
		errMsg["Incorrect_password"] = "incorrect password"
	}
	if strings.Contains(errString, "record not found") {
		errMsg["No_record"] = "no record found"
	}
	if strings.Contains(errString, "double join") {
		errMsg["Double_join"] = "you cannot join this meeting twice"
	}

	if len(errMsg) > 0 {
		return errMsg
	}

	if len(errMsg) == 0 {
		errMsg["Incorrect_details"] = "incorrect details"
		return errMsg
	}
	return nil
}
