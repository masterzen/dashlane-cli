package dashlane

import (
	"fmt"
	"net/url"
)

type ExistResult int

const (
	EXIST_YES ExistResult = iota
	EXIST_YES_OTP_NEWDEVICE
	EXIST_NO
	EXIST_ERROR
)

const EXIST_URI = "https://ws1.dashlane.com/6/authentication/exists"

func Exist(login string) (ExistResult, error) {
	data := url.Values{}
	data.Set("login", login)
	data.Set("isOTPAware", "true")

	body, err := PostData(EXIST_URI, data)
	if err != nil {
		return EXIST_ERROR, err
	}

	switch body {
	case "YES":
		return EXIST_YES, nil
	case "YES_OTP_NEWDEVICE":
		return EXIST_YES_OTP_NEWDEVICE, nil
	case "NO":
		return EXIST_NO, nil
	default:
		return EXIST_ERROR, fmt.Errorf("Unknown exist return value: %v", body)
	}
}
