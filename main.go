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
	_ "strconv"
	"strings"
	"text/template"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

const (
	MAX_UPLOAD_SIZE = 1024 * 1024
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

func getUserHandler(w http.ResponseWriter, r *http.Request) {
	if checkRequestMethodReturn(w, r, "get") {
		return
	}

	query := r.URL.Query()
	if checkMissingParamters(w, query, false, "id", "username") {
		return
	}

	var sql string
	if query.Has("id") {
		id := query["id"][0]
		sql = fmt.Sprintf("SELECT * FROM users where u_id = '%s'", id)
	} else if query.Has("username") {
		username := query["username"][0]
		sql = fmt.Sprintf("SELECT * FROM users where u_username = '%s'", username)
	}

	db, rows := queryForRows(w, sql)
	if db == nil || rows == nil {
		return
	}
	defer db.Close()
	defer rows.Close()

	var users []User

	for rows.Next() {
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
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		httpError(w, "Databse Rows Error", http.StatusInternalServerError)
	}

	fmt.Fprintln(w, toJson(users))
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

	availableExts := []string{".png", ".jpg", ".jpeg", ""}
	var bytes []byte
	for _, ext := range availableExts {
		path := fmt.Sprintf("./uploads/avatars/user_%s%s", id, ext)
		fileBytes, err := ioutil.ReadFile(path)
		if err != nil {
			println("Continue" + ext + "\n" + err.Error())
			continue
		}
		println("break" + ext)
		bytes = fileBytes
		break
	}
	if bytes == nil {
		fileBytes, err := ioutil.ReadFile("./uploads/avatars/DefaultAvatar.png")
		if err != nil {
			httpError(w, "Default File Not Found", http.StatusInternalServerError)
			return
		}
		bytes = fileBytes
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

		content := r.MultipartForm.Value["content"][0]
		images := r.MultipartForm.File["post_images"]
		println("content is : " + content)

		println(images)
		println(len(images))
		for i, header := range images {
			// if v.Size > MAX_UPLOAD_SIZE {

			// }
			file, err := header.Open()
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
			f, err := os.Create(fmt.Sprintf("./uploads/posts/post_%d%s", i, filepath.Ext(header.Filename)))
			if err != nil {
				httpError(w, err.Error(), http.StatusBadRequest)
				return
			}
			defer f.Close()
		}

		fmt.Printf("r.MultipartForm.Value: %v\n", r.MultipartForm.Value)
		fmt.Printf("r.MultipartForm.File: %v\n", r.MultipartForm.File)

	} else if err := checkRequestMethod(r, "get"); err == nil {
		//get a post
		getPost(w, r)
	} else {
		httpError(w, "Only get or post method is allowed", http.StatusMethodNotAllowed)
	}
}

func postPost(w http.ResponseWriter, r *http.Request) {
	if checkRequestMethodReturn(w, r, "post") {
		return
	}

	postBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		httpError(w, "No body was found : "+err.Error(), http.StatusBadRequest)
	}

	println(string(postBody[0:200]))
	fmt.Printf("r.MultipartForm.Value: %v\n", r.MultipartForm.Value)
	fmt.Printf("r.MultipartForm.Value: %v\n", r.MultipartForm.Value)
	// content := r.MultipartForm.Value["content"]
	// images := r.MultipartForm.Value["post_images"]
	// println(content)
	// println(images)
}

func getPost(w http.ResponseWriter, r *http.Request) {
	if checkRequestMethodReturn(w, r, "get") {
		return
	}

}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/", getIndexHandler)                              //get
	mux.HandleFunc("/user", getUserHandler)                           //get
	mux.HandleFunc("/user/checkExisting", getCheckUserExist)          //get
	mux.HandleFunc("/user/avatar", getAvatar)                         //get
	mux.HandleFunc("/user/update/username", postUpdateUsername)       //post
	mux.HandleFunc("/login", getTryLogin)                             //get
	mux.HandleFunc("/upload/avatar", postUploadAvatar)                //post
	mux.HandleFunc("/validation/email/send", postSendValidationEmail) //post
	mux.HandleFunc("/validation/email/validate", getValidateEmail)    //get
	mux.HandleFunc("/post", post)                                     //post/get

	if err := http.ListenAndServe(":4500", mux); err != nil {
		log.Fatal(err)
	}
}
