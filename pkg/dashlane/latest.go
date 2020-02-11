package dashlane

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/sirupsen/logrus"
)

const LATEST_URI = "https://ws1.dashlane.com/12/backup/latest"

func (dl *Dashlane) LatestToken(login string, code string) (string, error) {
	logrus.Debug("LatestToken: ", login, code)
	data := url.Values{}
	data.Set("login", login)
	data.Set("otp", code)
	data.Set("timestamp", "1")
	data.Set("lock", "nolock")
	data.Set("sharingTimestamp", "0")
	data.Set("sharingSkipped", "webapp")

	body, err := PostData(LATEST_URI, data)
	if err != nil {
		return "", err
	}

	var v map[string]interface{}

	dec := json.NewDecoder(strings.NewReader(body))
	if err := dec.Decode(&v); err != nil {
		return "", fmt.Errorf("invalid JSON %v", body)
	}

	if v["token"] == nil {
		return "", fmt.Errorf("Error, no token returned")
	}

	return v["token"].(string), nil
}

func (dl *Dashlane) LatestVault(login string, uki string) (map[string]interface{}, error) {
	logrus.Debug("LatestVault: ", login, uki)
	data := url.Values{}
	data.Set("login", login)
	data.Set("uki", uki)
	data.Set("timestamp", "1")
	data.Set("lock", "nolock")
	data.Set("sharingTimestamp", "0")
	data.Set("sharingSkipped", "webapp")

	body, err := PostData(LATEST_URI, data)
	if err != nil {
		return nil, err
	}

	var v map[string]interface{}

	dec := json.NewDecoder(strings.NewReader(body))
	if err := dec.Decode(&v); err != nil {
		return nil, fmt.Errorf("invalid JSON %v", body)
	}

	logrus.Debug("vault: ", v)

	if v["fullBackupFile"] == nil {
		return nil, fmt.Errorf("Error, no full backup returned")
	}

	return v, nil
}
