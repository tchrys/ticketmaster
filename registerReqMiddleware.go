package main

import (
	redisservice "learning/ticketmaster/redis-service"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

func extractReqBodyFromRegisterRequest(c *gin.Context) *EventRegisterRequest {
	var req EventRegisterRequest
	c.ShouldBindBodyWith(&req, binding.JSON)
	return &req
}

func registerReqBodyValidator() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req EventRegisterRequest
		if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
			c.AbortWithError(http.StatusBadRequest, err)
			return
		}
		c.Next()
	}
}

func tokenValidator() gin.HandlerFunc {
	return func(c *gin.Context) {
		request := extractReqBodyFromRegisterRequest(c)
		filterName := request.EventId + eventTokenBFSuffix
		// log.Default().Println("checking " + filterName + " for token " + request.TokenId)
		filterResult, err := redisservice.GetInstance().ExistsBF(filterName, request.TokenId)
		if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		if filterResult == 0 {
			c.AbortWithError(InvalidTokenHttpCode, NewInvalidTokenError())
			return
		}
		c.Next()
	}
}

func processEventRegisterRequest() gin.HandlerFunc {
	return func(c *gin.Context) {
		request := extractReqBodyFromRegisterRequest(c)
		streamEvent := map[string]interface{}{
			"username":      request.UserId,
			"userTimestamp": time.Now().Unix(),
		}
		streamName := request.EventId + eventRegistrationStreamSuffix
		err := redisservice.GetInstance().AddToStream(streamName, streamEvent)
		// log.Default().Println("Added user " + request.UserId + " to stream " + streamName)
		if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		c.JSON(http.StatusCreated, gin.H{
			"message": "You have been registered to the event ballot",
		})
	}

}
