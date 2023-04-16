package model

type RemoveCollection struct {
	TargetID string `json:"target_id"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type DeletePost struct {
	PostID   string `json:"post_id"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type DeletePostWithOnlyID struct {
	PostID   string `json:"post_id"`
}
