package main

import (
	"bytes"
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/mail"
	"net/smtp"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

const (
	MAX_UPLOAD_SIZE = 1024 * 1024
	MAX_IMAGES_POST = 9
	PORT            = 4500
	HOST            = "http://www.cqtest.top"
	// HOST            = "http://127.0.0.1"
)

type Output struct {
	Code    int
	Content string
	Error   int
}

func httpError(w http.ResponseWriter, content string, code int) {
	c := fmt.Sprintf("%s (%d)", content, code)
	json := toJson(Output{
		Code:    code,
		Content: c,
		Error:   0,
	})
	println(json)
	http.Error(w, json, code)
}

func httpErrorWithCode(w http.ResponseWriter, content string, code int, errorCode int) {
	c := fmt.Sprintf("%s (%d)", content, code)
	json := toJson(Output{
		Code:    code,
		Content: c,
		Error:   errorCode,
	})
	println(json)
	http.Error(w, json, code)
}

type Progress struct {
	TotalSize int64
	BytesRead int64
}

func (pr *Progress) Write(p []byte) (n int, err error) {
	n, err = len(p), nil
	pr.BytesRead += int64(n)
	pr.Print()
	return
}

func (pr *Progress) Print() {
	if pr.BytesRead == pr.TotalSize {
		fmt.Println("DONE!")
		return
	}
	fmt.Printf("File upload in progress: %d\n", pr.BytesRead)
}

type User struct {
	ID                 string
	Username           string
	Password           string
	Email              string
	ProfileDescription string
}

type Post struct {
	ID           string
	PublisherID  string
	PublishDate  string
	EditDate     string
	EditTimes    int
	TextContent  string
	Deleted      bool
	ImagesCount  int
	Tags         string
	Upvotes      int
	Downvotes    int
	Repost       int
	Comment      int
	Visibility   string
	Reply        string
	IsRepost     bool
	OriginPostID string
	ReposterID   string
}

func getIndexHandler(w http.ResponseWriter, r *http.Request) {
	if checkRequestMethodReturn(w, r, "get") {
		return
	}
	w.Header().Add("Content-Type", "text/html")
	http.ServeFile(w, r, "index.html")
}

func getMD5Hash(text string) string {
	hash := md5.Sum([]byte(text))
	return hex.EncodeToString(hash[:])
}

func toJson(object any) string {
	j, err := json.MarshalIndent(object, "", "\t")
	if err != nil {
		return "Error Converting Object to Json"
	} else {
		return string(j)
	}
}

func getDatabase(w http.ResponseWriter) (*sql.DB, error) {
	db, err := sql.Open("mysql", "wjx:123456@tcp(www.cqtest.top:3306)/wjx")
	if err != nil {
		return nil, err
	}
	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)
	return db, nil
}

func checkRequestMethodReturn(w http.ResponseWriter, r *http.Request, method string) bool {
	println(r.URL.RawPath)
	if !strings.EqualFold(r.Method, method) {
		httpError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return true
	} else {
		return false
	}
}

func checkRequestMethod(r *http.Request, method string) error {
	println(r.URL.RawPath)
	if strings.EqualFold(r.Method, method) {
		return nil
	} else {
		return errors.New("method not allowed")
	}
}

func isEmpty(str *string) bool {
	if str == nil {
		return false
	}
	if len(*str) == 0 {
		return false
	}
	if len(strings.TrimSpace(*str)) == 0 {
		return false
	}

	return true
}

// func unmarshallPostBody[T *any](r *http.Request) (T, error) {
// 	body, err := ioutil.ReadAll(r.Body)
// 	if err != nil {
// 		return nil, errors.New("read request body error\n" + err.Error())
// 	}
// 	var obj T
// 	err = json.Unmarshal(body, &obj)
// 	if err != nil {
// 		return nil, errors.New("json unmarshall error\n" + err.Error())
// 	}

// 	return obj, nil
// }

