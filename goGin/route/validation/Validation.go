package validation

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
	"net/mail"
	"net/smtp"
	"rainbowwolfer/myspacegoserver/api"
	"rainbowwolfer/myspacegoserver/goGin/config"
	"rainbowwolfer/myspacegoserver/goGin/ginTools"
)

func ValidationHandler() (string, ginTools.RouteMap) {
	routeMap := ginTools.NewRouteMap()

	routeMap.AddRoute(ginTools.Route{
		Name:   "/email/send",
		Fun:    sendValidationEmail_post,
		Method: http.MethodPost,
	})
	routeMap.AddRoute(ginTools.Route{
		Name:   "/email/validate",
		Fun:    validateEmail_get,
		Method: http.MethodGet,
	})

	// 可以读取url中的参数
	return "/validation", *routeMap
}

//Send Validation Email - Post
//Error Code ->
//1-email or username used
//2-already sent a email
func sendValidationEmail_post(context *gin.Context) {
	r := context.Request
	w := context.Writer
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

	//t, _ := template.ParseFiles("email_validation.html")
	var body bytes.Buffer

	mimeHeaders := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	body.Write([]byte(fmt.Sprintf("Subject: My Space Registration Email Validation \n%s\n\n", mimeHeaders)))

	url := fmt.Sprintf("%s:%d/validation/email/validate?email=%s&code=%s", config.HOST, config.PORT, email_query, code)

	context.HTML(http.StatusOK, "email_validation.html", gin.H{
		"Name": username_query,
		"Link": url,
	})

	/*	t.Execute(&body, struct {
			Name string
			Link string
		}{
			Name: username_query,
			Link: url,
		})
	*/
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

func validateEmail_get(context *gin.Context) {

	r := context.Request
	w := context.Writer
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
