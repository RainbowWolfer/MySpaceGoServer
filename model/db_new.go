package model

import "errors"

type NewUsername struct {
	ID          string `json:"id"`
	Username    string `json:"username"`
	Password    string `json:"password"`
	NewUsername string `json:"new_username"`
}

func (new NewUsername) CheckValid() bool {
	
	return false
}

type NewComment struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	PostID      string `json:"post_id"`
	TextContent string `json:"text_content"`
}

func (new NewComment) CheckValid() error {
	
	return errors.New("")
}

type NewPostVote struct {
	PostID   string `json:"post_id"`
	UserID   string `json:"user_id"`
	Cancel   bool   `json:"cancel"`
	Score    int    `json:"score"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type NewCommentVote struct {
	CommentID string `json:"comment_id"`
	UserID    string `json:"user_id"`
	Cancel    bool   `json:"cancel"`
	Score     int    `json:"score"`
	Email     string `json:"email"`
	Password  string `json:"password"`
}

type NewRepost struct {
	OriginPostID   string   `json:"origin_post_id"`
	PublisherID    string   `json:"publisher_id"`
	TextContent    string   `json:"text_content"`
	PostVisibility string   `json:"post_visibility"`
	ReplyLimit     string   `json:"reply_limit"`
	Tags           []string `json:"tags"`
	Email          string   `json:"email"`
	Password       string   `json:"password"`
}

type NewCollection struct {
	TargetID string `json:"target_id"`
	Type     string `json:"type"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type NewUserFollow struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	TargetID string `json:"target_id"`
	Cancel   bool   `json:"cancel"`
}
