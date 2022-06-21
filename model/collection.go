package model

type Collection struct {
	ID                string
	UserID            string
	TargetID          string
	Type              string
	Time              string
	PublisherID       *string
	PublisherUsername *string
	TextContent       *string
	ImagesCount       *int
	IsRepost          *bool
}
