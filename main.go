package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/mail"
	"net/smtp"
	"os"
	"path/filepath"
	"rainbowwolfer/myspacegoserver/api"
	"rainbowwolfer/myspacegoserver/handlers"
	"rainbowwolfer/myspacegoserver/model"
	"strconv"
	"strings"
	"text/template"

	_ "github.com/go-sql-driver/mysql"
)

const (
	MAX_UPLOAD_SIZE = 1024 * 1024
	MAX_IMAGES_POST = 9
	PORT            = 4500
	HOST            = "http://www.cqtest.top"
)

func getIndexHandler(w http.ResponseWriter, r *http.Request) {
	if api.CheckRequestMethodReturn(w, r, "get") {
		return
	}
	w.Header().Add("Content-Type", "text/html")
	http.ServeFile(w, r, "index.html")
}

//Login - Get
//Error Code ->
//1-registration pending
//2-user not found (email or password is wrong)
func tryLogin_get(w http.ResponseWriter, r *http.Request) {
	if api.CheckRequestMethodReturn(w, r, "get") {
		return
	}
	query := r.URL.Query()
	if api.CheckMissingParamters(w, query, true, "email", "password") {
		return
	}
	email := query["email"][0]
	password := query["password"][0]
	sql := fmt.Sprintf("call GetUserByLogin('%s','%s')", email, password)

	database, err := api.GetDatabase()
	if err != nil {
		api.HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer database.Close()

	rows, err := database.Query(sql)
	if err != nil {
		api.HttpError(w, "Query Error with :"+sql+"\n"+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	if !rows.Next() {
		sql = fmt.Sprintf("SELECT ev_id FROM email_validations WHERE ev_email = '%s' AND ev_password = '%s'", email, password)
		rows_ev, err := database.Query(sql)
		if err != nil {
			api.HttpError(w, "Query Error: "+err.Error(), http.StatusInternalServerError)
			return
		}

		if rows_ev.Next() {
			api.HttpErrorWithCode(w, "User is in registration pending", http.StatusBadRequest, 1)
			return
		} else {
			api.HttpErrorWithCode(w, "No User Found", http.StatusBadRequest, 2)
			return //no found result
		}
	}

	user, err := model.ReadUser(rows)

	if err != nil {
		api.HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintln(w, api.ToJson(user))
}

func getUser_get(w http.ResponseWriter, r *http.Request) {
	if api.CheckRequestMethodReturn(w, r, "get") {
		return
	}

	query := r.URL.Query()
	if api.CheckMissingParamters(w, query, false, "id", "username") {
		return
	}

	self_id := -1
	if query.Has("self_id") {
		self_id_str := query["self_id"][0]
		if !api.IsEmpty(&self_id_str) {
			number, err := strconv.Atoi(self_id_str)
			if err != nil {
				api.HttpError(w, err.Error(), http.StatusBadRequest)
				return
			}
			self_id = number
		}
	}

	var sql string
	var errorMsg string
	if query.Has("id") {
		id := query["id"][0]
		sql = fmt.Sprintf("call GetUserByID('%s',%d)", id, self_id)
		errorMsg = "id of - " + id
	} else if query.Has("username") {
		username := query["username"][0]
		sql = fmt.Sprintf("call GetUserByUsername('%s',%d)", username, self_id)
		errorMsg = "username of - " + username
	}

	database, err := api.GetDatabase()
	if err != nil {
		api.HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer database.Close()

	rows, err := database.Query(sql)
	if err != nil {
		api.HttpError(w, "Query Error with :"+sql+"\n"+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	if !rows.Next() {
		api.HttpError(w, "Cannot find user with "+errorMsg, http.StatusBadRequest)
		return
	}

	user, err := model.ReadUser(rows)

	if err != nil {
		api.HttpError(w, err.Error(), http.StatusInternalServerError)
	}

	if err := rows.Err(); err != nil {
		api.HttpError(w, "Databse Rows Error"+err.Error(), http.StatusInternalServerError)
	}

	fmt.Fprintln(w, api.ToJson(user))
}

func uploadAvatar_post(w http.ResponseWriter, r *http.Request) {
	if api.CheckRequestMethodReturn(w, r, "post") {
		return
	}

	query := r.URL.Query()
	if api.CheckMissingParamters(w, query, true, "id") {
		return
	}
	id := query["id"][0]

	if err := r.ParseMultipartForm(32 << 20); err != nil {
		api.HttpError(w, err.Error(), http.StatusBadRequest)
		return
	}

	// get a reference to the fileHeaders
	files := r.MultipartForm.File["file"]

	fileLen := len(files)
	if fileLen != 1 {
		api.HttpError(w, fmt.Sprintf("Can only apply 1 file. Currently received %d file(s)", fileLen), http.StatusBadRequest)
		return
	}

	fileHeader := files[0]

	if fileHeader.Size > MAX_UPLOAD_SIZE {
		api.HttpError(w, fmt.Sprintf("The uploaded image is too big: %s. Please use an image less than 1MB in size", fileHeader.Filename), http.StatusBadRequest)
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		api.HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer file.Close()

	buff := make([]byte, 512)
	_, err = file.Read(buff)
	if err != nil {
		api.HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	filetype := http.DetectContentType(buff)
	if filetype != "image/jpeg" && filetype != "image/jpg" && filetype != "image/png" {
		api.HttpError(w, "The provided file format is not allowed. Please upload a JPEG(JPG) or PNG image", http.StatusBadRequest)
		return
	}

	_, err = file.Seek(0, io.SeekStart)
	if err != nil {
		api.HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	database, err := api.GetDatabase()
	if err != nil {
		api.HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer database.Close()

	sql := fmt.Sprintf("select u_id from users where u_id = '%s'", id)
	rows, err := database.Query(sql)
	if err != nil {
		api.HttpError(w, "Query Error with :"+sql+"\n"+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	foundUser := rows.Next()

	if !foundUser {
		api.HttpError(w, "User Not Found", http.StatusBadRequest)
		return
	}

	err = os.MkdirAll("./uploads/avatars", os.ModePerm)
	if err != nil {
		api.HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	f, err := os.Create(fmt.Sprintf("./uploads/avatars/%s%s", "user_"+id, filepath.Ext(fileHeader.Filename)))
	if err != nil {
		api.HttpError(w, err.Error(), http.StatusBadRequest)
		return
	}

	defer f.Close()

	pr := &api.Progress{
		TotalSize: fileHeader.Size,
	}

	_, err = io.Copy(f, io.TeeReader(file, pr))
	if err != nil {
		api.HttpError(w, err.Error(), http.StatusBadRequest)
		return
	}

	fmt.Fprintf(w, "Upload successful")
}

func getAvatar_get(w http.ResponseWriter, r *http.Request) {
	if api.CheckRequestMethodReturn(w, r, "get") {
		return
	}

	query := r.URL.Query()
	if api.CheckMissingParamters(w, query, true, "id") {
		return
	}

	id := query["id"][0]
	defaultPath := "./uploads/avatars/DefaultAvatar.png"
	path := fmt.Sprintf("./uploads/avatars/user_%s", id)

	bytes := api.GetImageWithDefault(path, defaultPath)
	if bytes == nil {
		api.HttpError(w, "File not found", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(bytes)
}

func validEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}

type loginAuth struct {
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

//Send Validation Email - Post
//Error Code ->
//1-email or username used
//2-already sent a email
func sendValidationEmail_post(w http.ResponseWriter, r *http.Request) {
	if api.CheckRequestMethodReturn(w, r, "post") {
		return
	}

	postBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		api.HttpError(w, "No body was found : "+err.Error(), http.StatusBadRequest)
	}

	var obj struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Email    string `json:"email"`
	}
	err = json.Unmarshal(postBody, &obj)
	if err != nil {
		api.HttpError(w, "json unmarshall error:"+err.Error(), http.StatusBadRequest)
		return
	}

	username_query := obj.Username
	password_query := obj.Password
	email_query := obj.Email

	errorMessage := ""
	if api.IsEmpty(&username_query) {
		errorMessage += "Missing paramter 'username'\n"
	} else if api.IsEmpty(&password_query) {
		errorMessage += "Missing paramter 'password'\n"
	} else if api.IsEmpty(&email_query) {
		errorMessage += "Missing paramter 'email'\n"
	}
	if !api.IsEmpty(&errorMessage) {
		api.HttpError(w, errorMessage, http.StatusBadRequest)
		return
	}

	if !validEmail(email_query) {
		api.HttpError(w, fmt.Sprintf("(%s) is not a valid email address", email_query), http.StatusBadRequest)
		return
	}

	sql := fmt.Sprintf("select u_id FROM users where u_email = '%s' or u_username = '%s'", email_query, username_query)

	database, err := api.GetDatabase()
	if err != nil {
		api.HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer database.Close()

	rows, err := database.Query(sql)
	if err != nil {
		api.HttpError(w, "Query Error with :"+sql+"\n"+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	if rows.Next() {
		api.HttpErrorWithCode(w, fmt.Sprintf("email (%s) or username (%s) is used for another account.", email_query, username_query), http.StatusBadRequest, 1)
		return
	}

	sql = fmt.Sprintf("select ev_id FROM email_validations where ev_email = '%s'", email_query)
	rows, err = database.Query(sql)
	if err != nil {
		api.HttpError(w, "sql query error", http.StatusInternalServerError)
		return
	}

	if rows.Next() {
		api.HttpErrorWithCode(w, fmt.Sprintf("already sent a email to (%s). please wait", email_query), http.StatusBadRequest, 2)
		return
	}

	combined := username_query + password_query + email_query
	code := api.GetMD5Hash(combined)

	to := []string{email_query}

	t, _ := template.ParseFiles("email_validation.html")
	var body bytes.Buffer

	mimeHeaders := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	body.Write([]byte(fmt.Sprintf("Subject: My Space Registration Email Validation \n%s\n\n", mimeHeaders)))

	url := fmt.Sprintf("%s:%d/validation/email/validate?email=%s&code=%s", HOST, PORT, email_query, code)

	t.Execute(&body, struct {
		Name string
		Link string
	}{
		Name: username_query,
		Link: url,
	})

	from := "1519787190@qq.com"
	password := "awowxbgooevfgbjc"

	smtpHost := "smtp.qq.com"
	smtpPort := "587"

	// m := gomail.NewMessage()
	// m.SetHeader(`From`, from)
	// m.SetHeader(`From`, email_query)
	// m.SetHeader(`Subject`, body.String())
	// d := gomail.NewDialer(smtpHost, 587, from, password)
	// d.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	// if err := d.DialAndSend(m); err != nil {
	// 	api.HttpError(w, err.Error(), http.StatusInternalServerError)
	// 	return
	// }

	auth := LoginAuth(from, password)

	err = smtp.SendMail(smtpHost+":"+smtpPort, auth, from, to, body.Bytes())
	if err != nil {
		api.HttpError(w, "sending email failed : "+err.Error(), http.StatusInternalServerError)
		return
	}

	sql = fmt.Sprintf("INSERT INTO email_validations (ev_email,ev_username,ev_password,ev_code,ev_datetime) VALUES ('%s','%s','%s','%s',NOW())", email_query, username_query, password_query, code)
	_, err = database.Exec(sql)
	if err != nil {
		api.HttpError(w, "insert data failed :"+err.Error(), http.StatusBadRequest)
		return
	}

	fmt.Fprintln(w, "Email Sent Successfully!")
}

func validateEmail_get(w http.ResponseWriter, r *http.Request) {
	if api.CheckRequestMethodReturn(w, r, "get") {
		return
	}

	query := r.URL.Query()
	if api.CheckMissingParamters(w, query, true, "email", "code") {
		return
	}

	email := query["email"][0]
	code := query["code"][0]
	//SELECT ev_code FROM email_validations WHERE ev_email = '1519787190@qq.com' AND ev_datetime <= NOW()

	sql := fmt.Sprintf("SELECT ev_code,ev_email,ev_username,ev_password FROM email_validations WHERE ev_email = '%s' AND ev_datetime <= NOW()", email)

	database, err := api.GetDatabase()
	if err != nil {
		api.HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer database.Close()

	rows, err := database.Query(sql)
	if err != nil {
		api.HttpError(w, "Query Error with :"+sql+"\n"+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	if !rows.Next() {
		api.HttpError(w, "validation not found", http.StatusBadRequest)
		return
	}

	var db_code string
	var db_username string
	var db_password string
	var db_email string

	rows.Scan(&db_code, &db_email, &db_username, &db_password)

	if code != db_code {
		api.HttpError(w, "code not matched", http.StatusBadRequest)
		return
	}

	//delete validation
	sql = fmt.Sprintf("DELETE FROM email_validations WHERE ev_email = '%s'", db_email)
	_, err = database.Exec(sql)
	if err != nil {
		api.HttpError(w, "database delete validation failed\n"+err.Error(), http.StatusInternalServerError)
		return
	}

	//add new user
	sql = fmt.Sprintf("INSERT INTO users (u_username,u_password,u_email) VALUES ('%s','%s','%s')", db_username, db_password, db_email)
	_, err = database.Exec(sql)
	if err != nil {
		api.HttpError(w, "database insert new user failed\n"+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "text/html")
	http.ServeFile(w, r, "email_validation_success.html")
}

func checkUserExist_get(w http.ResponseWriter, r *http.Request) {
	if api.CheckRequestMethodReturn(w, r, "get") {
		return
	}

	query := r.URL.Query()
	if api.CheckMissingParamters(w, query, false, "username", "email") {
		return
	}

	var sql string
	if query.Has("username") {
		username := query["username"][0]
		sql = fmt.Sprintf("SELECT u_id FROM users where u_username = '%s'", username)
	} else if query.Has("email") {
		email := query["email"][0]
		sql = fmt.Sprintf("SELECT u_id FROM users where u_email = '%s'", email)
	}

	database, err := api.GetDatabase()
	if err != nil {
		api.HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer database.Close()

	rows, err := database.Query(sql)
	if err != nil {
		api.HttpError(w, "Query Error with :"+sql+"\n"+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	foundUser := rows.Next()

	fmt.Fprintln(w, foundUser)
}

//Update Username - Post
//Error Code ->
//1-username taken
func updateUsername_post(w http.ResponseWriter, r *http.Request) {
	if api.CheckRequestMethodReturn(w, r, "post") {
		return
	}

	defer r.Body.Close()

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		api.HttpError(w, "No body was found : "+err.Error(), http.StatusBadRequest)
	}

	var obj model.NewUsername
	err = json.Unmarshal(body, &obj)
	if err != nil {
		api.HttpError(w, "json unmarshall error:"+err.Error(), http.StatusBadRequest)
		return
	}

	errorMessage := ""
	if api.IsEmpty(&obj.ID) {
		errorMessage += "Missing paramter 'id'\n"
	}
	if api.IsEmpty(&obj.Username) {
		errorMessage += "Missing paramter 'username'\n"
	}
	if api.IsEmpty(&obj.Password) {
		errorMessage += "Missing paramter 'password'\n"
	}
	if api.IsEmpty(&obj.NewUsername) {
		errorMessage += "Missing paramter 'new_username'\n"
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

	sql := fmt.Sprintf("SELECT u_id FROM users WHERE u_username = '%s'", obj.NewUsername)
	rows, err := database.Query(sql)
	if err != nil {
		api.HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	if rows.Next() {
		api.HttpErrorWithCode(w, fmt.Sprintf("There is already a user named (%s)", obj.NewUsername), http.StatusBadRequest, 1)
		return
	}

	sql = fmt.Sprintf("UPDATE users SET u_username = '%s' WHERE u_id = '%s' AND u_username = '%s' AND u_password = '%s';", obj.NewUsername, obj.ID, obj.Username, obj.Password)

	res, err := database.Exec(sql)
	if err != nil {
		api.HttpError(w, "Query Error with :"+sql, http.StatusBadRequest)
		return
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		api.HttpError(w, "Effect 0 row", http.StatusBadRequest)
		return
	}

	fmt.Fprintln(w, rowsAffected)
}

func post_post_get(w http.ResponseWriter, r *http.Request) {
	if err := api.CheckRequestMethod(r, "post"); err == nil {
		if api.CheckRequestMethodReturn(w, r, "post") {
			return
		}

		if err := r.ParseMultipartForm(32 << 20); err != nil {
			api.HttpError(w, err.Error(), http.StatusBadRequest)
			return
		}

		content := r.MultipartForm.Value["content"][0]
		publisherID := r.MultipartForm.Value["publisher_id"][0]
		postVisibility := r.MultipartForm.Value["post_visibility"][0]
		replyVisibility := r.MultipartForm.Value["reply_visibility"][0]
		tags := strings.Split(r.MultipartForm.Value["tags"][0], "&#10;")
		images := r.MultipartForm.File["post_images"]
	
		for _, header := range images {
			if header.Size > MAX_UPLOAD_SIZE {
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
			_, err = file.Read(buff)
			if err != nil {
				api.HttpError(w, err.Error(), http.StatusInternalServerError)
				return
			}
			filetype := http.DetectContentType(buff)
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

		fmt.Fprintln(w, json)
	} else {
		api.HttpError(w, "Only get or post method is allowed", http.StatusMethodNotAllowed)
	}
}

func getPostImage_get(w http.ResponseWriter, r *http.Request) {
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

func comment_post_get(w http.ResponseWriter, r *http.Request) {
	if err := api.CheckRequestMethod(r, "post"); err == nil {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			api.HttpError(w, "No body was found : "+err.Error(), http.StatusBadRequest)
		}
		
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

func getPostComments_get(w http.ResponseWriter, r *http.Request) {
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

func votePost_post(w http.ResponseWriter, r *http.Request) {
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

func repost_post(w http.ResponseWriter, r *http.Request) {
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

	_, err = database.Exec(sql)
	if err != nil {
		api.HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

}

func voteComment_post(w http.ResponseWriter, r *http.Request) {
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

func getRepostRecords_get(w http.ResponseWriter, r *http.Request) {
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

func getScoreRecords_get(w http.ResponseWriter, r *http.Request) {
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

func addToCollection_post(w http.ResponseWriter, r *http.Request) {
	if api.CheckRequestMethodReturn(w, r, "post") {
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		api.HttpError(w, "No body was found : "+err.Error(), http.StatusBadRequest)
	}

	var obj model.NewCollection
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
	if api.IsEmpty(&obj.TargetID) {
		errorMessage += "Missing paramter 'target_id'\n"
	}
	if api.IsEmpty(&obj.Type) {
		errorMessage += "Missing paramter 'type'\n"
	}
	if !api.IsEmpty(&errorMessage) {
		api.HttpError(w, errorMessage, http.StatusBadRequest)
		return
	}

	if obj.Type != "POST" && obj.Type != "MESSAGE" {
		api.HttpError(w, "Type Error: ("+obj.Type+") is not defined", http.StatusBadRequest)
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
	} else if user_id <= 0 {
		api.HttpError(w, "User not found", http.StatusBadRequest)
		return
	}

	sql := fmt.Sprintf("insert into user_collections(uc_id_user, uc_id_target, uc_type) values (%d,'%s','%s');", user_id, obj.TargetID, obj.Type)
	_, err = database.Exec(sql)
	if err != nil {
		api.HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "Success")
}

func getCollections_get(w http.ResponseWriter, r *http.Request) {
	if api.CheckRequestMethodReturn(w, r, "get") {
		return
	}

	query := r.URL.Query()
	if api.CheckMissingParamters(w, query, true, "email", "password", "offset", "limit") {
		return
	}

	email := query["email"][0]
	password := query["password"][0]
	offset, err := strconv.Atoi(query["offset"][0])
	if err != nil {
		api.HttpError(w, err.Error(), http.StatusBadRequest)
		return
	}
	limit, err := strconv.Atoi(query["limit"][0])
	if err != nil {
		api.HttpError(w, err.Error(), http.StatusBadRequest)
		return
	}

	if api.IsEmpty(&email) || api.IsEmpty(&password) {
		api.HttpError(w, "email or password cannot be empty", http.StatusBadRequest)
		return
	}

	database, err := api.GetDatabase()
	if err != nil {
		api.HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer database.Close()

	user_id, err := model.GetUserID(database, email, password)
	if err != nil {
		api.HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	} else if user_id <= 0 {
		api.HttpError(w, "User not found", http.StatusBadRequest)
		return
	}

	sql := fmt.Sprintf("SELECT * FROM collections_view WHERE uc_id_user = %d LIMIT %d,%d;", user_id, offset, limit)
	
	rows, err := database.Query(sql)
	if err != nil {
		api.HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var list []model.Collection
	for rows.Next() {
		var item model.Collection
		if err = rows.Scan(
			&item.ID,
			&item.UserID,
			&item.TargetID,
			&item.Type,
			&item.Time,
			&item.PublisherID,
			&item.PublisherUsername,
			&item.TextContent,
			&item.ImagesCount,
			&item.IsRepost,
		); err != nil {
			api.HttpError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		list = append(list, item)
	}

	fmt.Fprint(w, api.ToJson(list))
}

func removeCollection_post(w http.ResponseWriter, r *http.Request) {
	if api.CheckRequestMethodReturn(w, r, "post") {
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		api.HttpError(w, "No body was found : "+err.Error(), http.StatusBadRequest)
	}

	var obj model.RemoveCollection
	err = json.Unmarshal(body, &obj)
	if err != nil {
		api.HttpError(w, "json unmarshall error:"+err.Error(), http.StatusBadRequest)
		return
	}

	println(fmt.Sprintf("%v", obj))

	errorMessage := ""
	if api.IsEmpty(&obj.Email) {
		errorMessage += "Missing paramter 'email'\n"
	}
	if api.IsEmpty(&obj.Password) {
		errorMessage += "Missing paramter 'password'\n"
	}
	if api.IsEmpty(&obj.TargetID) {
		errorMessage += "Missing paramter 'target_id'\n"
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
	} else if user_id <= 0 {
		api.HttpError(w, "User not found", http.StatusBadRequest)
		return
	}

	sql := fmt.Sprintf("DELETE FROM user_collections WHERE uc_id_target = '%s' and uc_id_user = %d;", obj.TargetID, user_id)

	_, err = database.Exec(sql)
	if err != nil {
		api.HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprint(w, "Success")
}

func getPostByID_get(w http.ResponseWriter, r *http.Request) {
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
}

func userFollow_post(w http.ResponseWriter, r *http.Request) {
	if api.CheckRequestMethodReturn(w, r, "post") {
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		api.HttpError(w, "No body was found : "+err.Error(), http.StatusBadRequest)
	}

	var obj model.NewUserFollow
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
	if api.IsEmpty(&obj.TargetID) {
		errorMessage += "Missing paramter 'target_id'\n"
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
	var sql string
	if obj.Cancel {
		sql = fmt.Sprintf("DELETE FROM users_follows WHERE uf_id_follower = %d and uf_id_target = %s;", user_id, obj.TargetID)
	} else {
		sql = fmt.Sprintf("insert into users_follows (uf_id_follower, uf_id_target) values (%d,%s);", user_id, obj.TargetID)
	}

	_, err = database.Exec(sql)
	if err != nil {
		api.HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprint(w, "Success")
}

func getUserByUserID_get(w http.ResponseWriter, r *http.Request) {
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

func getPostsAndFollowersCount_get(w http.ResponseWriter, r *http.Request) {
	if api.CheckRequestMethodReturn(w, r, "get") {
		return
	}
	query := r.URL.Query()
	if api.CheckMissingParamters(w, query, true, "user_id") {
		return
	}

	user_id := query["user_id"][0]

	database, err := api.GetDatabase()
	if err != nil {
		api.HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer database.Close()

	sql := fmt.Sprintf("CALL GetUserPostAndFollowersCount(%s);", user_id)
	rows, err := database.Query(sql)
	if err != nil {
		api.HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	if !rows.Next() {
		api.HttpError(w, "no row", http.StatusInternalServerError)
		return
	}

	var postsCount int
	var followersCount int
	err = rows.Scan(&postsCount, &followersCount)
	if err != nil {
		api.HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprint(w, api.ToJson([]int{postsCount, followersCount}))
}

func getUserFollowers_get(w http.ResponseWriter, r *http.Request) {
	if api.CheckRequestMethodReturn(w, r, "get") {
		return
	}
	query := r.URL.Query()
	if api.CheckMissingParamters(w, query, true, "user_id") {
		return
	}

	user_id := query["user_id"][0]
	self_id := -1

	database, err := api.GetDatabase()
	if err != nil {
		api.HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer database.Close()

	if query.Has("email") && query.Has("password") {
		email := query["email"][0]
		password := query["password"][0]
		if !api.IsEmpty(&email) && !api.IsEmpty(&password) {
			_id, err := model.GetUserID(database, email, password)
			if err != nil {
				api.HttpError(w, err.Error(), http.StatusInternalServerError)
				return
			}
			self_id = _id
		}
	}

	sql := fmt.Sprintf("CALL GetUserFollowers(%s,%d)", user_id, self_id)
	rows, err := database.Query(sql)
	if err != nil {
		api.HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var list []model.User

	for rows.Next() {
		item, err := model.ReadUser(rows)
		if err != nil {
			api.HttpError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		list = append(list, item)
	}

	fmt.Fprint(w, api.ToJson(list))
}

func deletePost_post(w http.ResponseWriter, r *http.Request) {
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

func getMessageContacts_get(w http.ResponseWriter, r *http.Request) {
	if api.CheckRequestMethodReturn(w, r, "get") {
		return
	}
	query := r.URL.Query()
	if api.CheckMissingParamters(w, query, true, "email", "password") {
		return
	}
	email := query["email"][0]
	password := query["password"][0]

	database, err := api.GetDatabase()
	if err != nil {
		api.HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer database.Close()

	user_id, err := model.GetUserID(database, email, password)
	if err != nil {
		api.HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if user_id == -1 {
		api.HttpError(w, "User not found", http.StatusBadRequest)
		return
	}

	sql := fmt.Sprintf("call GetMessageContacts(%d);", user_id)
	rows, err := database.Query(sql)
	if err != nil {
		api.HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var list []model.MessageContact
	for rows.Next() {
		item, err := model.ReadMessageContact(rows)
		if err != nil {
			api.HttpError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		list = append(list, item)
	}

	fmt.Fprint(w, api.ToJson(list))
}

func getMessages_get(w http.ResponseWriter, r *http.Request) {
	if api.CheckRequestMethodReturn(w, r, "get") {
		return
	}
	query := r.URL.Query()
	if api.CheckMissingParamters(w, query, true, "email", "password", "contact_id", "offset", "limit") {
		return
	}

	email := query["email"][0]
	password := query["password"][0]
	contact_id := query["contact_id"][0]
	offset := query["offset"][0]
	limit := query["limit"][0]

	database, err := api.GetDatabase()
	if err != nil {
		api.HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer database.Close()

	user_id, err := model.GetUserID(database, email, password)
	if err != nil {
		api.HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if user_id == -1 {
		api.HttpError(w, "User not found", http.StatusBadRequest)
		return
	}
	
	sql := fmt.Sprintf("call GetMessagesByContact(%d,%s,%s,%s);", user_id, contact_id, offset, limit)
	rows, err := database.Query(sql)
	if err != nil {
		api.HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var list []model.Message
	for rows.Next() {
		item, err := model.ReadMessage(rows)
		if err != nil {
			api.HttpError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		list = append(list, item)
	}

	fmt.Fprint(w, api.ToJson(list))
}

func flagReceived_post(w http.ResponseWriter, r *http.Request) {
	if api.CheckRequestMethodReturn(w, r, "post") {
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		api.HttpError(w, "No body was found : "+err.Error(), http.StatusBadRequest)
	}

	var obj model.FlagMessage
	err = json.Unmarshal(body, &obj)
	if err != nil {
		api.HttpError(w, "json unmarshall error:"+err.Error(), http.StatusBadRequest)
		return
	}

	if err = obj.CheckValid(); err != nil {
		api.HttpError(w, err.Error(), http.StatusBadRequest)
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
	if user_id == -1 {
		api.HttpError(w, "User not found", http.StatusBadRequest)
		return
	}

	sql := fmt.Sprintf("CALL FlagHasReceived(%d,%s);", user_id, obj.SenderID)
	_, err = database.Exec(sql)
	if err != nil {
		api.HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprint(w, "success")
}

func flagUnread_post(w http.ResponseWriter, r *http.Request) {
	if api.CheckRequestMethodReturn(w, r, "post") {
		return
	}

}

func main() {
	println(api.Now())
	mux := http.NewServeMux()

	mux.HandleFunc("/", getIndexHandler) //get

	mux.HandleFunc("/user", getUser_get)                                          //get
	mux.HandleFunc("/user/checkExisting", checkUserExist_get)                     //get
	mux.HandleFunc("/user/avatar", getAvatar_get)                                 //get
	mux.HandleFunc("/user/update/username", updateUsername_post)                  //post
	mux.HandleFunc("/user/follow", userFollow_post)                               //post
	mux.HandleFunc("/user/getFollowers", getUserFollowers_get)                    //get
	mux.HandleFunc("/user/postsAndFollowersCount", getPostsAndFollowersCount_get) //get

	mux.HandleFunc("/login", tryLogin_get)              //get
	mux.HandleFunc("/upload/avatar", uploadAvatar_post) //post

	mux.HandleFunc("/validation/email/send", sendValidationEmail_post) //post
	mux.HandleFunc("/validation/email/validate", validateEmail_get)    //get

	mux.HandleFunc("/post", post_post_get)                      //post/get
	mux.HandleFunc("/post/delete", deletePost_post)             //post
	mux.HandleFunc("/post/user", getUserByUserID_get)           //get
	mux.HandleFunc("/post/id", getPostByID_get)                 //get
	mux.HandleFunc("/post/images", getPostImage_get)            //get
	mux.HandleFunc("/post/comment", comment_post_get)           //post/get
	mux.HandleFunc("/post/comment/vote", voteComment_post)      //get
	mux.HandleFunc("/post/comments", getPostComments_get)       //get
	mux.HandleFunc("/post/vote", votePost_post)                 //post
	mux.HandleFunc("/post/repost", repost_post)                 //post
	mux.HandleFunc("/post/repostRecords", getRepostRecords_get) //get
	mux.HandleFunc("/post/scoreRecords", getScoreRecords_get)   //get

	mux.HandleFunc("/collections/add", addToCollection_post)     //post
	mux.HandleFunc("/collections/remove", removeCollection_post) //post
	mux.HandleFunc("/collections", getCollections_get)           //get

	mux.HandleFunc("/message/contacts", getMessageContacts_get) //get
	mux.HandleFunc("/message/get", getMessages_get)             //get
	mux.HandleFunc("/message/flagReceived", flagReceived_post)  //post
	mux.HandleFunc("/message/flagUnread", flagUnread_post)      //post

	mux.HandleFunc("/admin/clearunusedpostimages", handlers.ClearUnusedPostImages)
	mux.HandleFunc("/admin/reinflatedefaultposts", handlers.ReinflateDefaultPosts)

	if err := http.ListenAndServe(":4500", mux); err != nil {
		log.Fatal(err)
	}

}
