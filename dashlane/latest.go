package dashlane

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	jww "github.com/spf13/jwalterweatherman"
)

const LATEST_URI = "https://www.dashlane.com/12/backup/latest"

func LatestToken(login string, code string) (string, error) {
	jww.DEBUG.Println("LatestToken: ", login, code)
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

func LatestVault(login string, uki string) (map[string]interface{}, error) {
	jww.DEBUG.Println("LatestVault: ", login, uki)
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

	jww.DEBUG.Println("vault: ", v)

	if v["fullBackupFile"] == nil {
		return nil, fmt.Errorf("Error, no full backup returned")
	}

	return v, nil
}
