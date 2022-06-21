package user

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
	"rainbowwolfer/myspacegoserver/api"
	"rainbowwolfer/myspacegoserver/goGin/ginTools"
	"rainbowwolfer/myspacegoserver/model"
	"strconv"
)

func UserHandler() (string, ginTools.RouteMap) {
	routeMap := ginTools.NewRouteMap()

	routeMap.AddRoute(ginTools.Route{
		Name:   "/",
		Fun:    getUser_get,
		Method: http.MethodGet,
	})
	routeMap.AddRoute(ginTools.Route{
		Name:   "/checkExisting",
		Fun:    checkUserExist_get,
		Method: http.MethodGet,
	})

	routeMap.AddRoute(ginTools.Route{
		Name:   "/avatar",
		Fun:    getAvatar_get,
		Method: http.MethodGet,
	})
	routeMap.AddRoute(ginTools.Route{
		Name:   "/update/username",
		Fun:    updateUsername_post,
		Method: http.MethodPost,
	})
	routeMap.AddRoute(ginTools.Route{
		Name:   "/follow",
		Fun:    userFollow_post,
		Method: http.MethodPost,
	})

	routeMap.AddRoute(ginTools.Route{
		Name:   "/getFollowers",
		Fun:    getUserFollowers_get,
		Method: http.MethodGet,
	})
	routeMap.AddRoute(ginTools.Route{
		Name:   "/postsAndFollowersCount",
		Fun:    getPostsAndFollowersCount_get,
		Method: http.MethodGet,
	})

	// 可以读取url中的参数
	return "/user", *routeMap
}

func getUser_get(context *gin.Context) {
	r := context.Request
	w := context.Writer
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

func getAvatar_get(context *gin.Context) {
	r := context.Request
	w := context.Writer
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

	// println(bytes)
	// println(bytes == nil)
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(bytes)
}

//Update Username - Post
//Error Code ->
//1-username taken
func updateUsername_post(context *gin.Context) {
	r := context.Request
	w := context.Writer
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
func getPostsAndFollowersCount_get(context *gin.Context) {
	r := context.Request
	w := context.Writer
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

func checkUserExist_get(context *gin.Context) {
	r := context.Request
	w := context.Writer
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

func getUserFollowers_get(context *gin.Context) {
	r := context.Request
	w := context.Writer
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
func userFollow_post(context *gin.Context) {
	r := context.Request
	w := context.Writer
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
	}
	var sql string
	if obj.Cancel {
		sql = fmt.Sprintf("DELETE FROM users_follows WHERE uf_id_follower = %d and uf_id_target = %s;", user_id, obj.TargetID)
	} else {
		sql = fmt.Sprintf("insert into users_follows (uf_id_follower, uf_id_target) values (%d,%s);", user_id, obj.TargetID)
	}

	println(sql)
	_, err = database.Exec(sql)
	if err != nil {
		api.HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprint(w, "Success")
}
