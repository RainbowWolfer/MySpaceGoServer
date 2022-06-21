package route

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"rainbowwolfer/myspacegoserver/api"
	"rainbowwolfer/myspacegoserver/goGin/ginTools"
	"rainbowwolfer/myspacegoserver/model"
)

func LoginHandler() *ginTools.RouteMap {
	routeMap := ginTools.NewRouteMap()

	routeMap.AddRoute(ginTools.Route{
		Name:   "/login",
		Fun:    tryLogin_get,
		Method: http.MethodGet,
	})
	// 可以读取url中的参数
	return routeMap
}

//Login - Get
//Error Code ->
//1-registration pending
//2-user not found (email or password is wrong)
func tryLogin_get(context *gin.Context) {
	r := context.Request
	w := context.Writer
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
