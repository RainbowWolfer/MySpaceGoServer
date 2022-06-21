package route

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"rainbowwolfer/myspacegoserver/goGin/ginTools"
)

func IndexHandler() *ginTools.RouteMap {
	routeMap := ginTools.NewRouteMap()

	routeMap.AddRoute(ginTools.Route{
		Name:   "/",
		Fun:    getIndexHandler,
		Method: http.MethodGet,
	})
	// 可以读取url中的参数
	return routeMap
}
func getIndexHandler(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", gin.H{})
}
