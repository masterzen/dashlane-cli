package dashlane

import (
	"encoding/json"

	"github.com/spf13/afero"
)

// UserConfig
type UserConfig struct {
	Username string `json:"username"`
	Uki      string `json:"uki"`
}

type Dashlane struct {
	Filesystem afero.Fs
	Vault      string
	Config     string
}

func New(fs afero.Fs, vault string, config string) *Dashlane {
	d := new(Dashlane)
	d.Filesystem = fs
	d.Vault = vault
	d.Config = config
	return d
}

func (dl *Dashlane) SaveUserCreds(username string, uki string) error {
	userConfig := UserConfig{
		Username: username,
		Uki:      uki,
	}

	b, err := json.Marshal(userConfig)
	if err != nil {
		return err
	}

	return afero.WriteFile(dl.Filesystem, dl.Config, b, 0600)
}
