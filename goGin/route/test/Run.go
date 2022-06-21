package main

import (
	"rainbowwolfer/myspacegoserver/goGin/ginTools"
	"rainbowwolfer/myspacegoserver/goGin/route"
	"rainbowwolfer/myspacegoserver/goGin/route/user"
)

func main() {
	engine := ginTools.NewDefaultEngine(8080)
	engine.SetFuncMap(route.AllMethodMap())
	//tempPath := "goGin/templates/**/**/*"
	//engine.LoadHTMLGlob(tempPath)
	engine.Engine.LoadHTMLFiles(ginTools.WalkFiles("goGin/templates/")...)
	path, routeMap := route.ReturnJsonMap()

	staticPath := "goGin/static"
	engine.Static("/static", staticPath)

	engine.AddGroupRouteMap(path, *routeMap)
	engine.AddGroupRouteMap(user.UserHandler()).RunServer()
}
