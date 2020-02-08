package command

import (
	"path"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/afero"

	"github.com/alecthomas/kong"
	"github.com/sirupsen/logrus"
)

type debugFlag bool

// Context of the app
type Context struct {
	Filesystem     afero.Fs
	DashlaneDir    string
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
	cli := new(cli)
	kongctx := kong.Parse(cli)

	dir, err := homedir.Dir()
	if err != nil {
		kongctx.FatalIfErrorf(err)
	}

	context := &Context{
		Filesystem:     afero.NewOsFs(),
		DashlaneDir:    path.Join(dir, ".dashlane"),
		DashlaneVault:  path.Join(dir, ".dashlane", "vault.json"),
		DashlaneConfig: path.Join(dir, ".dashlane", "config.json"),
	}

	context.Filesystem.MkdirAll(context.DashlaneDir, 0700)

	err = kongctx.Run(context)
	kongctx.FatalIfErrorf(err)
}

func (d debugFlag) BeforeApply() error {
	logrus.SetLevel(logrus.DebugLevel)
	return nil
}
