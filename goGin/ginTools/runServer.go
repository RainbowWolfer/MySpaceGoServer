package ginTools

import (
	"github.com/gin-gonic/gin"
	"html/template"
	"net/http"
	"strconv"
	"strings"
)

type EngineInfo struct {
	Engine    *gin.Engine
	Port      int
	GroupPath map[string]RouteMap
}

func NewDefaultEngine(port int) *EngineInfo {
	return &EngineInfo{
		Engine:    gin.Default(),
		Port:      port,
		GroupPath: make(map[string]RouteMap),
	}
}

func (receiver *EngineInfo) RunServer() {

	for groupPath := range receiver.GroupPath {
		group := receiver.Engine.Group(groupPath)
		receiver.addGroup(group, receiver.GroupPath[groupPath])
	}

	var build strings.Builder
	build.WriteString(":")
	build.WriteString(strconv.Itoa(receiver.Port))

	err := receiver.Engine.Run(build.String())
	if err != nil {
		panic("失败了 ！！")
	}
}
func (receiver *EngineInfo) SetFuncMap(funcMap template.FuncMap) *EngineInfo {
	receiver.Engine.SetFuncMap(funcMap)
	return receiver
}

// Static 配置静态web目录（配置文件）
func (receiver *EngineInfo) Static(relativePath, root string) *EngineInfo {
	receiver.Engine.Static(relativePath, root)
	return receiver

}

// LoadHTMLGlob 加载模板 多目录操作-
func (receiver *EngineInfo) LoadHTMLGlob(pattern string) *EngineInfo {
	receiver.Engine.LoadHTMLGlob(pattern)
	return receiver

}

func (receiver *EngineInfo) AddRouteMap(routes *RouteMap) *EngineInfo {
	for _, route := range routes.route1Map {
		switch route.Method {
		case http.MethodGet:
			receiver.Engine.GET(route.Name, route.Fun)
		case http.MethodPut:
			receiver.Engine.PUT(route.Name, route.Fun)
		case http.MethodPost:
			receiver.Engine.POST(route.Name, route.Fun)
		case http.MethodDelete:
			receiver.Engine.DELETE(route.Name, route.Fun)
		default:
			receiver.Engine.GET(route.Name, route.Fun)
		}
	}
	return receiver

}

func (receiver *EngineInfo) AddGroupRouteMap(relativePath string, routeMap RouteMap) *EngineInfo {
	receiver.GroupPath[relativePath] = routeMap
	return receiver
}
func (receiver *EngineInfo) addGroup(group *gin.RouterGroup, routes RouteMap) *EngineInfo {
	for _, route := range routes.route1Map {
		switch route.Method {
		case http.MethodGet:
			group.GET(route.Name, route.Fun)
		case http.MethodPut:
			group.PUT(route.Name, route.Fun)
		case http.MethodPost:
			group.POST(route.Name, route.Fun)
		default:
			group.GET(route.Name, route.Fun)
		}
	}
	return receiver
}
