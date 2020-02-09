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
	"github.com/sirupsen/logrus"
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
	Username string `required name:"username" help:"Username."`
	Code     string `arg required name:"code" help:"Code."`
}

func (r *RegisterCmd) Run(ctx *Context) error {
	var login = r.Username
	var code = r.Code
	logrus.Debug("registerExec for:", login)
	res, err := dashlane.Exist(login)
	if err != nil {
		return err
	}

	uki := ""

	switch res {
	case dashlane.EXIST_YES:
		logrus.Debug("registerExec returned: EXIST_YES")
		// ask for a token by email
		response, err := dashlane.SendToken(login)
		if err != nil {
			return err
		}
		if response != dashlane.UKI_SUCCESS {
			return fmt.Errorf("Error while requesting token: %v", response)
		}
	case dashlane.EXIST_YES_OTP_NEWDEVICE:
		logrus.Debug("registerExec returned: EXIST_YES_OTP_NEWDEVICE")
		if len(code) > 0 {
			if token, err := dashlane.LatestToken(login, code); err == nil {
				// register now
				logrus.Debug("registerExec token is: ", token)
				uki, err := generate()
				if err != nil {
					return err
				}
				logrus.Debug("registerExec uki is: ", uki)
				if err = dashlane.RegisterUki("dashlane-cli", login, token, uki); err != nil {
					return err
				}
				logrus.Info("Computer registered with uki: ", uki)
			} else {
				return err
			}
		} else {
			return errors.New("This account requires an OTP code to be provided with --code")
		}
	default:
		return fmt.Errorf("There is no account for this login")
	}

	ctx.SaveUserCreds(login, uki)
	return nil
}

func getMD5Hash(text string) string {
	hasher := md5.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}

func generate() (string, error) {
	r, err := rand.Int(rand.Reader, big.NewInt(268435456))
	if err != nil {
		return "", err
	}

	var time = fmt.Sprintf("%d", time.Now().Unix())
	var text = runtime.GOOS + runtime.GOARCH + time + r.Text(16)
	var hashed = getMD5Hash(text)

	return hashed + "-webaccess-" + time, nil
}

func (c *CodeCmd) Run(ctx *Context) error {
	var login = c.Username
	var token = c.Code
	uki, err := generate()
	if err != nil {
		return err
	}

	fmt.Printf("Registering")

	if err = dashlane.RegisterUki("dashlane-cli", login, token, uki); err != nil {
		return err
	}

	fmt.Printf("Computer registered with uki %v\n", uki)
	ctx.SaveUserCreds(login, uki)
	return nil
}
