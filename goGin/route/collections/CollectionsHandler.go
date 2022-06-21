package collections

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

func CollectionsHandler() (string, ginTools.RouteMap) {
	routeMap := ginTools.NewRouteMap()

	routeMap.AddRoute(ginTools.Route{
		Name:   "/add",
		Fun:    addToCollection_post,
		Method: http.MethodPost,
	})
	routeMap.AddRoute(ginTools.Route{
		Name:   "/remove",
		Fun:    removeCollection_post,
		Method: http.MethodPost,
	})

	routeMap.AddRoute(ginTools.Route{
		Name:   "/",
		Fun:    getCollections_get,
		Method: http.MethodGet,
	})

	// 可以读取url中的参数
	return "/collections", *routeMap
}

func addToCollection_post(context *gin.Context) {
	r := context.Request
	w := context.Writer
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

func getCollections_get(context *gin.Context) {
	r := context.Request
	w := context.Writer
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

	println(sql)

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

func removeCollection_post(context *gin.Context) {
	r := context.Request
	w := context.Writer
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
