package dashlane

import (
	"fmt"
	"net/url"
)

const SENDTOKEN_URI = "https://www.dashlane.com/6/authentication/sendtoken"
const REGISTER_URI = "https://www.dashlane.com/6/authentication/registeruki"

type SendTokenResult int

const (
	UKI_SUCCESS SendTokenResult = iota
	UKI_OTP_NEEDED
	UKI_ERROR
)

func SendToken(login string) (SendTokenResult, error) {
	data := url.Values{}
	data.Set("login", login)
	data.Set("isOTPAware", "true")

	body, err := PostData(SENDTOKEN_URI, data)
	if err != nil {
		return UKI_ERROR, err
	}

	switch body {
	case "SUCCESS":
		return UKI_SUCCESS, nil
	case "OTP_NEEDED":
		return UKI_OTP_NEEDED, nil
	default:
		return UKI_ERROR, fmt.Errorf("Unknown return value: %v", body)
	}
}

func RegisterUki(devicename, login, token, uki string) error {
	data := url.Values{}
	data.Set("login", login)
	data.Set("devicename", devicename)
	data.Set("token", token)
	data.Set("uki", uki)
	data.Set("platform", "webaccess")
	data.Set("temporary", "0")

	body, err := PostData(REGISTER_URI, data)
	if err != nil {
		return err
	}

	switch body {
	case "SUCCESS":
		return nil
	default:
		return fmt.Errorf("Unknown return value: %v", body)
	}
}
