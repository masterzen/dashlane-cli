package command

import (
	"fmt"
	"runtime"

	"github.com/masterzen/dashlane-cli/pkg/version"
)

type VersionCmd struct{}

func (v *VersionCmd) Run(ctx *Context) error {
	fmt.Printf("dashlane-cli version: %s\n", version.GetFullVersion())
	fmt.Printf("Target OS/Arch: %s %s\n", runtime.GOOS, runtime.GOARCH)
	fmt.Printf("Built with Go Version: %s\n", runtime.Version())

	return nil
}
