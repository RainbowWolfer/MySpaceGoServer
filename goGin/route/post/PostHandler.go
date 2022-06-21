package post

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"rainbowwolfer/myspacegoserver/api"
	"rainbowwolfer/myspacegoserver/goGin/config"
	"rainbowwolfer/myspacegoserver/goGin/ginTools"
	"rainbowwolfer/myspacegoserver/model"
	"strconv"
	"strings"
)

func PostHandler() (string, ginTools.RouteMap) {
	routeMap := ginTools.NewRouteMap()

	routeMap.AddRoute(ginTools.Route{
		Name:   "/", // post_post_get)
		Fun:    post_post_get,
		Method: http.MethodGet,
	})

	//post/get
	routeMap.AddRoute(ginTools.Route{
		Name:   "/delete", // deletePost_post)
		Fun:    deletePost_post,
		Method: http.MethodPost,
	})

	/**
	getUserByUserID_get get
	*/
	routeMap.AddRoute(ginTools.Route{
		Name:   "/user",
		Fun:    getUserByUserID_get,
		Method: http.MethodGet,
	})

	/**
	getPostByID_get
	*/
	routeMap.AddRoute(ginTools.Route{
		Name:   "/id",
		Fun:    getPostByID_get,
		Method: http.MethodGet,
	}) //get
	routeMap.AddRoute(ginTools.Route{
		Name:   "/images", // getPostImage_get
		Fun:    getPostImage_get,
		Method: http.MethodGet,
	}) //get

	routeMap.AddRoute(ginTools.Route{
		Name:   "/comment", // comment_post_get
		Fun:    comment_post_get,
		Method: http.MethodGet,
	}) //post/get

	// post/comment/vote
	routeMap.AddRoute(ginTools.Route{
		Name:   "/comment/vote", //voteComment_post
		Fun:    voteComment_post,
		Method: http.MethodPost,
	}) //get

	// /post/comments
	routeMap.AddRoute(ginTools.Route{
		Name:   "/comments", // getPostComments_get
		Fun:    getPostComments_get,
		Method: http.MethodGet,
	}) //get

	// /post/vote  post
	routeMap.AddRoute(ginTools.Route{
		Name:   "/vote", // votePost_post
		Fun:    votePost_post,
		Method: http.MethodPost,
	})

	// post/repost repost_post
	routeMap.AddRoute(ginTools.Route{
		Name:   "/repost", // repost_post
		Fun:    repost_post,
		Method: http.MethodPost,
	})

	//removeCollection_post post
	routeMap.AddRoute(ginTools.Route{
		Name:   "/repostRecords", // getRepostRecords_get
		Fun:    getRepostRecords_get,
		Method: http.MethodGet,
	})

	//getCollections_get get
	routeMap.AddRoute(ginTools.Route{
		Name:   "/scoreRecords", // getScoreRecords_get
		Fun:    getScoreRecords_get,
		Method: http.MethodGet,
	})

	// 可以读取url中的参数
	return "/post", *routeMap
}
func post_post_get(context *gin.Context) {
	r := context.Request
	w := context.Writer
	if err := api.CheckRequestMethod(r, "post"); err == nil {
		//post a post
		// postPost(w, r)
		if api.CheckRequestMethodReturn(w, r, "post") {
			return
		}

		if err := r.ParseMultipartForm(32 << 20); err != nil {
			api.HttpError(w, err.Error(), http.StatusBadRequest)
			return
		}

		// println(r.MultipartForm.File)
		// println(r.MultipartForm.Value)

		content := r.MultipartForm.Value["content"][0]
		publisherID := r.MultipartForm.Value["publisher_id"][0]
		postVisibility := r.MultipartForm.Value["post_visibility"][0]
		replyVisibility := r.MultipartForm.Value["reply_visibility"][0]
		tags := strings.Split(r.MultipartForm.Value["tags"][0], "&#10;")
		images := r.MultipartForm.File["post_images"]
		// println("content is : " + content)
		// println(publisherID)
		// println(postVisibility)
		// println(replyVisibility)
		// println(strings.Join(tags, ","))

		// println(images)
		// println(len(images))
		for _, header := range images {
			if header.Size > config.MAX_UPLOAD_SIZE {
				api.HttpError(w, fmt.Sprintf("The uploaded image is too big: %s. Please use an image less than 1MB in size", header.Filename), http.StatusBadRequest)
				return
			}
		}

		database, err := api.GetDatabase()
		if err != nil {
			api.HttpError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer database.Close()

		sql := fmt.Sprintf("select u_id from users where u_id = '%s'", publisherID)
		rows, err := database.Query(sql)
		if err != nil {
			api.HttpError(w, "Query Error with :"+sql+"\n"+err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		if !rows.Next() {
			api.HttpError(w, "Cannot find the publishder ID", http.StatusBadRequest)
			return
		}

		visibility := ""
		if postVisibility == "0" {
			visibility = "all"
		} else if postVisibility == "1" {
			visibility = "follower"
		} else if postVisibility == "2" {
			visibility = "none"
		}

		reply := ""
		if replyVisibility == "0" {
			reply = "all"
		} else if replyVisibility == "1" {
			reply = "follower"
		} else if replyVisibility == "2" {
			reply = "none"
		}

		joinedTags := ""
		// println(tags)
		if len(tags) != 0 {
			tags = api.DeleteEmpty(tags)
			joinedTags = strings.Join(tags, ",")
			sql = fmt.Sprintf("call add_tags('%s')", joinedTags)
			_, err := database.Exec(sql)
			if err != nil {
				api.HttpError(w, "add tags go wrong"+err.Error(), http.StatusInternalServerError)
				return
			}
		}
		// println(joinedTags)

		sql = fmt.Sprintf("INSERT INTO posts (p_publisher_id, p_publish_date, p_edit_date, p_text_content, p_visibility, p_reply, p_images_count, p_tags) VALUES ('%s',NOW(),NOW(),'%s','%s','%s','%d','%s')", publisherID, content, visibility, reply, len(images), joinedTags)
		result, err := database.Exec(sql)
		if err != nil {
			api.HttpError(w, "insert post error"+err.Error(), http.StatusInternalServerError)
			return
		}

		id, err := result.LastInsertId()
		if err != nil {
			api.HttpError(w, "Cannot get last inserted id"+err.Error(), http.StatusInternalServerError)
			return
		}

		for i, header := range images {
			file, err := header.Open()
			if err != nil {
				api.HttpError(w, err.Error(), http.StatusInternalServerError)
				return
			}
			defer file.Close()
			buff := make([]byte, 512)
			// print("before: ")
			// println(buff)
			_, err = file.Read(buff)
			if err != nil {
				api.HttpError(w, err.Error(), http.StatusInternalServerError)
				return
			}
			// print("after: ")
			// println(buff)
			filetype := http.DetectContentType(buff)
			// println("file type is: " + filetype)
			// println(filepath.Ext(header.Filename))
			if filetype != "image/jpeg" && filetype != "image/jpg" && filetype != "image/png" {
				api.HttpError(w, "The provided file format is not allowed. Please upload a JPEG(JPG) or PNG image", http.StatusBadRequest)
				continue
			}
			_, err = file.Seek(0, io.SeekStart)
			if err != nil {
				api.HttpError(w, err.Error(), http.StatusInternalServerError)
				return
			}
			ext := filepath.Ext(header.Filename)
			println(header.Filename + "_" + ext)
			err = os.MkdirAll("./uploads/posts", os.ModePerm)
			if err != nil {
				api.HttpError(w, err.Error(), http.StatusInternalServerError)
				return
			}
			f, err := os.Create(fmt.Sprintf("./uploads/posts/post_%d_%d", id, i))
			if err != nil {
				api.HttpError(w, err.Error(), http.StatusBadRequest)
				return
			}
			defer f.Close()

			pr := &api.Progress{
				TotalSize: header.Size,
			}

			_, err = io.Copy(f, io.TeeReader(file, pr))
			if err != nil {
				api.HttpError(w, err.Error(), http.StatusBadRequest)
				return
			}

		}

		fmt.Printf("r.MultipartForm.Value: %v\n", r.MultipartForm.Value)
		fmt.Printf("r.MultipartForm.File: %v\n", r.MultipartForm.File)

	} else if err := api.CheckRequestMethod(r, "get"); err == nil {
		query := r.URL.Query()

		if api.CheckMissingParamters(w, query, true, "posts_type", "offset") {
			return
		}

		posts_type := query["posts_type"][0]
		offset := query["offset"][0]

		limit := 10
		if query.Has("limit") {
			limit, err = strconv.Atoi(query["limit"][0])
			if err != nil {
				api.HttpError(w, "limit is not a valid number"+err.Error(), http.StatusBadRequest)
				return
			}
		}

		database, err := api.GetDatabase()
		if err != nil {
			api.HttpError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer database.Close()

		user_id := -1
		if query.Has("email") && query.Has("password") {
			email := query["email"][0]
			password := query["password"][0]
			if !api.IsEmpty(&email) && !api.IsEmpty(&password) {
				_user_id, err := model.GetUserID(database, email, password)
				if err != nil {
					api.HttpError(w, err.Error(), http.StatusBadRequest)
				}
				user_id = _user_id
				// sql_user := fmt.Sprintf("select u_id from users where u_email = '%s' and u_password = '%s'", email, password)
				// rows, err := database.Query(sql_user)
				// if err != nil {
				// 	api.HttpError(w, "query user error"+err.Error(), http.StatusInternalServerError)
				// 	return
				// }
				// defer rows.Close()
				// if !rows.Next() {
				// 	api.HttpError(w, "Cannot find user", http.StatusBadRequest)
				// 	return
				// }

				// if err = rows.Scan(&user_id); err != nil {
				// 	api.HttpError(w, err.Error(), http.StatusInternalServerError)
				// 	return
				// }
			}
		}

		var sql string

		if strings.EqualFold(posts_type, "All") { //time based
			sql = fmt.Sprintf("CALL GetPostsByTime(%d,%s,%d);", user_id, offset, limit)
		} else if strings.EqualFold(posts_type, "Hot") {
			sql = fmt.Sprintf("CALL GetPostsByScore(%d,%s,%d)", user_id, offset, limit)
		} else if strings.EqualFold(posts_type, "Random") {
			if !query.Has("seed") {
				api.HttpError(w, "require parameter 'seed'", http.StatusBadRequest)
				return
			}
			seed := query["seed"][0]
			sql = fmt.Sprintf("CALL GetPostsByRandom(%d,%s,%d,%s)", user_id, offset, limit, seed)
		} else if strings.EqualFold(posts_type, "Following") {
			if user_id == 0 || user_id == -1 {
				api.HttpError(w, "user not found", http.StatusBadRequest)
				return
			}
			sql = fmt.Sprintf("CALL GetPostsByFollow(%d,%s,%d)", user_id, offset, limit)
		} else {
			api.HttpError(w, "Cannot find type of :"+posts_type, http.StatusBadRequest)
			return
		}

		println(sql)
		rows, err := database.Query(sql)
		if err != nil {
			api.HttpError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var posts []model.Post
		for rows.Next() {
			post, err := model.ReadPost(rows)
			if err != nil {
				println("skipping row" + err.Error())
				continue
			}
			posts = append(posts, post)
			// if post.HasReposted {
			// 	fmt.Printf("%+v\n\n\n", post)
			// }
		}

		if err = rows.Err(); err != nil {
			api.HttpError(w, "Databse Rows Error", http.StatusInternalServerError)
		}

		json := api.ToJson(posts)

		println(len(posts))
		fmt.Fprintln(w, json)
	} else {
		api.HttpError(w, "Only get or post method is allowed", http.StatusMethodNotAllowed)
	}
}

/*type loginAuth struct {
	username string
	password string
}

func LoginAuth(username, password string) smtp.Auth {
	return &loginAuth{username, password}
}
func (a *loginAuth) Start(server *smtp.ServerInfo) (string, []byte, error) {
	return "LOGIN", []byte(a.username), nil
}
func (a *loginAuth) Next(fromServer []byte, more bool) ([]byte, error) {
	if more {
		switch string(fromServer) {
		case "Username:":
			return []byte(a.username), nil
		case "Password:":
			return []byte(a.password), nil
		default:
			return nil, errors.New("unkown from server")
		}
	}
	return nil, nil
}
*/
func getPostImage_get(context *gin.Context) {
	r := context.Request
	w := context.Writer
	if api.CheckRequestMethodReturn(w, r, "get") {
		return
	}

	query := r.URL.Query()
	if api.CheckMissingParamters(w, query, true, "id", "index") {
		return
	}

	id := query["id"][0]
	index_str := query["index"][0]

	index, err := strconv.ParseInt(index_str, 0, 8)
	if err != nil {
		api.HttpError(w, fmt.Sprintf("%s is not a valid number", index_str), http.StatusBadRequest)
		return
	}

	if index < 0 || index > 10 {
		api.HttpError(w, fmt.Sprintf("%d is not within range", index), http.StatusBadRequest)
		return
	}

	path := fmt.Sprintf("./uploads/posts/post_%s_%d", id, index)

	bytes := api.GetImage(path)
	if bytes == nil {
		api.HttpError(w, "Cannot find file", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(bytes)
}
func comment_post_get(context *gin.Context) {
	r := context.Request
	w := context.Writer
	if err := api.CheckRequestMethod(r, "post"); err == nil {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			api.HttpError(w, "No body was found : "+err.Error(), http.StatusBadRequest)
		}
		println(string(body))

		var obj model.NewComment
		err = json.Unmarshal(body, &obj)
		if err != nil {
			api.HttpError(w, "json unmarshall error:"+err.Error(), http.StatusBadRequest)
			return
		}
		obj.CheckValid()

		errorMessage := ""
		if api.IsEmpty(&obj.Email) {
			errorMessage += "Missing paramter 'email'\n"
		}
		if api.IsEmpty(&obj.Password) {
			errorMessage += "Missing paramter 'password'\n"
		}
		if api.IsEmpty(&obj.PostID) {
			errorMessage += "Missing paramter 'post_id'\n"
		}
		if api.IsEmpty(&obj.TextContent) {
			errorMessage += "Missing paramter 'text_content'\n"
		}
		if !api.IsEmpty(&errorMessage) {
			api.HttpError(w, errorMessage, http.StatusBadRequest)
			return
		}

		database, err := api.GetDatabase()
		if err != nil {
			api.HttpError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer database.Close()

		//check user valid
		sql := fmt.Sprintf("select u_id, u_username, u_email, u_profileDescription from users where u_email = '%s' and u_password = '%s';", obj.Email, obj.Password)

		rows, err := database.Query(sql)
		if err != nil {
			api.HttpError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		if !rows.Next() {
			api.HttpError(w, "user not found", http.StatusBadRequest)
			return
		}
		var user struct {
			UserID   int
			Username string
			Email    string
			Profile  string
		}
		err = rows.Scan(&user.UserID, &user.Username, &user.Email, &user.Profile)
		if err != nil {
			api.HttpError(w, "user id scan error"+err.Error(), http.StatusInternalServerError)
			return
		}

		//add comment
		sql = fmt.Sprintf("INSERT INTO comments (c_id_user, c_id_post, c_text_content, c_datetime) VALUES ('%d','%s','%s',NOW());", user.UserID, obj.PostID, obj.TextContent)

		result, err := database.Exec(sql)

		if err != nil {
			api.HttpError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		last_id, err := result.LastInsertId()
		if err != nil {
			api.HttpError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		//actually, i can just read it from database
		fmt.Fprintln(w, api.ToJson(model.Comment{
			ID:          fmt.Sprintf("%d", last_id),
			UserID:      fmt.Sprintf("%d", user.UserID),
			PostID:      obj.PostID,
			TextContent: obj.TextContent,
			DateTime:    api.Now(),
			Username:    user.Username,
			Email:       user.Email,
			Profile:     user.Profile,
			Voted:       -1,
			Upvotes:     0,
			Downvotes:   0,
		}))

	} else if err := api.CheckRequestMethod(r, "get"); err == nil {
		query := r.URL.Query()
		if api.CheckMissingParamters(w, query, true, "id") {
			return
		}
		id := query["id"][0]

		database, err := api.GetDatabase()
		if err != nil {
			api.HttpError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer database.Close()

		sql := fmt.Sprintf("select * from comments where c_id = '%s'", id)

		rows, err := database.Query(sql)
		if err != nil {
			api.HttpError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		if !rows.Next() {
			api.HttpError(w, "not comment found", http.StatusBadRequest)
			return
		}

		var comment model.Comment
		err = rows.Scan(
			&comment.ID,
			&comment.UserID,
			&comment.PostID,
			&comment.TextContent,
			&comment.DateTime,
			&comment.Upvotes,
			&comment.Downvotes,
		)

		if err != nil {
			api.HttpError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		fmt.Fprintln(w, api.ToJson(comment))
	} else {
		api.HttpError(w, "Only get or post method is allowed", http.StatusMethodNotAllowed)
	}
}
func getPostComments_get(context *gin.Context) {
	r := context.Request
	w := context.Writer
	if api.CheckRequestMethodReturn(w, r, "get") {
		return
	}
	query := r.URL.Query()
	if api.CheckMissingParamters(w, query, true, "post_id", "offset") {
		return
	}

	postID := query["post_id"][0]
	offset := query["offset"][0]

	database, err := api.GetDatabase()
	if err != nil {
		api.HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer database.Close()

	limit := 15
	if query.Has("limit") {
		limit, err = strconv.Atoi(query["limit"][0])
		if err != nil {
			api.HttpError(w, "limit is not a valid number"+err.Error(), http.StatusBadRequest)
			return
		}
	}

	user_id := -1
	if query.Has("email") && query.Has("password") {
		email := query["email"][0]
		password := query["password"][0]
		if !api.IsEmpty(&email) && !api.IsEmpty(&password) {
			sql_user := fmt.Sprintf("select u_id from users where u_email = '%s' and u_password = '%s'", email, password)
			rows, err := database.Query(sql_user)
			if err != nil {
				api.HttpError(w, "query user error"+err.Error(), http.StatusInternalServerError)
				return
			}
			defer rows.Close()
			if !rows.Next() {
				api.HttpError(w, "Cannot find user", http.StatusBadRequest)
				return
			}

			if err = rows.Scan(&user_id); err != nil {
				api.HttpError(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
	}

	sql := fmt.Sprintf("call GetCommentsByTime(%d,'%s','%s',%d);", user_id, postID, offset, limit)

	rows, err := database.Query(sql)
	if err != nil {
		api.HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var list []model.Comment

	for rows.Next() {
		var item model.Comment
		err = rows.Scan(
			&item.ID,
			&item.UserID,
			&item.PostID,
			&item.TextContent,
			&item.DateTime,
			&item.Username,
			&item.Email,
			&item.Profile,
			&item.Upvotes,
			&item.Downvotes,
			&item.Voted,
		)

		if err != nil {
			api.HttpError(w, err.Error(), http.StatusInternalServerError)
			continue
		}
		list = append(list, item)
	}

	fmt.Fprintln(w, api.ToJson(list))
}
func votePost_post(context *gin.Context) {
	r := context.Request
	w := context.Writer
	if api.CheckRequestMethodReturn(w, r, "post") {
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		api.HttpError(w, "No body was found : "+err.Error(), http.StatusBadRequest)
	}

	var obj model.NewPostVote
	err = json.Unmarshal(body, &obj)
	if err != nil {
		api.HttpError(w, "json unmarshall error:"+err.Error(), http.StatusBadRequest)
		return
	}

	errorMessage := ""
	if api.IsEmpty(&obj.PostID) {
		errorMessage += "Missing paramter 'post_id'\n"
	}
	if api.IsEmpty(&obj.UserID) {
		errorMessage += "Missing paramter 'user_id'\n"
	}
	if api.IsEmpty(&obj.Email) {
		errorMessage += "Missing paramter 'email'\n"
	}
	if api.IsEmpty(&obj.Password) {
		errorMessage += "Missing paramter 'password'\n"
	}
	if !api.IsEmpty(&errorMessage) {
		api.HttpError(w, errorMessage, http.StatusBadRequest)
		return
	}

	database, err := api.GetDatabase()
	if err != nil {
		api.HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer database.Close()

	userID, err := model.GetUserID(database, obj.Email, obj.Password)
	if err != nil {
		api.HttpError(w, "During User Check "+err.Error(), http.StatusInternalServerError)
		return
	} else if userID == -1 {
		api.HttpError(w, "No Uesr Found", http.StatusBadRequest)
		return
	}

	sql := fmt.Sprintf("CALL VotePost(%s,%s,%t,%d)", obj.UserID, obj.PostID, obj.Cancel, obj.Score)

	rows, err := database.Query(sql)
	if err != nil {
		api.HttpError(w, "During SQL : "+sql+" : "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var line string
	rows.Scan(&line)

	fmt.Fprint(w, line)
}
func repost_post(context *gin.Context) {
	r := context.Request
	w := context.Writer
	if api.CheckRequestMethodReturn(w, r, "post") {
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		api.HttpError(w, "No body was found : "+err.Error(), http.StatusBadRequest)
	}

	var obj model.NewRepost
	err = json.Unmarshal(body, &obj)
	if err != nil {
		api.HttpError(w, "json unmarshall error:"+err.Error(), http.StatusBadRequest)
		return
	}

	println(fmt.Sprintf("%v", obj))

	errorMessage := ""
	if api.IsEmpty(&obj.OriginPostID) {
		errorMessage += "Missing paramter 'origin_post_id'\n"
	}
	if api.IsEmpty(&obj.PostVisibility) {
		errorMessage += "Missing paramter 'post_visibility'\n"
	}
	if api.IsEmpty(&obj.PublisherID) {
		errorMessage += "Missing paramter 'publisher_id'\n"
	}
	if api.IsEmpty(&obj.ReplyLimit) {
		errorMessage += "Missing paramter 'reply_limit'\n"
	}
	if api.IsEmpty(&obj.TextContent) {
		errorMessage += "Missing paramter 'text_content'\n"
	}
	if api.IsEmpty(&obj.Email) {
		errorMessage += "Missing paramter 'email'\n"
	}
	if api.IsEmpty(&obj.Password) {
		errorMessage += "Missing paramter 'password'\n"
	}
	if !api.IsEmpty(&errorMessage) {
		api.HttpError(w, errorMessage, http.StatusBadRequest)
		return
	}

	database, err := api.GetDatabase()
	if err != nil {
		api.HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer database.Close()

	//check user
	sql := fmt.Sprintf("select u_id from users where u_email = '%s' and u_password = '%s'", obj.Email, obj.Password)
	rows, err := database.Query(sql)
	if err != nil {
		api.HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	if !rows.Next() {
		api.HttpError(w, "user not found", http.StatusBadRequest)
		return
	}

	joinedTags := ""
	println(obj.Tags)
	if len(obj.Tags) != 0 {
		obj.Tags = api.DeleteEmpty(obj.Tags)
		joinedTags = strings.Join(obj.Tags, ",")
		sql = fmt.Sprintf("call add_tags('%s')", joinedTags)
		_, err := database.Exec(sql)
		if err != nil {
			api.HttpError(w, "add tags go wrong"+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	sql = fmt.Sprintf("INSERT INTO posts (p_publisher_id, p_publish_date, p_edit_date, p_text_content, p_visibility, p_reply, p_images_count, p_tags, p_is_repost, p_id_origin_post) VALUES ('%s',NOW(),NOW(),'%s','%s','%s',0,'%s',TRUE,'%s')", obj.PublisherID, obj.TextContent, obj.PostVisibility, obj.ReplyLimit, obj.Tags, obj.OriginPostID)
	println(sql)

	_, err = database.Exec(sql)
	if err != nil {
		api.HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

}
func voteComment_post(context *gin.Context) {
	r := context.Request
	w := context.Writer
	if api.CheckRequestMethodReturn(w, r, "post") {
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		api.HttpError(w, "No body was found : "+err.Error(), http.StatusBadRequest)
	}

	var obj model.NewCommentVote
	err = json.Unmarshal(body, &obj)
	if err != nil {
		api.HttpError(w, "json unmarshall error:"+err.Error(), http.StatusBadRequest)
		return
	}

	errorMessage := ""
	if api.IsEmpty(&obj.CommentID) {
		errorMessage += "Missing paramter 'email'\n"
	}
	if api.IsEmpty(&obj.UserID) {
		errorMessage += "Missing paramter 'email'\n"
	}
	if api.IsEmpty(&obj.Email) {
		errorMessage += "Missing paramter 'email'\n"
	}
	if api.IsEmpty(&obj.Password) {
		errorMessage += "Missing paramter 'password'\n"
	}
	if !api.IsEmpty(&errorMessage) {
		api.HttpError(w, errorMessage, http.StatusBadRequest)
		return
	}

	database, err := api.GetDatabase()
	if err != nil {
		api.HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer database.Close()

	userID, err := model.GetUserID(database, obj.Email, obj.Password)
	if err != nil {
		api.HttpError(w, "During User Check "+err.Error(), http.StatusInternalServerError)
		return
	} else if userID == -1 {
		api.HttpError(w, "No Uesr Found", http.StatusBadRequest)
		return
	}

	sql := fmt.Sprintf("CALL VoteComment(%d, '%s', %t, %d);", userID, obj.CommentID, obj.Cancel, obj.Score)

	rows, err := database.Query(sql)
	if err != nil {
		api.HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var line string
	rows.Scan(&line)

	fmt.Fprint(w, line)
}
func getRepostRecords_get(context *gin.Context) {
	r := context.Request
	w := context.Writer
	if api.CheckRequestMethodReturn(w, r, "get") {
		return
	}

	query := r.URL.Query()
	if api.CheckMissingParamters(w, query, true, "post_id", "offset") {
		return
	}

	var err error

	postID := query["post_id"][0]
	offset := query["offset"][0]

	limit := 15
	if query.Has("limit") {
		limit, err = strconv.Atoi(query["limit"][0])
		if err != nil {
			api.HttpError(w, "limit is not a valid number"+err.Error(), http.StatusBadRequest)
			return
		}
	}

	database, err := api.GetDatabase()
	if err != nil {
		api.HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer database.Close()

	sql := fmt.Sprintf("CALL GetRepostRecords('%s', '%s', %d);", postID, offset, limit)

	rows, err := database.Query(sql)
	if err != nil {
		api.HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var records []model.RepostRecord
	for rows.Next() {
		var record model.RepostRecord
		err = rows.Scan(&record.PostID, &record.UserID, &record.Username, &record.Time, &record.Quote)
		if err != nil {
			api.HttpError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		records = append(records, record)
	}

	fmt.Fprint(w, api.ToJson(records))
}
func getScoreRecords_get(context *gin.Context) {
	r := context.Request
	w := context.Writer
	if api.CheckRequestMethodReturn(w, r, "get") {
		return
	}

	query := r.URL.Query()
	if api.CheckMissingParamters(w, query, true, "post_id", "offset") {
		return
	}

	var err error

	postID := query["post_id"][0]
	offset := query["offset"][0]

	limit := 15
	if query.Has("limit") {
		limit, err = strconv.Atoi(query["limit"][0])
		if err != nil {
			api.HttpError(w, "limit is not a valid number"+err.Error(), http.StatusBadRequest)
			return
		}
	}

	database, err := api.GetDatabase()
	if err != nil {
		api.HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer database.Close()

	sql := fmt.Sprintf("CALL GetScoreRecords('%s', '%s', %d);", postID, offset, limit)

	rows, err := database.Query(sql)
	if err != nil {
		api.HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var records []model.ScoreRecord
	for rows.Next() {
		var record model.ScoreRecord
		err = rows.Scan(&record.LikeID, &record.UserID, &record.Username, &record.Time, &record.Vote)
		if err != nil {
			api.HttpError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		records = append(records, record)
	}

	fmt.Fprint(w, api.ToJson(records))
}
func getPostByID_get(context *gin.Context) {
	r := context.Request
	w := context.Writer
	if api.CheckRequestMethodReturn(w, r, "get") {
		return
	}
	query := r.URL.Query()
	if api.CheckMissingParamters(w, query, true, "post_id", "email", "password") {
		return
	}

	postID := query["post_id"][0]
	email := query["email"][0]
	password := query["password"][0]

	database, err := api.GetDatabase()
	if err != nil {
		api.HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer database.Close()

	sql := fmt.Sprintf("call GetPostByID('%s','%s','%s');", postID, email, password)

	rows, err := database.Query(sql)
	if err != nil {
		api.HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	if !rows.Next() {
		api.HttpError(w, "no post found", http.StatusBadRequest)
		return
	}

	post, err := model.ReadPost(rows)
	if err != nil {
		api.HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprint(w, api.ToJson(post))

	println(api.ToJson(post))
}
func getUserByUserID_get(context *gin.Context) {
	r := context.Request
	w := context.Writer
	if api.CheckRequestMethodReturn(w, r, "get") {
		return
	}
	query := r.URL.Query()
	if api.CheckMissingParamters(w, query, true, "email", "password", "target_id", "offset", "limit") {
		return
	}

	email := query["email"][0]
	password := query["password"][0]
	target_id := query["target_id"][0]
	offset := query["offset"][0]
	limit := query["limit"][0]

	database, err := api.GetDatabase()
	if err != nil {
		api.HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer database.Close()

	user_id := -1
	if !api.IsEmpty(&email) && !api.IsEmpty(&password) {
		_user_id, err := model.GetUserID(database, email, password)
		if err != nil {
			api.HttpError(w, err.Error(), http.StatusBadRequest)
		}
		user_id = _user_id
	}

	sql := fmt.Sprintf("CALL GetPostsByTargetID(%d,%s,%s,%s);", user_id, target_id, offset, limit)
	rows, err := database.Query(sql)
	if err != nil {
		api.HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var list []model.Post
	for rows.Next() {
		item, err := model.ReadPost(rows)
		if err != nil {
			api.HttpError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		list = append(list, item)
	}

	fmt.Fprint(w, api.ToJson(list))
}
func deletePost_post(context *gin.Context) {
	r := context.Request
	w := context.Writer
	if api.CheckRequestMethodReturn(w, r, "post") {
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		api.HttpError(w, "No body was found : "+err.Error(), http.StatusBadRequest)
	}

	var obj model.DeletePost
	err = json.Unmarshal(body, &obj)
	if err != nil {
		api.HttpError(w, "json unmarshall error:"+err.Error(), http.StatusBadRequest)
		return
	}

	errorMessage := ""
	if api.IsEmpty(&obj.Email) {
		errorMessage += "Missing paramter 'email'\n"
	}
	if api.IsEmpty(&obj.Password) {
		errorMessage += "Missing paramter 'password'\n"
	}
	if api.IsEmpty(&obj.PostID) {
		errorMessage += "Missing paramter 'post_id'\n"
	}
	if !api.IsEmpty(&errorMessage) {
		api.HttpError(w, errorMessage, http.StatusBadRequest)
		return
	}

	database, err := api.GetDatabase()
	if err != nil {
		api.HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer database.Close()

	user_id, err := model.GetUserID(database, obj.Email, obj.Password)
	if err != nil {
		api.HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sql := fmt.Sprintf("select p_publisher_id from posts where p_id = %s;", obj.PostID)
	rows, err := database.Query(sql)
	if err != nil {
		api.HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	if !rows.Next() {
		api.HttpError(w, "No Post Found", http.StatusBadRequest)
		return
	}

	var publisher_id int
	rows.Scan(&publisher_id)

	if user_id != publisher_id {
		api.HttpError(w, "You cannot delete post other than yourself's", http.StatusBadRequest)
		return
	}

	sql = fmt.Sprintf("CALL DeletePost(%s);", obj.PostID)
	_, err = database.Exec(sql)
	if err != nil {
		api.HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprint(w, "success")
}
