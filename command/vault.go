package command

import (
	"encoding/json"
	"errors"
	"fmt"
	"path"
	"strings"

	"github.com/howeyc/gopass"
	"github.com/masterzen/dashlane-cli/dashlane"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

var vaultCmd = &cobra.Command{
	Use:     "vault [fetch|get|list]",
	Short:   "Manage the vault",
	Long:    `dashlane-cli vault allows to get passwords and fetch the latest vault.`,
	Example: `dashlane-cli vault get www.google.com`,
}

var fetchCmd = &cobra.Command{
	Use:   "fetch <login> <uki>",
	Short: "Fetch the <login> vault from the registered computer <uki>",
	Long: `dashlane-cli vault fetch allows to get the latest vault.

`,
	Example: `dashlane-cli vault fetch myself@gmail.com blablah-webapp-blah`,
	RunE:    fetchExec,
}

var listCmd = &cobra.Command{
	Use:   "list [pattern]",
	Short: "List the content of the vault matching pattern",
	Long: `dashlane-cli vault list allows to list the content of given vault.
The vault has to be fetch with 'dashlane-cli vault fetch' first
`,
	Example: `dashlane-cli vault list www.google.com`,
	RunE:    listExec,
}

var getCmd = &cobra.Command{
	Use:   "get <site>",
	Short: "Get a vault password",
	Long: `dashlane-cli vault get allows to get the password of given vault entry.
The vault has to be fetch with 'dashlane-cli vault fetch' first
`,
	Example: `dashlane-cli vault get www.google.com`,
	RunE:    getExec,
}

func init() {
	RootCmd.AddCommand(vaultCmd)
	vaultCmd.AddCommand(fetchCmd)
	vaultCmd.AddCommand(listCmd)
	vaultCmd.AddCommand(getCmd)
}

func fetchExec(cmd *cobra.Command, args []string) error {
	vault, err := dashlane.LatestVault(args[0], args[1])
	if err != nil {
		return err
	}

	b, err := json.Marshal(vault)
	if err != nil {
		return err
	}

	return afero.WriteFile(Filesystem, path.Join(DashlaneDir, "vault.json"), []byte(b), 0600)
}

func getExec(cmd *cobra.Command, args []string) error {
	return nil
}

func listExec(cmd *cobra.Command, args []string) error {
	// read vault from vault dir
	b, err := afero.Exists(Filesystem, path.Join(DashlaneDir, "vault.json"))
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
	rawFileVault, err := afero.ReadFile(Filesystem, path.Join(DashlaneDir, "vault.json"))
	if err != nil {
		return err
	}
	jsonVault, err := dashlane.LoadVault(rawFileVault)
	if err != nil {
		return err
	}

	vault, err := dashlane.ParseVault(jsonVault.FullBackupFile, string(password))
	if err != nil {
		return err
	}

	lookupVaultData(vault.List.Passwords, args[0])
	lookupVaultData(vault.List.Notes, args[0])
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
