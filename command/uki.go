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
	jww "github.com/spf13/jwalterweatherman"
)

type UkiCmd struct {
	Code     CodeCmd     `cmd help:"Finalize this computer registration."`
	Register RegisterCmd `cmd help:"Register this computer under the <username> account."`
}

type RegisterCmd struct {
	Username string `arg required name:"username" help:"Username."`
	Code     string `optional name:"code" short:"c" help:"optional OTP code."`
}

type CodeCmd struct {
	Username string `arg required name:"username" help:"Username."`
	Code     string `arg required name:"code" help:"Code."`
}

func (r *RegisterCmd) Run(ctx *Context) error {
	var login = r.Username
	var code = r.Code
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
		if len(code) > 0 {
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

func (c *CodeCmd) Run(ctx *Context) error {
	var login = c.Username
	var token = c.Code
	var uki = generate()

	fmt.Printf("Registering")

	if err := dashlane.RegisterUki("dashlane-cli", login, token, uki); err != nil {
		return err
	}

	fmt.Printf("Computer registered with uki %v\n", uki)
	return nil
}
