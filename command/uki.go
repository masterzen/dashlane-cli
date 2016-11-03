package command

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"runtime"
	"time"

	"github.com/masterzen/dashlane-cli/dashlane"
	"github.com/spf13/cobra"
	jww "github.com/spf13/jwalterweatherman"
)

// ukiCmd represents the master uki command
var ukiCmd = &cobra.Command{
	Use:     "uki [flags] [register|code]",
	Aliases: []string{"computer"},
	Short:   "Manage computer registration",
	Long:    `dashlane-cli uki allows to register a new computer to dashlane.`,
	Example: `dashlane-cli uki register`,
}

var registerCmd = &cobra.Command{
	Use:   "register <username>",
	Short: "Register this computer under the <username> account",
	Long: `dashlane-cli uki register allows to register a new computer to dashlane.

`,
	Example: `dashlane-cli uki register myself@gmail.com`,
	RunE:    registerExec,
}

var codeCmd = &cobra.Command{
	Use:   "code <code>",
	Short: "Finalize this computer registration",
	Long: `dashlane-cli uki code allows to enter the confirmation code.

`,
	Example: `dashlane-cli uki code 0f304f04504040f40`,
	RunE:    codeExec,
}

func init() {
	RootCmd.AddCommand(ukiCmd)
	ukiCmd.AddCommand(registerCmd)
	ukiCmd.AddCommand(codeCmd)

	registerCmd.Flags().StringP("code", "c", "", "optional OTP code")
}

func registerExec(cmd *cobra.Command, args []string) error {
	var login = args[0]
	jww.DEBUG.Println("registerExec for:", login)
	res, err := dashlane.Exist(login)
	if err != nil {
		return err
	}

	switch res {
	case dashlane.EXIST_YES:
		jww.DEBUG.Println("registerExec returned: EXIST_YES")
		// ask for a token by email
		response, err := dashlane.SendToken(login)
		if err != nil {
			return err
		}
		if response != dashlane.UKI_SUCCESS {
			return fmt.Errorf("Error while requesting token: %v", response)
		}
	case dashlane.EXIST_YES_OTP_NEWDEVICE:
		jww.DEBUG.Println("registerExec returned: EXIST_YES_OTP_NEWDEVICE")
		if cmd.Flags().Changed("code") {
			// retrieve the server side token
			code, _ := cmd.Flags().GetString("code")
			if token, err := dashlane.LatestToken(login, code); err == nil {
				// register now
				jww.DEBUG.Println("registerExec token is: ", token)
				uki := generate()
				jww.DEBUG.Println("registerExec uki is: ", uki)
				if err = dashlane.RegisterUki("dashlane-cli", login, token, uki); err != nil {
					return err
				} else {
					jww.INFO.Println("Computer registered with uki: ", uki)
				}
			} else {
				return err
			}
		} else {
			return errors.New("This account requires an OTP code to be provided with --code")
		}
	default:
		return fmt.Errorf("There is no account for this login")
	}

	return nil
}

func getMD5Hash(text string) string {
	hasher := md5.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}

func generate() string {
	r, err := rand.Int(rand.Reader, big.NewInt(268435456))
	if err != nil {
		panic(err)
	}

	var time = fmt.Sprintf("%d", time.Now().Unix())
	var text = runtime.GOOS + runtime.GOARCH + time + r.Text(16)
	var hashed = getMD5Hash(text)

	return hashed + "-webaccess-" + time
}

func codeExec(cmd *cobra.Command, args []string) error {
	var login = args[0]
	var token = args[1]
	var uki = generate()

	fmt.Printf("Registering")

	if err := dashlane.RegisterUki("dashlane-cli", login, token, uki); err != nil {
		return err
	}

	fmt.Printf("Computer registered with uki %v\n", uki)
	return nil
}
