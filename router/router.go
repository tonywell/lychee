package router

import (
	"lychee/controller"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func Routers() *gin.Engine {
	g := gin.Default()
	g.Use(cors.Default())

	g.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	route := g.Group("/stream")
	{
		route.POST("/push", controller.AcceptRespAndPushToFfmpeg)
		route.POST("/upload/:channel", controller.Mpeg1Video)
		route.GET("/live/:channel", controller.Wsplay)
	}
	return g
}
