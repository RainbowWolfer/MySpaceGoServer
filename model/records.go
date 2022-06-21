package model

type RepostRecord struct {
	PostID   string `json:"post_id"`
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Time     string `json:"time"`
	Quote    string `json:"quote"`
}

type ScoreRecord struct {
	LikeID   string `json:"like_id"`
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Time     string `json:"time"`
	Vote     int    `json:"vote"`
}
