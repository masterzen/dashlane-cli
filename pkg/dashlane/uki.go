package dashlane

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math/big"
	"net/url"
	"runtime"
	"time"
)

const SENDTOKEN_URI = "https://ws1.dashlane.com/6/authentication/sendtoken"
const REGISTER_URI = "https://ws1.dashlane.com/6/authentication/registeruki"

type SendTokenResult int

const (
	UKI_SUCCESS SendTokenResult = iota
	UKI_OTP_NEEDED
	UKI_ERROR
)

func (dl *Dashlane) SendToken(login string) (SendTokenResult, error) {
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

func (dl *Dashlane) RegisterUki(devicename, login, token, uki string) error {
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

func getMD5Hash(text string) string {
	hasher := md5.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}

func (dl *Dashlane) GenerateUki() (string, error) {
	r, err := rand.Int(rand.Reader, big.NewInt(268435456))
	if err != nil {
		return "", err
	}

	var time = fmt.Sprintf("%d", time.Now().Unix())
	var text = runtime.GOOS + runtime.GOARCH + time + r.Text(16)
	var hashed = getMD5Hash(text)

	return hashed + "-webaccess-" + time, nil
}
