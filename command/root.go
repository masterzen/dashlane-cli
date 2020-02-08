package command

import (
	"encoding/json"
	"path"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/afero"

	"github.com/alecthomas/kong"
	"github.com/sirupsen/logrus"
)

// UserConfig
type UserConfig struct {
	Username string `json:"username"`
	Uki      string `json:"uki"`
}

type debugFlag bool

// Context of the app
type Context struct {
	Filesystem     afero.Fs
	DashlaneVault  string
	DashlaneConfig string
}

type cli struct {
	Debug debugFlag `help:"Enable debug logging."`

	Uki     UkiCmd     `cmd help:"Manage computer registration."`
	Vault   VaultCmd   `cmd help:"Manage the vault."`
	Version VersionCmd `cmd help:"Displays dahslane-cli version."`
}

// Execute the commands
func Execute() {

	dir, err := homedir.Dir()
	if err != nil {
		logrus.WithError(err).Error("Can't get user home directory")
	}
	dashlaneDir := path.Join(dir, ".dashlane")

	context := &Context{
		Filesystem:     afero.NewOsFs(),
		DashlaneVault:  path.Join(dashlaneDir, "vault.json"),
		DashlaneConfig: path.Join(dashlaneDir, "config.json"),
	}
	context.Filesystem.MkdirAll(dashlaneDir, 0700)

	cli := new(cli)
	kconfig := kong.Configuration(kong.JSON, context.DashlaneConfig)
	kongctx := kong.Parse(cli, kconfig)
	err = kongctx.Run(context)
	kongctx.FatalIfErrorf(err)
}

func (d debugFlag) BeforeApply() error {
	logrus.SetLevel(logrus.DebugLevel)
	return nil
}

func (ctx *Context) SaveUserCreds(username string, uki string) error {
	userConfig := UserConfig{
		Username: username,
		Uki:      uki,
	}

	b, err := json.Marshal(userConfig)
	if err != nil {
		return err
	}

	return afero.WriteFile(ctx.Filesystem, ctx.DashlaneConfig, b, 0600)
}
