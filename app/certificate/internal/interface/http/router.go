package http

import (
	"github.com/gin-gonic/gin"
	"meross_iot/app/certificate/internal/interface/http/controller"
)

func InitRouter(e *gin.Engine)  {
	v1 := e.Group("/v1")
	{
		// 获取证书
		//v1.GET("/device/certificate/:uuid")
		// 生成证书
		v1.PUT("device/certificate/:uuid", controller.Create)
	}
}

