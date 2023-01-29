package main

import (
	"net/http"
	"path/filepath"

	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()

	router.LoadHTMLGlob("static/**.html")

	router.StaticFile("/robots.txt", "./static/robots/robots.txt")
	router.StaticFile("/humans.txt", "./static/humans/humans.txt")

	router.Use(static.Serve("/", static.LocalFile("static", false)))

	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"title": "Authentik-Manager",
		})
	})

	SetEndpoints(router)
	router.Run(":8080") // listen and serve on 0.0.0.0:8080 by default
}

func addStaticGlob(glob string, router *gin.Engine) {
	files, err := filepath.Glob(glob)
	if err != nil {
		panic(err.Error)
	}
	for _, file := range files {
		router.StaticFile("/"+filepath.Base(file), file)
	}

}
