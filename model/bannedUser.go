package model

type BannedUserInfo struct {
	UserID   string `json:"UserID"`
	IsBanned bool   `json:"IsBanned"`
}
