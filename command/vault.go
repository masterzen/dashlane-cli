package command

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/howeyc/gopass"
	"github.com/masterzen/dashlane-cli/pkg/dashlane"
	"github.com/spf13/afero"
)

type VaultCmd struct {
	Fetch FetchCmd `cmd help:"Fetch the <login> vault from the registered computer <uki>."`
	Get   GetCmd   `cmd help:"Get a vault password."`
	List  ListCmd  `cmd help:"List the content of the vault matching pattern."`
}

type FetchCmd struct {
	Login string `required name:"username" help:"Username."`
	Uki   string `required name:"uki" help:"Uki."`
}

type GetCmd struct {
	Site string `arg required name:"site" help:"Site."`
}

type ListCmd struct {
	Pattern string `arg required name:"pattern" help:"Pattern."`
}

func (f *FetchCmd) Run(ctx *Context) error {
	vault, err := ctx.Dl.LatestVault(f.Login, f.Uki)
	if err != nil {
		return err
	}

	b, err := json.Marshal(vault)
	if err != nil {
		return err
	}

	return afero.WriteFile(ctx.Dl.Filesystem, ctx.Dl.Vault, []byte(b), 0600)
}

func (g *GetCmd) Run(ctx *Context) error {
	return nil
}

func (l *ListCmd) Run(ctx *Context) error {
	// read vault from vault dir
	b, err := afero.Exists(ctx.Dl.Filesystem, ctx.Dl.Vault)
	if err != nil {
		return err
	}

	if !b {
		return errors.New("Vault file doesn't exist, you should run `dashlane-cli vault fetch`")
	}

	// ask for password
	fmt.Print("Password: ")
	password, err := gopass.GetPasswd()
	if err != nil {
		return err
	}

	// Read the vault and parse the json
	rawFileVault, err := afero.ReadFile(ctx.Dl.Filesystem, ctx.Dl.Vault)
	if err != nil {
		return err
	}
	jsonVault, err := ctx.Dl.LoadVault(rawFileVault)
	if err != nil {
		return err
	}

	vault, err := ctx.Dl.ParseVault(jsonVault.FullBackupFile, password)
	if err != nil {
		return err
	}

	lookupVaultData(vault.List.Passwords, l.Pattern)
	lookupVaultData(vault.List.Notes, l.Pattern)
	return nil
}

func lookupVaultData(items []dashlane.VaultItem, pattern string) {
	for _, item := range items {
		for _, data := range item.Datas {
			if data.Key == "Title" && strings.Contains(data.Value, pattern) {
				fmt.Println(item.Datas)
				fmt.Println("==============")
			}
		}
	}
}