func checkMissingParamters(w http.ResponseWriter, query url.Values, and bool, paras ...string) bool {
	var missings []string = []string{}
	for _, p := range paras {
		if !query.Has(p) {
			missings = append(missings, p)
			continue
		}
		content := query[p][0]
		if len(strings.TrimSpace(content)) == 0 {
			missings = append(missings, p)
		}
	}
	if and {
		if len(missings) == 0 {
			return false
		} else {
			var errorLines []string
			for _, m := range missings {
				errorLines = append(errorLines, "'"+m+"'")
			}
			httpError(w, "Missing Parameter "+strings.Join(errorLines, " And "), http.StatusBadRequest)
			return true
		}
	} else {
		if len(missings) == len(paras) {
			var errorLines []string
			for _, m := range missings {
				errorLines = append(errorLines, "'"+m+"'")
			}
			httpError(w, "Missing Parameter "+strings.Join(errorLines, " Or "), http.StatusBadRequest)
			return true
		} else {
			return false
		}
	}
}
func deleteEmpty(s []string) []string {
	var r []string
	for _, str := range s {
		if str != "" {
			r = append(r, str)
		}
	}
	return r
}

func queryForRows(w http.ResponseWriter, sql string) (*sql.DB, *sql.Rows) {
	db, err := getDatabase(w)
	if err != nil {
		httpError(w, "Database Error", http.StatusInternalServerError)
	}
	rows, err := db.Query(sql)
	if err != nil {
		httpError(w, "Query Error with :"+sql, http.StatusInternalServerError)
	}
	return db, rows
}

func getImage(path string) []byte {
	availableExts := []string{"", ".png", ".jpg", ".jpeg"}
	var bytes []byte = nil
	for _, ext := range availableExts {
		p := fmt.Sprintf("%s%s", path, ext)
		println("getting file: " + p)
		fileBytes, err := ioutil.ReadFile(p)
		if err != nil {
			println("not found")
			continue
		}
		println("found!")
		bytes = fileBytes
		break
	}
	return bytes
}

func getImageWithDefault(path string, defaultPath string) []byte {
	bytes := getImage(path)
	if bytes == nil {
		b, err := ioutil.ReadFile(defaultPath)
		if err != nil {
			println("Default file not found")
			return nil
		}
		bytes = b
	}
	return bytes
}

//Login - Get
//Error Code ->
//1-registration pending
//2-user not found (email or password is wrong)
func getTryLogin(w http.ResponseWriter, r *http.Request) {
	if checkRequestMethodReturn(w, r, "get") {
		return
	}
	query := r.URL.Query()
	if checkMissingParamters(w, query, true, "email", "password") {
		return
	}
	email := query["email"][0]
	password := query["password"][0]
	sql := fmt.Sprintf("select * from users where u_email='%s' and u_password='%s'", email, password)

	db, rows := queryForRows(w, sql)
	if db == nil || rows == nil {
		return
	}
	defer db.Close()
	defer rows.Close()

	if !rows.Next() {
		sql = fmt.Sprintf("SELECT ev_id FROM email_validations WHERE ev_email = '%s' AND ev_password = '%s'", email, password)
		rows, err := db.Query(sql)
		if err != nil {
			httpError(w, "Query Error: "+err.Error(), http.StatusInternalServerError)
			return
		}

		if rows.Next() {
			httpErrorWithCode(w, "User is in registration pending", http.StatusBadRequest, 1)
			return
		} else {
			httpErrorWithCode(w, "No User Found", http.StatusBadRequest, 2)
			return //no found result
		}
	}

	var user User
	if err := rows.Scan(
		&user.ID,
		&user.Username,
		&user.Password,
		&user.Email,
		&user.ProfileDescription,
	); err != nil {
		httpError(w, "User Convert Error", http.StatusInternalServerError)
	}

	fmt.Fprintln(w, toJson(user))
}

