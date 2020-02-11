package command

import (
	"path"

	"github.com/masterzen/dashlane-cli/pkg/dashlane"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/afero"

	"github.com/alecthomas/kong"
	"github.com/sirupsen/logrus"
)

type debugFlag bool

// Context of the app
type Context struct {
	Dl *dashlane.Dashlane
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

	fs := afero.NewOsFs()
	DashlaneConfig := path.Join(dashlaneDir, "config.json")
	DashlaneVault := path.Join(dashlaneDir, "vault.json")

	context := &Context{
		Dl: dashlane.New(fs, DashlaneVault, DashlaneConfig),
	}
	fs.MkdirAll(dashlaneDir, 0700)

	cli := new(cli)
	kconfig := kong.Configuration(kong.JSON, DashlaneConfig)
	kongctx := kong.Parse(cli, kconfig)
	err = kongctx.Run(context)
	kongctx.FatalIfErrorf(err)
}

func (d debugFlag) BeforeApply() error {
	logrus.SetLevel(logrus.DebugLevel)
	return nil
}
