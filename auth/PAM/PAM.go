package pam

import (
	"errors"
	"log"

	"github.com/msteinert/pam"
)

func PAMAuth(username string, password string) {
	t, err := pam.StartFunc("check_user", username, func(s pam.Style, msg string) (string, error) {
		switch s {
		case pam.PromptEchoOff:
			return password, nil
		case pam.PromptEchoOn, pam.ErrorMsg, pam.TextInfo:
			return "", nil
		}
		return "", errors.New("Unrecognized PAM message style")
	})

	if err != nil {
		t.Authenticate(0)
	} else {
		log.Fatal("Authentication failed")
	}
}
