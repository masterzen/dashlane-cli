package dashlane

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"crypto/tls"

	jww "github.com/spf13/jwalterweatherman"
)

func PostData(uri string, data url.Values) (string, error) {
	jww.DEBUG.Println("PostData: ", uri, data)
	tr := &http.Transport{
		TLSClientConfig:    &tls.Config{InsecureSkipVerify: true},
		DisableCompression: true,
		Proxy: http.ProxyFromEnvironment,
	}
	client := &http.Client{Transport: tr}
	resp, err := client.PostForm(uri, data)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return "", fmt.Errorf("Post failed impossible to parse body: %v", err)
	}

	var strBody = string(body)
	if resp.StatusCode != 200 {
		return strBody, fmt.Errorf("Post (%v) failed with status: %v, error: %v", uri, resp.StatusCode, strBody)
	}

	return strBody, nil
}
