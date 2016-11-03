package command

import (
	"fmt"
	"runtime"

	"github.com/masterzen/dashlane-cli/version"
	"github.com/spf13/cobra"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Displays dahslane-cli version",
	Long:  ``,
	Run:   cmdVersion,
}

func init() {
	RootCmd.AddCommand(versionCmd)
}

func cmdVersion(cmd *cobra.Command, args []string) {
	fmt.Printf("dashlane-cli version: %s\n", version.GetFullVersion())
	fmt.Printf("Target OS/Arch: %s %s\n", runtime.GOOS, runtime.GOARCH)
	fmt.Printf("Built with Go Version: %s\n", runtime.Version())
}
