package route

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"rainbowwolfer/myspacegoserver/api"
	"rainbowwolfer/myspacegoserver/goGin/ginTools"
)

func IndexHandler() ginTools.RouteMap {
	routeMap := ginTools.NewRouteMap()

	routeMap.AddRoute(ginTools.Route{
		Name:   "/",
		Fun:    getIndexHandler,
		Method: http.MethodPost,
	})
	// 可以读取url中的参数
	return *routeMap
}
func getIndexHandler(context *gin.Context) {
	r := context.Request
	w := context.Writer
	if api.CheckRequestMethodReturn(w, r, "get") {
		return
	}
	w.Header().Add("Content-Type", "text/html")
	http.ServeFile(w, r, "index.html")
}