func getUser(w http.ResponseWriter, r *http.Request) {
	if checkRequestMethodReturn(w, r, "get") {
		return
	}

	query := r.URL.Query()
	if checkMissingParamters(w, query, false, "id", "username") {
		return
	}

	var sql string
	var errorMsg string
	if query.Has("id") {
		id := query["id"][0]
		sql = fmt.Sprintf("SELECT * FROM users where u_id = '%s'", id)
		errorMsg = "id of - " + id
	} else if query.Has("username") {
		username := query["username"][0]
		sql = fmt.Sprintf("SELECT * FROM users where u_username = '%s'", username)
		errorMsg = "username of - " + username
	}

	db, rows := queryForRows(w, sql)
	if db == nil || rows == nil {
		return
	}
	defer db.Close()
	defer rows.Close()

	if !rows.Next() {
		httpError(w, "Cannot find user with "+errorMsg, http.StatusBadRequest)
		return
	}
	var user User
	err := rows.Scan(
		&user.ID,
		&user.Username,
		&user.Password,
		&user.Email,
		&user.ProfileDescription,
	)

	if err != nil {
		httpError(w, "User Convert Error", http.StatusInternalServerError)
	}

	if err := rows.Err(); err != nil {
		httpError(w, "Databse Rows Error", http.StatusInternalServerError)
	}

	fmt.Fprintln(w, toJson(user))
}

