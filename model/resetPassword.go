package model

import (
	"errors"
	"rainbowwolfer/myspacegoserver/api"
)

type SendResetPasswordEmail struct {
	Email string `json:"email"`
}

type ResetPassword struct {
	Email       string `json:"email"`
	Code        string `json:"code"`
	NewPassword string `json:"new_password"`
}

func (r *ResetPassword) CheckValid() error {
	errMsg := ""
	if api.IsEmpty(&r.Email) {
		errMsg += "Missing paramter 'email'\n"
	}
	if api.IsEmpty(&r.NewPassword) {
		errMsg += "Missing paramter 'new_password'\n"
	}
	if api.IsEmpty(&errMsg) {
		return nil
	} else {
		return errors.New(errMsg)
	}
}
