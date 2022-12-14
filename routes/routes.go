package routes

import (
	// "Gin/controller"
	// "response"
	"git.huawei.com/goclient/controller"
	"git.huawei.com/goclient/response"
	"github.com/gin-gonic/gin"
)

func Load(r *gin.Engine) {
	r.POST("/upchain", convert(controller.UpChain))
	r.GET("/querybyphone", convert(controller.QueryByPhone))
	r.GET("/querybyhash", convert(controller.QueryByHash))
	r.POST("/modify", convert(controller.Modify))
	r.NoRoute(controller.NoRoute)
}

func convert(f func(ctx *gin.Context) *response.Response) gin.HandlerFunc {
	return func(c *gin.Context) {
		resp := f(c)
		data := resp.GetData()
		switch item := data.(type) {
		case string:
			c.String(200, item)
		case gin.H:
			c.JSON(200, item)
		}
	}
}
