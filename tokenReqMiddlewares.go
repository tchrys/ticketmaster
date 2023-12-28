package main

import (
	redisservice "learning/ticketmaster/redis-service"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/google/uuid"
)

func extractReqBodyFromTokenRequest(c *gin.Context) EventTokenRequest {
	var eventTokenRequest EventTokenRequest
	c.ShouldBindBodyWith(&eventTokenRequest, binding.JSON)
	return eventTokenRequest
}

func tokenReqBodyValidator() gin.HandlerFunc {
	return func(c *gin.Context) {
		var eventTokenRequest EventTokenRequest
		if err := c.ShouldBindBodyWith(&eventTokenRequest, binding.JSON); err != nil {
			c.AbortWithError(http.StatusBadRequest, err)
			return
		}
		c.Next()
	}
}

func processTokenRequest() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenRequest := extractReqBodyFromTokenRequest(c)
		token := uuid.NewString()
		filterName := tokenRequest.EventId + eventTokenBFSuffix
		err := redisservice.GetInstance().AddToBF(filterName, token)
		if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		// log.Default().Println("token " + token + " added to " + filterName)
		c.Status(http.StatusAccepted)
		c.JSON(http.StatusCreated, gin.H{
			"token": token,
		})
	}
}
