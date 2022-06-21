package model

type Comment struct {
	ID          string
	UserID      string
	PostID      string
	TextContent string
	DateTime    string
	Username    string
	Email       string
	Profile     string
	Upvotes     int
	Downvotes   int
	Voted       int
}
