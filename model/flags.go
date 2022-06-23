package model

import (
	"errors"
	"rainbowwolfer/myspacegoserver/api"
)

type FlagMessage struct {
	Email           string `json:"email"`
	Password        string `json:"password"`
	SenderID        string `json:"sender_id"`
	FlagHasReceived bool   `json:"flag_has_received"`
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
