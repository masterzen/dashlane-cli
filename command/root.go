package command

import (
	"path"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	jww "github.com/spf13/jwalterweatherman"
)

var cfgFile string

var (
	verbose     bool
	Filesystem  afero.Fs
	DashlaneDir string
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "dashlane-cli",
	Short: "Command line tool to decrypt your dashlane vault and access passwords",
	Long: `dashlane-cli is the main command, used to access the vault.

Complete documentation is available at https://github.com/masterzen/dashlane-cli.
`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		if verbose {
			jww.SetStdoutThreshold(jww.LevelDebug)
		} else {
			jww.SetStdoutThreshold(jww.LevelInfo)
		}
	},
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() error {
	return RootCmd.Execute()
}

func init() {
	InitOSFS()

	jww.SetStdoutThreshold(jww.LevelDebug)

	RootCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Run more verbosely")

	dir, err := homedir.Dir()
	if err != nil {
		panic("Unable to find homedir")
	}
	DashlaneDir = path.Join(dir, ".dashlane")
	Filesystem.MkdirAll(DashlaneDir, 0700)
}

func InitOSFS() {
	Filesystem = afero.NewOsFs()
}

func InitMemFS() {
	Filesystem = afero.NewMemMapFs()
}
