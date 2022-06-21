package model

import "database/sql"

type Post struct {
	ID                 string
	PublisherID        string
	PublishDate        string
	EditDate           string
	EditTimes          int
	TextContent        string
	Deleted            bool
	ImagesCount        int
	Tags               string
	Visibility         string
	Reply              string
	IsRepost           bool
	OriginPostID       string
	Upvotes            int
	Downvotes          int
	Comments           int
	Reposts            int
	PublisherUsername  string
	PublisherEmail     string
	PublisherProfile   string
	OriginUserID       *string
	OriginUserUsername *string
	OriginUserEmail    *string
	OriginUserProfile  *string
	OriginPublishDate  *string
	OriginEditDate     *string
	OriginEditTimes    *int
	OriginTextContent  *string
	OriginDeleted      *bool
	OriginImagesCount  *int
	OriginTags         *string
	OriginVisibility   *string
	OriginReply        *string
	OriginIsRepost     *bool
	OriginOriginPostID *string
	OriginUpvotes      int
	OriginDownvotes    int
	OriginComments     int
	OriginReposts      int
	Score              int
	Voted              int //-1(undefined) 0(downvoted) 1(upvoted)
	HasReposted        bool
	OriginScore        *int
	OriginVoted        *int
}

func ReadPost(rows *sql.Rows) (Post, error) {
	var post Post

	if err := rows.Scan(
		&post.ID,
		&post.PublisherID,
		&post.PublishDate,
		&post.EditDate,
		&post.EditTimes,
		&post.TextContent,
		&post.Deleted,
		&post.ImagesCount,
		&post.Tags,
		&post.Visibility,
		&post.Reply,
		&post.IsRepost,
		&post.OriginPostID,
		&post.Upvotes,
		&post.Downvotes,
		&post.Comments,
		&post.Reposts,
		&post.PublisherUsername,
		&post.PublisherEmail,
		&post.PublisherProfile,
		&post.OriginUserID,
		&post.OriginUserUsername,
		&post.OriginUserEmail,
		&post.OriginUserProfile,
		&post.OriginPublishDate,
		&post.OriginEditDate,
		&post.OriginEditTimes,
		&post.OriginTextContent,
		&post.OriginDeleted,
		&post.OriginImagesCount,
		&post.OriginTags,
		&post.OriginVisibility,
		&post.OriginReply,
		&post.OriginIsRepost,
		&post.OriginOriginPostID,
		&post.OriginUpvotes,
		&post.OriginDownvotes,
		&post.OriginComments,
		&post.OriginReposts,
		&post.Score,
		&post.Voted,
		&post.HasReposted,
		&post.OriginScore,
		&post.OriginVoted,
	); err != nil {
		return Post{}, err
	}
	return post, nil
}
