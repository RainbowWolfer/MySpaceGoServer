package main

import (
	"bytes"
	_ "context"
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/mail"
	"net/smtp"
	"net/url"
	"os"
	"path/filepath"
	_ "regexp"
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

func now() string {
	dt := time.Now()
	//Format MM-DD-YYYY hh:mm:ss
	return dt.Format("2006-01-02 15:04:05")
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
	ID                 int
	Username           string
	Password           string
	Email              string
	ProfileDescription string
}

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
	OriginScore        *int
	OriginVoted        *int
}

type Comment struct {
	ID          string
	UserID      string
	PostID      string
	TextContent string
	DateTime    string
	Upvotes     int
	Downvotes   int
}

type NewVote struct {
	PostID   string `json:"post_id"`
	UserID   string `json:"user_id"`
	Cancel   bool   `json:"cancel"`
	Score    int    `json:"score"`
	Email    string `json:"email"`
	Password string `json:"password"`
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

func readUser(rows *sql.Rows) (User, error) {
	var user User
	if err := rows.Scan(
		&user.ID,
		&user.Username,
		&user.Password,
		&user.Email,
		&user.ProfileDescription,
	); err != nil {
		return User{}, errors.New("User Convert Error" + err.Error())
	}
	return user, nil
}

func readPost(rows *sql.Rows) (Post, error) {
	var post Post
	err := rows.Scan(
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
		&post.OriginScore,
		&post.OriginVoted,
	)
	return post, err
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

func getDatabase() (*sql.DB, error) {
	var err error
	database, err := sql.Open("mysql", "wjx:123456@tcp(www.cqtest.top:3306)/wjx")
	println(fmt.Sprintf("Connection in use %d", database.Stats().InUse))
	println("Open new Database Connection")
	if err != nil {
		return nil, err
	}
	database.SetConnMaxLifetime(time.Second * 10)
	database.SetMaxOpenConns(500)
	database.SetMaxIdleConns(500)
	return database, nil
}

func checkRequestMethodReturn(w http.ResponseWriter, r *http.Request, method string) bool {
	// println(r.URL.RawPath)
	if !strings.EqualFold(r.Method, method) {
		httpError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return true
	} else {
		return false
	}
}

func checkRequestMethod(r *http.Request, method string) error {
	// println(r.URL.RawPath)
	if strings.EqualFold(r.Method, method) {
		return nil
	} else {
		return errors.New("method not allowed")
	}
}

func isEmpty(str *string) bool {
	if str == nil {
		return true
	}
	if len(*str) == 0 {
		return true
	}
	if len(strings.TrimSpace(*str)) == 0 {
		return true
	}

	return false
}

func checkUser(db *sql.DB, email string, pasword string) (int, error) {
	sql := fmt.Sprintf("select u_id from users where u_email = '%s' and u_password = '%s';", email, pasword)
	println(sql)
	rows, err := db.Query(sql)
	if err != nil {
		return -1, err
	}
	if !rows.Next() {
		return -1, nil
	}
	var userID int
	err = rows.Scan(&userID)
	if err != nil {
		return -1, err
	}
	return userID, nil
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

func getImage(path string) []byte {
	availableExts := []string{"", ".png", ".jpg", ".jpeg"}
	var bytes []byte = nil
	for _, ext := range availableExts {
		p := fmt.Sprintf("%s%s", path, ext)
		// println("getting file: " + p)
		fileBytes, err := ioutil.ReadFile(p)
		if err != nil {
			// println("not found")
			continue
		}
		// println("found!")
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

	database, err := getDatabase()
	if err != nil {
		httpError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer database.Close()

	rows, err := database.Query(sql)
	if err != nil {
		httpError(w, "Query Error with :"+sql+"\n"+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	if !rows.Next() {
		sql = fmt.Sprintf("SELECT ev_id FROM email_validations WHERE ev_email = '%s' AND ev_password = '%s'", email, password)
		rows_ev, err := database.Query(sql)
		if err != nil {
			httpError(w, "Query Error: "+err.Error(), http.StatusInternalServerError)
			return
		}

		if rows_ev.Next() {
			httpErrorWithCode(w, "User is in registration pending", http.StatusBadRequest, 1)
			return
		} else {
			httpErrorWithCode(w, "No User Found", http.StatusBadRequest, 2)
			return //no found result
		}
	}

	user, err := readUser(rows)

	if err != nil {
		httpError(w, err.Error(), http.StatusInternalServerError)
		return
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

	database, err := getDatabase()
	if err != nil {
		httpError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer database.Close()

	rows, err := database.Query(sql)
	if err != nil {
		httpError(w, "Query Error with :"+sql+"\n"+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	if !rows.Next() {
		httpError(w, "Cannot find user with "+errorMsg, http.StatusBadRequest)
		return
	}

	user, err := readUser(rows)

	if err != nil {
		httpError(w, err.Error(), http.StatusInternalServerError)
	}

	if err := rows.Err(); err != nil {
		httpError(w, "Databse Rows Error"+err.Error(), http.StatusInternalServerError)
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

	// println(buff)

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

	database, err := getDatabase()
	if err != nil {
		httpError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer database.Close()

	sql := fmt.Sprintf("select u_id from users where u_id = '%s'", id)
	rows, err := database.Query(sql)
	if err != nil {
		httpError(w, "Query Error with :"+sql+"\n"+err.Error(), http.StatusInternalServerError)
		return
	}
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
	// println(f.Name())

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

	// println(bytes)
	// println(bytes == nil)
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

	database, err := getDatabase()
	if err != nil {
		httpError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer database.Close()

	rows, err := database.Query(sql)
	if err != nil {
		httpError(w, "Query Error with :"+sql+"\n"+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	if rows.Next() {
		httpErrorWithCode(w, fmt.Sprintf("email (%s) or username (%s) is used for another account.", email_query, username_query), http.StatusBadRequest, 1)
		return
	}

	sql = fmt.Sprintf("select ev_id FROM email_validations where ev_email = '%s'", email_query)
	rows, err = database.Query(sql)
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
	_, err = database.Exec(sql)
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

	database, err := getDatabase()
	if err != nil {
		httpError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer database.Close()

	rows, err := database.Query(sql)
	if err != nil {
		httpError(w, "Query Error with :"+sql+"\n"+err.Error(), http.StatusInternalServerError)
		return
	}
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
	_, err = database.Exec(sql)
	if err != nil {
		httpError(w, "database delete validation failed\n"+err.Error(), http.StatusInternalServerError)
		return
	}

	//add new user
	sql = fmt.Sprintf("INSERT INTO users (u_username,u_password,u_email) VALUES ('%s','%s','%s')", db_username, db_password, db_email)
	_, err = database.Exec(sql)
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

	database, err := getDatabase()
	if err != nil {
		httpError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer database.Close()

	rows, err := database.Query(sql)
	if err != nil {
		httpError(w, "Query Error with :"+sql+"\n"+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	foundUser := rows.Next()

	fmt.Fprintln(w, foundUser)
}

type NewUsername struct {
	ID          string `json:"id"`
	Username    string `json:"username"`
	Password    string `json:"password"`
	NewUsername string `json:"new_username"`
}
type NewComment struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	PostID      string `json:"post_id"`
	TextContent string `json:"text_content"`
}

//Update Username - Post
//Error Code ->
//1-username taken
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
	if isEmpty(&obj.ID) {
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

	database, err := getDatabase()
	if err != nil {
		httpError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer database.Close()

	sql := fmt.Sprintf("SELECT u_id FROM users WHERE u_username = '%s'", obj.NewUsername)
	rows, err := database.Query(sql)
	if err != nil {
		httpError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	if rows.Next() {
		httpErrorWithCode(w, fmt.Sprintf("There is already a user named (%s)", obj.NewUsername), http.StatusBadRequest, 1)
		return
	}

	sql = fmt.Sprintf("UPDATE users SET u_username = '%s' WHERE u_id = '%s' AND u_username = '%s' AND u_password = '%s';", obj.NewUsername, obj.ID, obj.Username, obj.Password)

	res, err := database.Exec(sql)
	if err != nil {
		httpError(w, "Query Error with :"+sql, http.StatusBadRequest)
		return
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		httpError(w, "Effect 0 row", http.StatusBadRequest)
		return
	}

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
			if header.Size > MAX_UPLOAD_SIZE {
				httpError(w, fmt.Sprintf("The uploaded image is too big: %s. Please use an image less than 1MB in size", header.Filename), http.StatusBadRequest)
				return
			}
		}

		database, err := getDatabase()
		if err != nil {
			httpError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer database.Close()

		sql := fmt.Sprintf("select u_id from users where u_id = '%s'", publisherID)
		rows, err := database.Query(sql)
		if err != nil {
			httpError(w, "Query Error with :"+sql+"\n"+err.Error(), http.StatusInternalServerError)
			return
		}
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
			_, err := database.Exec(sql)
			if err != nil {
				httpError(w, "add tags go wrong"+err.Error(), http.StatusInternalServerError)
				return
			}
		}
		// println(joinedTags)

		sql = fmt.Sprintf("INSERT INTO posts (p_publisher_id, p_publish_date, p_edit_date, p_text_content, p_visibility, p_reply, p_images_count, p_tags) VALUES ('%s',NOW(),NOW(),'%s','%s','%s','%d','%s')", publisherID, content, visibility, reply, len(images), joinedTags)
		result, err := database.Exec(sql)
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
			// print("before: ")
			// println(buff)
			_, err = file.Read(buff)
			if err != nil {
				httpError(w, err.Error(), http.StatusInternalServerError)
				return
			}
			// print("after: ")
			// println(buff)
			filetype := http.DetectContentType(buff)
			// println("file type is: " + filetype)
			// println(filepath.Ext(header.Filename))
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

		if checkMissingParamters(w, query, true, "posts_type", "offset") {
			return
		}

		posts_type := query["posts_type"][0]
		offset := query["offset"][0]

		limit := 10
		if query.Has("limit") {
			limit, err = strconv.Atoi(query["limit"][0])
			if err != nil {
				httpError(w, "limit is not a valid number"+err.Error(), http.StatusBadRequest)
				return
			}
		}

		database, err := getDatabase()
		if err != nil {
			httpError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer database.Close()

		println(database.Stats().Idle)
		println(database.Stats().InUse)

		println(database.Stats().OpenConnections)
		println(database.Stats().WaitDuration)

		user_id := -1
		if query.Has("email") && query.Has("password") {
			email := query["email"][0]
			password := query["password"][0]
			if !isEmpty(&email) && !isEmpty(&password) {
				sql_user := fmt.Sprintf("select u_id from users where u_email = '%s' and u_password = '%s'", email, password)
				rows, err := database.Query(sql_user)
				if err != nil {
					httpError(w, "query user error"+err.Error(), http.StatusInternalServerError)
					return
				}
				defer rows.Close()
				if !rows.Next() {
					httpError(w, "Cannot find user", http.StatusBadRequest)
					return
				}

				if err = rows.Scan(&user_id); err != nil {
					httpError(w, err.Error(), http.StatusInternalServerError)
					return
				}
			}
		}

		var sql string

		if strings.EqualFold(posts_type, "All") { //time based
			sql = fmt.Sprintf("CALL GetPostsByTime(%d,%s,%d);", user_id, offset, limit)
		} else if strings.EqualFold(posts_type, "Hot") {
			sql = fmt.Sprintf("CALL GetPostsByScore(%d,%s,%d)", user_id, offset, limit)
		} else if strings.EqualFold(posts_type, "Random") {
			if !query.Has("seed") {
				httpError(w, "require parameter 'seed'", http.StatusBadRequest)
				return
			}
			seed := query["seed"][0]
			sql = fmt.Sprintf("CALL GetPostsByRandom(%d,%s,%d,%s)", user_id, offset, limit, seed)
		} else if strings.EqualFold(posts_type, "Following") {
			if user_id == 0 || user_id == -1 {
				httpError(w, "user not found", http.StatusBadRequest)
				return
			}
			sql = fmt.Sprintf("CALL GetPostsByFollow(%d,%s,%d)", user_id, offset, limit)
		} else {
			httpError(w, "Cannot find type of :"+posts_type, http.StatusBadRequest)
			return
		}

		println(sql)
		rows, err := database.Query(sql)
		println("END")
		if err != nil {
			httpError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var posts []Post
		for rows.Next() {
			post, err := readPost(rows)
			if err != nil {
				println("skipping row" + err.Error())
				continue
			}
			posts = append(posts, post)
			if post.IsRepost {
				fmt.Printf("%+v\n\n\n", post)
			}
		}

		if err = rows.Err(); err != nil {
			httpError(w, "Databse Rows Error", http.StatusInternalServerError)
		}

		json := toJson(posts)

		// println(json)
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

func clearUnusedPostImages(w http.ResponseWriter, r *http.Request) {
	if checkRequestMethodReturn(w, r, "get") {
		return
	}
	query := r.URL.Query()
	if checkMissingParamters(w, query, true, "key") {
		return
	}

}

func reinflateDefaultPosts(w http.ResponseWriter, r *http.Request) {
	rand.Seed(time.Now().UnixNano())
	if checkRequestMethodReturn(w, r, "get") {
		return
	}
	query := r.URL.Query()
	if checkMissingParamters(w, query, true, "key") {
		return
	}
	key := query["key"][0]
	if key != "eb9f60e5c17ec16a7dfbf79321b79afa" {
		httpError(w, "key error", http.StatusBadRequest)
		return
	}

	database, err := getDatabase()
	if err != nil {
		httpError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer database.Close()

	_, err = database.Exec("DELETE FROM posts;")
	if err != nil {
		httpError(w, "delete error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = database.Exec("DELETE FROM users;")
	if err != nil {
		httpError(w, "delete error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	sql := "INSERT INTO users VALUES (1,'myspace','myspace','RainbowWolfer@outlook.com','This is official account for MySpace. Feel free to tell us what improvoments should be made or just come small talking. All are welcomed!');"
	// println(sql)
	_, err = database.Exec(sql)

	if err != nil {
		httpError(w, "insert user error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	usersCount := rand.Intn(20) + 2

	for i := 2; i < usersCount; i++ {
		sql = fmt.Sprintf("INSERT INTO users VALUES (%d,'Test Dummy %d','123456','%d@test.com','Test Dummy #%d');", i, i, i, i)
		// println(sql)
		_, err = database.Exec(sql)

		if err != nil {
			httpError(w, "insert user error: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	list := []string{
		"Lorem Ipsum is simply dummy text of the printing and typesetting industry. Lorem Ipsum has been the industry\\'s standard dummy text ever since the 1500s, when an unknown printer took a galley of type and scrambled it to make a type specimen book.",
		" It has survived not only five centuries, but also the leap into electronic typesetting, remaining essentially unchanged. It was popularised in the 1960s with the release of Letraset sheets containing Lorem Ipsum passages, and more recently with desktop publishing software like Aldus PageMaker including versions of Lorem Ipsum.",
		"It is a long established fact that a reader will be distracted by the readable content of a page when looking at its layout. ",
		"The point of using Lorem Ipsum is that it has a more-or-less normal distribution of letters, as opposed to using \"Content here, content here\", making it look like readable English. ",
		"Many desktop publishing packages and web page editors now use Lorem Ipsum as their default model text, and a search for \"lorem ipsum\" will uncover many web sites still in their infancy. ",
		"Various versions have evolved over the years, sometimes by accident, sometimes on purpose (injected humour and the like).",
		"Contrary to popular belief, Lorem Ipsum is not simply random text. It has roots in a piece of classical Latin literature from 45 BC, making it over 2000 years old. Richard McClintock, a Latin professor at Hampden-Sydney College in Virginia, looked up one of the more obscure Latin words, consectetur, from a Lorem Ipsum passage, and going through the cites of the word in classical literature, discovered the undoubtable source.",
		"Lorem Ipsum comes from sections 1.10.32 and 1.10.33 of \"de Finibus Bonorum et Malorum\" (The Extremes of Good and Evil) by Cicero, written in 45 BC. This book is a treatise on the theory of ethics, very popular during the Renaissance. The first line of Lorem Ipsum, \"Lorem ipsum dolor sit amet..\", comes from a line in section 1.10.32.",
		"The standard chunk of Lorem Ipsum used since the 1500s is reproduced below for those interested.",
		"Sections 1.10.32 and 1.10.33 from \"de Finibus Bonorum et Malorum\" by Cicero are also reproduced in their exact original form, accompanied by English versions from the 1914 translation by H. Rackham.",
		"Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed vitae finibus metus. Etiam ultrices magna vitae ligula sodales suscipit. Sed eu nibh in dolor pharetra varius vitae sit amet dolor.",
		"Morbi auctor pharetra ipsum vitae tempus. Sed risus risus, iaculis eget sapien eu, suscipit vulputate nulla.",
		"Donec vel purus non lacus euismod imperdiet eu tempus sapien. Donec non dui sed odio eleifend sagittis quis commodo enim. Proin nec magna sem.",
	}

	random := rand.Intn(100)
	println(random)
	for i := 1; i < random+100; i++ {
		text := list[rand.Intn(len(list))]
		publisher_id := rand.Intn(usersCount-1) + 1
		sql = fmt.Sprintf("INSERT INTO posts VALUES (%d,%d, TIMESTAMPADD(SECOND,%d,NOW()), TIMESTAMPADD(SECOND,%d,NOW()),0,'%s',FALSE,0,'official,test,LoremIpsum',%d,%d,0,0,'all','all',FALSE,-1,-1);", i, publisher_id, i, i, text, rand.Intn(500), rand.Intn(100))
		// println(sql)
		_, err = database.Exec(sql)

		if err != nil {
			httpError(w, "insert data error: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}
	fmt.Fprintf(w, "Successfully infalte default data (%d) with users (%d)", random+100, usersCount-1)
}

func comment(w http.ResponseWriter, r *http.Request) {
	if err := checkRequestMethod(r, "post"); err == nil {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			httpError(w, "No body was found : "+err.Error(), http.StatusBadRequest)
		}
		println(string(body))

		var obj NewComment
		err = json.Unmarshal(body, &obj)
		if err != nil {
			httpError(w, "json unmarshall error:"+err.Error(), http.StatusBadRequest)
			return
		}

		errorMessage := ""
		if isEmpty(&obj.Email) {
			errorMessage += "Missing paramter 'email'\n"
		}
		if isEmpty(&obj.Password) {
			errorMessage += "Missing paramter 'password'\n"
		}
		if isEmpty(&obj.PostID) {
			errorMessage += "Missing paramter 'post_id'\n"
		}
		if isEmpty(&obj.TextContent) {
			errorMessage += "Missing paramter 'text_content'\n"
		}
		if !isEmpty(&errorMessage) {
			httpError(w, errorMessage, http.StatusBadRequest)
			return
		}

		database, err := getDatabase()
		if err != nil {
			httpError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer database.Close()

		//check user valid
		sql := fmt.Sprintf("select u_id from users where u_email = '%s' and u_password = '%s';", obj.Email, obj.Password)

		rows, err := database.Query(sql)
		if err != nil {
			httpError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		if !rows.Next() {
			httpError(w, "user not found", http.StatusBadRequest)
			return
		}

		var userID int
		err = rows.Scan(&userID)
		if err != nil {
			httpError(w, "user id scan error"+err.Error(), http.StatusInternalServerError)
			return
		}

		//add comment

		sql = fmt.Sprintf("INSERT INTO comments (c_id_user, c_id_post, c_text_content, c_datetime) VALUES ('%d','%s','%s',NOW());", userID, obj.PostID, obj.TextContent)

		result, err := database.Exec(sql)

		if err != nil {
			httpError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		last_id, err := result.LastInsertId()
		if err != nil {
			httpError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		fmt.Fprintln(w, toJson(Comment{
			ID:          fmt.Sprintf("%d", last_id),
			UserID:      fmt.Sprintf("%d", userID),
			PostID:      obj.PostID,
			TextContent: obj.TextContent,
			DateTime:    now(),
			Upvotes:     0,
			Downvotes:   0,
		}))

	} else if err := checkRequestMethod(r, "get"); err == nil {
		query := r.URL.Query()
		if checkMissingParamters(w, query, true, "id") {
			return
		}
		id := query["id"][0]

		database, err := getDatabase()
		if err != nil {
			httpError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer database.Close()

		sql := fmt.Sprintf("select * from comments where c_id = '%s'", id)

		rows, err := database.Query(sql)
		if err != nil {
			httpError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		if !rows.Next() {
			httpError(w, "not comment found", http.StatusBadRequest)
			return
		}

		var comment Comment
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
			httpError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		fmt.Fprintln(w, toJson(comment))
	} else {
		httpError(w, "Only get or post method is allowed", http.StatusMethodNotAllowed)
	}
}

func getPostComments(w http.ResponseWriter, r *http.Request) {
	if checkRequestMethodReturn(w, r, "get") {
		return
	}
	query := r.URL.Query()
	if checkMissingParamters(w, query, true, "post_id") {
		return
	}

	postID := query["post_id"][0]

	database, err := getDatabase()
	if err != nil {
		httpError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer database.Close()

	sql := fmt.Sprintf("select * from comments where c_id_post = '%s'", postID)

	rows, err := database.Query(sql)
	if err != nil {
		httpError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var list []Comment

	for rows.Next() {
		var item Comment
		err = rows.Scan(
			&item.ID,
			&item.UserID,
			&item.PostID,
			&item.TextContent,
			&item.DateTime,
			&item.Upvotes,
			&item.Downvotes,
		)

		if err != nil {
			httpError(w, err.Error(), http.StatusInternalServerError)
			continue
		}
		list = append(list, item)
	}

	fmt.Fprintln(w, toJson(list))
}

func postVote(w http.ResponseWriter, r *http.Request) {
	if checkRequestMethodReturn(w, r, "post") {
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		httpError(w, "No body was found : "+err.Error(), http.StatusBadRequest)
	}

	var obj NewVote
	err = json.Unmarshal(body, &obj)
	if err != nil {
		httpError(w, "json unmarshall error:"+err.Error(), http.StatusBadRequest)
		return
	}

	errorMessage := ""
	if isEmpty(&obj.PostID) {
		errorMessage += "Missing paramter 'post_id'\n"
	}
	if isEmpty(&obj.UserID) {
		errorMessage += "Missing paramter 'user_id'\n"
	}
	if isEmpty(&obj.Email) {
		errorMessage += "Missing paramter 'email'\n"
	}
	if isEmpty(&obj.Password) {
		errorMessage += "Missing paramter 'password'\n"
	}
	if !isEmpty(&errorMessage) {
		httpError(w, errorMessage, http.StatusBadRequest)
		return
	}

	database, err := getDatabase()
	if err != nil {
		httpError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer database.Close()

	userID, err := checkUser(database, obj.Email, obj.Password)
	if err != nil {
		httpError(w, "During User Check "+err.Error(), http.StatusInternalServerError)
		return
	} else if userID == -1 {
		httpError(w, "No Uesr Found", http.StatusBadRequest)
		return
	}

	sql := fmt.Sprintf("CALL VotePost(%s,%s,%t,%d)", obj.UserID, obj.PostID, obj.Cancel, obj.Score)

	rows, err := database.Query(sql)
	if err != nil {
		httpError(w, "During SQL : "+sql+" : "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var line string
	rows.Scan(&line)

	fmt.Fprint(w, line)
}

func repost(w http.ResponseWriter, r *http.Request) {
	if checkRequestMethodReturn(w, r, "post") {
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		httpError(w, "No body was found : "+err.Error(), http.StatusBadRequest)
	}

	var obj NewRepost
	err = json.Unmarshal(body, &obj)
	if err != nil {
		httpError(w, "json unmarshall error:"+err.Error(), http.StatusBadRequest)
		return
	}

	println(fmt.Sprintf("%v", obj))

	errorMessage := ""
	if isEmpty(&obj.OriginPostID) {
		errorMessage += "Missing paramter 'origin_post_id'\n"
	}
	if isEmpty(&obj.PostVisibility) {
		errorMessage += "Missing paramter 'post_visibility'\n"
	}
	if isEmpty(&obj.PublisherID) {
		errorMessage += "Missing paramter 'publisher_id'\n"
	}
	if isEmpty(&obj.ReplyLimit) {
		errorMessage += "Missing paramter 'reply_limit'\n"
	}
	if isEmpty(&obj.TextContent) {
		errorMessage += "Missing paramter 'text_content'\n"
	}
	if isEmpty(&obj.Email) {
		errorMessage += "Missing paramter 'email'\n"
	}
	if isEmpty(&obj.Password) {
		errorMessage += "Missing paramter 'password'\n"
	}
	if !isEmpty(&errorMessage) {
		httpError(w, errorMessage, http.StatusBadRequest)
		return
	}

	database, err := getDatabase()
	if err != nil {
		httpError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer database.Close()

	//check user
	sql := fmt.Sprintf("select u_id from users where u_email = '%s' and u_password = '%s'", obj.Email, obj.Password)
	rows, err := database.Query(sql)
	if err != nil {
		httpError(w, err.Error(), http.StatusInternalServerError)
		return
	} else if !rows.Next() {
		httpError(w, "user not found", http.StatusBadRequest)
		return
	}

	joinedTags := ""
	println(obj.Tags)
	if len(obj.Tags) != 0 {
		obj.Tags = deleteEmpty(obj.Tags)
		joinedTags = strings.Join(obj.Tags, ",")
		sql = fmt.Sprintf("call add_tags('%s')", joinedTags)
		_, err := database.Exec(sql)
		if err != nil {
			httpError(w, "add tags go wrong"+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	sql = fmt.Sprintf("INSERT INTO posts (p_publisher_id, p_publish_date, p_edit_date, p_text_content, p_visibility, p_reply, p_images_count, p_tags, p_is_repost, p_id_origin_post) VALUES ('%s',NOW(),NOW(),'%s','%s','%s',0,'%s',TRUE,'%s')", obj.PublisherID, obj.TextContent, obj.PostVisibility, obj.ReplyLimit, obj.Tags, obj.OriginPostID)
	println(sql)

	_, err = database.Exec(sql)
	if err != nil {
		httpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

}

func main() {
	println(now())
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
	mux.HandleFunc("/post/comment", comment)                          //post/get
	mux.HandleFunc("/post/comments", getPostComments)                 //get
	mux.HandleFunc("/post/vote", postVote)                            //post
	mux.HandleFunc("/post/repost", repost)                            //post

	mux.HandleFunc("/admin/clearunusedpostimages", clearUnusedPostImages)
	mux.HandleFunc("/admin/reinflatedefaultposts", reinflateDefaultPosts)

	if err := http.ListenAndServe(":4500", mux); err != nil {
		log.Fatal(err)
	}
}
