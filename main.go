package main

import (
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	r.Use(gin.Logger())

	r.POST("/event/start", eventStarter())
	r.POST("/token", tokenReqBodyValidator(), processTokenRequest())
	r.POST("/event/register", registerReqBodyValidator(), tokenValidator(), processEventRegisterRequest())
	r.POST("/event/result", eventResult())

	r.Run()
}