func postUploadAvatar(w http.ResponseWriter, r *http.Request) {
	if checkRequestMethodReturn(w, r, "post") {
		return
	}

	query := r.URL.Query()
	if checkMissingParamters(w, query, true, "id") {
		return
	}
	id := query["id"][0]

	if err := r.ParseMultipartForm(32 << 20); err != nil {
		httpError(w, err.Error(), http.StatusBadRequest)
		return
	}

	// get a reference to the fileHeaders
	files := r.MultipartForm.File["file"]

	fileLen := len(files)
	if fileLen != 1 {
		httpError(w, fmt.Sprintf("Can only apply 1 file. Currently received %d file(s)", fileLen), http.StatusBadRequest)
		return
	}

	fileHeader := files[0]

	if fileHeader.Size > MAX_UPLOAD_SIZE {
		httpError(w, fmt.Sprintf("The uploaded image is too big: %s. Please use an image less than 1MB in size", fileHeader.Filename), http.StatusBadRequest)
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		httpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer file.Close()

	buff := make([]byte, 512)
	_, err = file.Read(buff)
	if err != nil {
		httpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	println(buff)

	filetype := http.DetectContentType(buff)
	if filetype != "image/jpeg" && filetype != "image/jpg" && filetype != "image/png" {
		httpError(w, "The provided file format is not allowed. Please upload a JPEG(JPG) or PNG image", http.StatusBadRequest)
		return
	}

	_, err = file.Seek(0, io.SeekStart)
	if err != nil {
		httpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sql := fmt.Sprintf("select u_id from users where u_id = '%s'", id)
	db, rows := queryForRows(w, sql)
	if db == nil || rows == nil {
		return
	}
	defer db.Close()
	defer rows.Close()

	foundUser := rows.Next()

	if !foundUser {
		httpError(w, "User Not Found", http.StatusBadRequest)
		return
	}

	err = os.MkdirAll("./uploads/avatars", os.ModePerm)
	if err != nil {
		httpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	//"_"+strconv.FormatInt(time.Now().UnixNano(), 10)
	f, err := os.Create(fmt.Sprintf("./uploads/avatars/%s%s", "user_"+id, filepath.Ext(fileHeader.Filename)))
	if err != nil {
		httpError(w, err.Error(), http.StatusBadRequest)
		return
	}

	defer f.Close()

	pr := &Progress{
		TotalSize: fileHeader.Size,
	}

	_, err = io.Copy(f, io.TeeReader(file, pr))
	if err != nil {
		httpError(w, err.Error(), http.StatusBadRequest)
		return
	}

	//this is file path in project directory
	println(f.Name())

	// sql = fmt.Sprintf("UPDATE users SET u_avatarPath = '%s' WHERE u_id = '%s';", f.Name(), id)
	// if rows, err = db.Query(sql); err != nil {
	// 	httpError(w, "Query Error with1: "+err.Error(), http.StatusInternalServerError)
	// }
	// if err = rows.Err(); err != nil {
	// 	httpError(w, "Query Error with2: "+err.Error(), http.StatusInternalServerError)
	// }

	// //this is project path
	// path, err := os.Getwd()
	// if err != nil {
	// 	httpError(w, err.Error(), http.StatusBadRequest)
	// }

	// fmt.Println(path)

	fmt.Fprintf(w, "Upload successful")
}

func getAvatar(w http.ResponseWriter, r *http.Request) {
	if checkRequestMethodReturn(w, r, "get") {
		return
	}

	query := r.URL.Query()
	if checkMissingParamters(w, query, true, "id") {
		return
	}

	id := query["id"][0]
	defaultPath := "./uploads/avatars/DefaultAvatar.png"
	path := fmt.Sprintf("./uploads/avatars/user_%s", id)

	bytes := getImageWithDefault(path, defaultPath)
	if bytes == nil {
		httpError(w, "File not found", http.StatusBadRequest)
		return
	}

	println(bytes)
	println(bytes == nil)
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
func postSendValidationEmail(w http.ResponseWriter, r *http.Request) {
	if checkRequestMethodReturn(w, r, "post") {
		return
	}

	postBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		httpError(w, "No body was found : "+err.Error(), http.StatusBadRequest)
	}

	var obj struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Email    string `json:"email"`
	}
	err = json.Unmarshal(postBody, &obj)
	if err != nil {
		httpError(w, "json unmarshall error:"+err.Error(), http.StatusBadRequest)
		return
	}

	username_query := obj.Username
	password_query := obj.Password
	email_query := obj.Email

	errorMessage := ""
	if isEmpty(&username_query) {
		errorMessage += "Missing paramter 'username'\n"
	} else if isEmpty(&password_query) {
		errorMessage += "Missing paramter 'password'\n"
	} else if isEmpty(&email_query) {
		errorMessage += "Missing paramter 'email'\n"
	}
	if !isEmpty(&errorMessage) {
		httpError(w, errorMessage, http.StatusBadRequest)
		return
	}

	if !validEmail(email_query) {
		httpError(w, fmt.Sprintf("(%s) is not a valid email address", email_query), http.StatusBadRequest)
		return
	}

	sql := fmt.Sprintf("select u_id FROM users where u_email = '%s' or u_username = '%s'", email_query, username_query)

	db, rows := queryForRows(w, sql)
	if db == nil || rows == nil {
		return
	}
	defer db.Close()
	defer rows.Close()

	if rows.Next() {
		httpErrorWithCode(w, fmt.Sprintf("email (%s) or username (%s) is used for another account.", email_query, username_query), http.StatusBadRequest, 1)
		return
	}

	sql = fmt.Sprintf("select ev_id FROM email_validations where ev_email = '%s'", email_query)
	rows, err = db.Query(sql)
	if err != nil {
		httpError(w, "sql query error", http.StatusInternalServerError)
		return
	}

	if rows.Next() {
		httpErrorWithCode(w, fmt.Sprintf("already sent a email to (%s). please wait", email_query), http.StatusBadRequest, 2)
		return
	}

	combined := username_query + password_query + email_query
	code := getMD5Hash(combined)

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

	from := "rainbowwolfer@outlook.com"
	password := "Windows15best"

	smtpHost := "smtp.office365.com"
	smtpPort := "587"

	auth := LoginAuth(from, password)

	err = smtp.SendMail(smtpHost+":"+smtpPort, auth, from, to, body.Bytes())
	if err != nil {
		httpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sql = fmt.Sprintf("INSERT INTO email_validations (ev_email,ev_username,ev_password,ev_code,ev_datetime) VALUES ('%s','%s','%s','%s',NOW())", email_query, username_query, password_query, code)
	_, err = db.Exec(sql)
	if err != nil {
		httpError(w, "insert data failed :"+err.Error(), http.StatusBadRequest)
		return
	}

	fmt.Fprintln(w, "Email Sent Successfully!")
}

func getValidateEmail(w http.ResponseWriter, r *http.Request) {
	if checkRequestMethodReturn(w, r, "get") {
		return
	}

	query := r.URL.Query()
	if checkMissingParamters(w, query, true, "email", "code") {
		return
	}

	email := query["email"][0]
	code := query["code"][0]
	//SELECT ev_code FROM email_validations WHERE ev_email = '1519787190@qq.com' AND ev_datetime <= NOW()

	sql := fmt.Sprintf("SELECT ev_code,ev_email,ev_username,ev_password FROM email_validations WHERE ev_email = '%s' AND ev_datetime <= NOW()", email)

	db, rows := queryForRows(w, sql)
	if db == nil || rows == nil {
		return
	}
	defer db.Close()
	defer rows.Close()

	if !rows.Next() {
		httpError(w, "validation not found", http.StatusBadRequest)
		return
	}

	var db_code string
	var db_username string
	var db_password string
	var db_email string

	rows.Scan(&db_code, &db_email, &db_username, &db_password)

	if code != db_code {
		httpError(w, "code not matched", http.StatusBadRequest)
		return
	}

	//delete validation
	sql = fmt.Sprintf("DELETE FROM email_validations WHERE ev_email = '%s'", db_email)
	_, err := db.Exec(sql)
	if err != nil {
		httpError(w, "database delete validation failed\n"+err.Error(), http.StatusInternalServerError)
		return
	}

	//add new user
	sql = fmt.Sprintf("INSERT INTO users (u_username,u_password,u_email) VALUES ('%s','%s','%s')", db_username, db_password, db_email)
	_, err = db.Exec(sql)
	if err != nil {
		httpError(w, "database insert new user failed\n"+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "text/html")
	http.ServeFile(w, r, "email_validation_success.html")
}

func getCheckUserExist(w http.ResponseWriter, r *http.Request) {
	if checkRequestMethodReturn(w, r, "get") {
		return
	}

	query := r.URL.Query()
	if checkMissingParamters(w, query, false, "username", "email") {
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

	db, rows := queryForRows(w, sql)
	if db == nil || rows == nil {
		return
	}
	defer db.Close()
	defer rows.Close()

	foundUser := rows.Next()

	fmt.Fprintln(w, foundUser)
}

type NewUsername struct {
	Id          string `json:"id"`
	Username    string `json:"username"`
	Password    string `json:"password"`
	NewUsername string `json:"new_username"`
}

func postUpdateUsername(w http.ResponseWriter, r *http.Request) {
	if checkRequestMethodReturn(w, r, "post") {
		return
	}

	defer r.Body.Close()

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		httpError(w, "No body was found : "+err.Error(), http.StatusBadRequest)
	}

	var obj NewUsername
	err = json.Unmarshal(body, &obj)
	if err != nil {
		httpError(w, "json unmarshall error:"+err.Error(), http.StatusBadRequest)
		return
	}

	errorMessage := ""
	if isEmpty(&obj.Id) {
		errorMessage += "Missing paramter 'id'\n"
	}
	if isEmpty(&obj.Username) {
		errorMessage += "Missing paramter 'username'\n"
	}
	if isEmpty(&obj.Password) {
		errorMessage += "Missing paramter 'password'\n"
	}
	if isEmpty(&obj.NewUsername) {
		errorMessage += "Missing paramter 'new_username'\n"
	}
	if !isEmpty(&errorMessage) {
		httpError(w, errorMessage, http.StatusBadRequest)
		return
	}

	db, err := getDatabase(w)
	if err != nil {
		httpError(w, "Database Error :"+err.Error(), http.StatusBadRequest)
		return
	}

	sql := fmt.Sprintf("SELECT u_id FROM users WHERE u_username = '%s'", obj.NewUsername)
	rows, err := db.Query(sql)

	if rows.Next() || err != nil {
		httpError(w, fmt.Sprintf("There is already a user named (%s)", obj.NewUsername), http.StatusBadRequest)
		return
	}

	sql = fmt.Sprintf("UPDATE users SET u_username = '%s' WHERE u_id = '%s' AND u_username = '%s' AND u_password = '%s';", obj.NewUsername, obj.Id, obj.Username, obj.Password)

	res, err := db.Exec(sql)
	if err != nil {
		httpError(w, "Query Error with :"+sql, http.StatusBadRequest)
		return
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		httpError(w, "Effect 0 row", http.StatusBadRequest)
		return
	}

	defer db.Close()
	fmt.Fprintln(w, rowsAffected)
}

func post(w http.ResponseWriter, r *http.Request) {
	if err := checkRequestMethod(r, "post"); err == nil {
		//post a post
		// postPost(w, r)
		if checkRequestMethodReturn(w, r, "post") {
			return
		}

		if err := r.ParseMultipartForm(32 << 20); err != nil {
			httpError(w, err.Error(), http.StatusBadRequest)
			return
		}

		println(r.MultipartForm.File)
		println(r.MultipartForm.Value)

		content := r.MultipartForm.Value["content"][0]
		publisherID := r.MultipartForm.Value["publisher_id"][0]
		postVisibility := r.MultipartForm.Value["post_visibility"][0]
		replyVisibility := r.MultipartForm.Value["reply_visibility"][0]
		tags := strings.Split(r.MultipartForm.Value["tags"][0], "&#10;")
		images := r.MultipartForm.File["post_images"]
		println("content is : " + content)
		println(publisherID)
		println(postVisibility)
		println(replyVisibility)
		println(strings.Join(tags, ","))

		println(images)
		println(len(images))
		for _, header := range images {
			if header.Size > MAX_UPLOAD_SIZE {
				httpError(w, fmt.Sprintf("The uploaded image is too big: %s. Please use an image less than 1MB in size", header.Filename), http.StatusBadRequest)
				return
			}
		}

		sql := fmt.Sprintf("select u_id from users where u_id = '%s'", publisherID)
		db, rows := queryForRows(w, sql)
		if db == nil || rows == nil {
			return
		}
		defer db.Close()
		defer rows.Close()

		if !rows.Next() {
			httpError(w, "Cannot find the publishder ID", http.StatusBadRequest)
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
			tags = deleteEmpty(tags)
			joinedTags = strings.Join(tags, ",")
			sql = fmt.Sprintf("call add_tags('%s')", joinedTags)
			_, err := db.Exec(sql)
			if err != nil {
				httpError(w, "add tags go wrong"+err.Error(), http.StatusInternalServerError)
				return
			}
		}
		// println(joinedTags)

		sql = fmt.Sprintf("INSERT INTO posts (p_publisher_id, p_publish_date, p_edit_date, p_text_content, p_visibility, p_reply, p_images_count, p_tags) VALUES ('%s',NOW(),NOW(),'%s','%s','%s','%d','%s')", publisherID, content, visibility, reply, len(images), joinedTags)
		result, err := db.Exec(sql)
		if err != nil {
			httpError(w, "insert post error"+err.Error(), http.StatusInternalServerError)
			return
		}

		id, err := result.LastInsertId()
		if err != nil {
			httpError(w, "Cannot get last inserted id"+err.Error(), http.StatusInternalServerError)
			return
		}

		for i, header := range images {
			file, err := header.Open()
			if err != nil {
				httpError(w, err.Error(), http.StatusInternalServerError)
				return
			}
			defer file.Close()
			buff := make([]byte, 512)
			print("before: ")
			println(buff)
			_, err = file.Read(buff)
			if err != nil {
				httpError(w, err.Error(), http.StatusInternalServerError)
				return
			}
			print("after: ")
			println(buff)
			filetype := http.DetectContentType(buff)
			println("file type is: " + filetype)
			println(filepath.Ext(header.Filename))
			if filetype != "image/jpeg" && filetype != "image/jpg" && filetype != "image/png" {
				httpError(w, "The provided file format is not allowed. Please upload a JPEG(JPG) or PNG image", http.StatusBadRequest)
				continue
			}
			_, err = file.Seek(0, io.SeekStart)
			if err != nil {
				httpError(w, err.Error(), http.StatusInternalServerError)
				return
			}
			ext := filepath.Ext(header.Filename)
			println(header.Filename + "_" + ext)
			err = os.MkdirAll("./uploads/posts", os.ModePerm)
			if err != nil {
				httpError(w, err.Error(), http.StatusInternalServerError)
				return
			}
			f, err := os.Create(fmt.Sprintf("./uploads/posts/post_%d_%d", id, i))
			if err != nil {
				httpError(w, err.Error(), http.StatusBadRequest)
				return
			}
			defer f.Close()

			pr := &Progress{
				TotalSize: header.Size,
			}

			_, err = io.Copy(f, io.TeeReader(file, pr))
			if err != nil {
				httpError(w, err.Error(), http.StatusBadRequest)
				return
			}

		}

		fmt.Printf("r.MultipartForm.Value: %v\n", r.MultipartForm.Value)
		fmt.Printf("r.MultipartForm.File: %v\n", r.MultipartForm.File)

	} else if err := checkRequestMethod(r, "get"); err == nil {
		query := r.URL.Query()

		if checkMissingParamters(w, query, true, "email", "password", "posts_type") {
			return
		}

		email := query["email"][0]
		password := query["password"][0]
		posts_type := query["posts_type"][0]

		//check user
		sql := fmt.Sprintf("select u_id from users where u_email = '%s' and u_password = '%s'", email, password)
		db, rows := queryForRows(w, sql)
		if db == nil || rows == nil {
			return
		}
		defer db.Close()
		defer rows.Close()

		if !rows.Next() {
			httpError(w, "Cannot find user", http.StatusBadRequest)
			return
		}

		if posts_type == "All" {

		} else if posts_type == "Hot" {

		} else if posts_type == "Random" {

		} else if posts_type == "Following" {

		} else {
			httpError(w, "Cannot find type of :"+posts_type, http.StatusBadRequest)
			return
		}

		sql = "select * from posts order by p_publish_date DESC"
		rows, err := db.Query(sql)
		if err != nil {
			httpError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var posts []Post
		for rows.Next() {
			var post Post
			err = rows.Scan(
				&post.ID,
				&post.PublisherID,
				&post.PublishDate,
				&post.EditDate,
				&post.EditTimes,
				&post.TextContent,
				&post.Deleted,
				&post.ImagesCount,
				&post.Tags,
				&post.Upvotes,
				&post.Downvotes,
				&post.Repost,
				&post.Comment,
				&post.Visibility,
				&post.Reply,
				&post.IsRepost,
				&post.OriginPostID,
				&post.ReposterID,
			)
			if err != nil {
				println("skipping row" + err.Error())
				continue
			}
			posts = append(posts, post)
			fmt.Printf("%+v\n", post)
		}

		if err = rows.Err(); err != nil {
			httpError(w, "Databse Rows Error", http.StatusInternalServerError)
		}

		json := toJson(posts)

		println(json)
		fmt.Fprintln(w, json)
	} else {
		httpError(w, "Only get or post method is allowed", http.StatusMethodNotAllowed)
	}
}

func getPostImage(w http.ResponseWriter, r *http.Request) {
	if checkRequestMethodReturn(w, r, "get") {
		return
	}

	query := r.URL.Query()
	if checkMissingParamters(w, query, true, "id", "index") {
		return
	}

	id := query["id"][0]
	index_str := query["index"][0]

	index, err := strconv.ParseInt(index_str, 0, 8)
	if err != nil {
		httpError(w, fmt.Sprintf("%s is not a valid number", index_str), http.StatusBadRequest)
		return
	}

	if index < 0 || index > 10 {
		httpError(w, fmt.Sprintf("%d is not within range", index), http.StatusBadRequest)
		return
	}

	path := fmt.Sprintf("./uploads/posts/post_%s_%d", id, index)

	bytes := getImage(path)
	if bytes == nil {
		httpError(w, "Cannot find file", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(bytes)
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/", getIndexHandler)                              //get
	mux.HandleFunc("/user", getUser)                                  //get
	mux.HandleFunc("/user/checkExisting", getCheckUserExist)          //get
	mux.HandleFunc("/user/avatar", getAvatar)                         //get
	mux.HandleFunc("/user/update/username", postUpdateUsername)       //post
	mux.HandleFunc("/login", getTryLogin)                             //get
	mux.HandleFunc("/upload/avatar", postUploadAvatar)                //post
	mux.HandleFunc("/validation/email/send", postSendValidationEmail) //post
	mux.HandleFunc("/validation/email/validate", getValidateEmail)    //get
	mux.HandleFunc("/post", post)                                     //post/get
	mux.HandleFunc("/post/images", getPostImage)                      //get

	if err := http.ListenAndServe(":4500", mux); err != nil {
		log.Fatal(err)
	}
}
