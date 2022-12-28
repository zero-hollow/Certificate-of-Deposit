// version
package main

import (
	"git.huawei.com/goclient/controller"
	"git.huawei.com/goclient/routes"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	routes.Load(r)
	r.Run(":8000")
	defer controller.Conn.Close()
	close(controller.ResponseChannel)
}
