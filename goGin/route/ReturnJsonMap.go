package route

import (
	"GoWeb/domain"
	"GoWeb/goGin/ginTools"
	"github.com/gin-gonic/gin"
	"net/http"
)

func ReturnJsonMap() (string, *ginTools.RouteMap) {
	routeMap := ginTools.NewRouteMap()

	// return Json map
	routeMap.AddRoute(ginTools.Route{
		Name: "/returnMap",
		Fun: func(context *gin.Context) {
			context.JSON(200, map[string]interface{}{
				"name": "as",
				"age":  12,
			})
		},
		Method: http.MethodGet,
	})

	// return Json
	routeMap.AddRoute(ginTools.Route{
		Name: "/returnJson",
		Fun: func(context *gin.Context) {
			context.JSON(200, gin.H{
				"name": "a2s",
				"age":  1234,
			})
		},
		Method: http.MethodGet,
	})

	// return struct
	routeMap.AddRoute(ginTools.Route{
		Name: "/returnStruct",
		Fun: func(context *gin.Context) {
			user := &domain.User{Name: "As", Addr: "鞍山市", Age: 12}
			context.JSON(200, &user)
		},
		Method: http.MethodGet,
	})

	// 可以读取url中的参数
	routeMap.AddRoute(ginTools.Route{
		Name: "/returnStructP",
		Fun: func(context *gin.Context) {
			user := &domain.User{Name: "As", Addr: "jsonp"}

			//可以读取url中的参数
			context.JSONP(200, &user)
		},
		Method: http.MethodGet,
	})
	return "/json", routeMap
}
