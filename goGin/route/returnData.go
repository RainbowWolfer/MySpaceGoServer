package route

import (
	"GoWeb/domain"
	"GoWeb/goGin/ginTools"
	"github.com/gin-gonic/gin"
	"net/http"
)

//gin get 方法
func getXml() (string, ginTools.RouteMap) {
	routeMap := ginTools.NewRouteMap()

	routeMap.AddRoute(
		ginTools.Route{
			Name: "/getXml",
			Fun: func(context *gin.Context) {
				context.XML(http.StatusOK, gin.H{
					"user": &domain.User{Name: "Ad", Addr: "DDs"},
				})
			},
			Method: http.MethodGet,
		})

	return "/xml", *routeMap
}
