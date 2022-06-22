package model

import (
	"errors"
	"rainbowwolfer/myspacegoserver/api"
)

type FlagMessage struct {
	Email           string
	Password        string
	SenderID        string
	FlagHasReceived bool
}

func (new FlagMessage) CheckValid() error {
	errMsg := ""
	if api.IsEmpty(&new.Email) {
		errMsg += "Missing paramter 'email'\n"
	}
	if api.IsEmpty(&new.Password) {
		errMsg += "Missing paramter 'password'\n"
	}
	if api.IsEmpty(&new.SenderID) {
		errMsg += "Missing paramter 'sender_id'\n"
	}
	if api.IsEmpty(&errMsg) {
		return nil
	} else {
		return errors.New(errMsg)
	}
}
