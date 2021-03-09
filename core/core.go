// Copyright 2021 Harran Ali. All rights reserved.
// Use of this source code is governed by MIT-style
// license that can be found in the LICENSE file.

package core

import (
	"io"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/harranali/gincoat/core/database"
	"github.com/harranali/gincoat/core/env"
	"github.com/harranali/gincoat/core/middlewaresengine"
	"github.com/harranali/gincoat/core/pkgintegrator"
	"github.com/harranali/gincoat/core/routing"
	"github.com/unrolled/secure"
)

// App struct
type App struct{}

// DB represents Database variable name
const DB = "db"

// New initiates the app struct
func New() *App {
	return &App{}
}

//Bootstrap initiate app
func (app *App) Bootstrap() {
	// set the app mode
	setAppMode()

	//load env vars
	env.Load()

	//initiate package integrator
	pkgintegrator.New()

	//initiate middlewares engine
	middlewaresengine.New()

	//initiate routing engine
	routing.New()

	//initiate db connection
	database.New()

	//TODO support multible dbs
	//register database driver
	pkgintegrator.Resolve().Integrate(Mysql(database.Resolve()))

}

// Run execute the app
func (app *App) Run(portNumber string) {
	//fallack to port number to 80 if not set
	if portNumber == "" {
		portNumber = "80"
	}
	//update log to file
	logsFile, _ := os.Create("logs/app.log")
	gin.DefaultWriter = io.MultiWriter(logsFile, os.Stdout)

	//initiate gin engines
	httpGinEngine := gin.Default()
	httpsGinEngine := gin.Default()

	httpsOn, _ := strconv.ParseBool(env.Get("APP_HTTPS_ON"))
	redirectToHTTPS, _ := strconv.ParseBool(env.Get("APP_REDIRECT_HTTP_TO_HTTPS"))

	if httpsOn {
		//serve the https
		httpsGinEngine = app.integratePackages(httpsGinEngine)
		httpsGinEngine = app.integratePackages(httpsGinEngine)
		router := routing.ResolveRouter()
		httpsGinEngine = app.registerRoutes(httpsGinEngine, router)
		certFile := env.Get("APP_HTTPS_CERT_FILE_PATH")
		keyFile := env.Get("APP_HTTPS_KEY_FILE_PATH")
		host := app.getHTTPSHost() + ":443"
		go httpsGinEngine.RunTLS(host, certFile, keyFile)
	}

	//redirect http to https
	if httpsOn && redirectToHTTPS {
		secureFunc := func() gin.HandlerFunc {
			return func(c *gin.Context) {
				secureMiddleware := secure.New(secure.Options{
					SSLRedirect: true,
					SSLHost:     app.getHTTPSHost() + ":443",
				})
				err := secureMiddleware.Process(c.Writer, c.Request)

				// If there was an error, do not continue.
				if err != nil {
					return
				}

				c.Next()
			}
		}()
		redirectEngine := gin.New()
		redirectEngine.Use(secureFunc)
		redirectEngine.Run(":" + portNumber)
	}

	//serve the http version
	httpGinEngine = app.integratePackages(httpGinEngine)
	httpGinEngine = app.integratePackages(httpGinEngine)
	router := routing.ResolveRouter()
	httpGinEngine = app.registerRoutes(httpGinEngine, router)
	httpGinEngine.Run(":" + portNumber)
}

func (app *App) handleRoute(route routing.Route, ginEngine *gin.Engine) {
	switch route.Method {
	case "get":
		ginEngine.GET(route.Path, route.Handlers...)
	case "post":
		ginEngine.POST(route.Path, route.Handlers...)
	case "delete":
		ginEngine.DELETE(route.Path, route.Handlers...)
	case "patch":
		ginEngine.PATCH(route.Path, route.Handlers...)
	case "put":
		ginEngine.PUT(route.Path, route.Handlers...)
	case "options":
		ginEngine.OPTIONS(route.Path, route.Handlers...)
	case "head":
		ginEngine.HEAD(route.Path, route.Handlers...)
	}
}

func setAppMode() {
	mode := os.Getenv("MODE")
	if mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	} else if mode == "test" {
		gin.SetMode(gin.TestMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}
}

func (app *App) integratePackages(engine *gin.Engine) *gin.Engine {
	for _, pkgIntegration := range pkgintegrator.Resolve().GetIntegrations() {
		engine.Use(pkgIntegration)
	}

	return engine
}

func (app *App) useMiddlewares(engine *gin.Engine) *gin.Engine {
	for _, middleware := range middlewaresengine.Resolve().GetMiddlewares() {
		engine.Use(middleware)
	}

	return engine
}

func (app *App) registerRoutes(engine *gin.Engine, router *routing.Router) *gin.Engine {
	for _, route := range router.GetRoutes() {
		app.handleRoute(route, engine)
	}

	return engine
}

func (app *App) getHTTPSHost() string {
	host := env.Get("APP_HTTPS_HOST")
	//if not set get http instead
	if host == "" {
		host = env.Get("APP_HTTP_HOST")
	}
	//if both not set use local host
	if host == "" {
		host = "localhost"
	}
	return host
}
