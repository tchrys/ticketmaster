package main

import (
	"encoding/json"
	redisservice "learning/ticketmaster/redis-service"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

func eventResult() gin.HandlerFunc {
	return func(c *gin.Context) {
		var request EventResultRequest
		if err := c.ShouldBindBodyWith(&request, binding.JSON); err != nil {
			c.AbortWithError(http.StatusBadRequest, err)
			return
		}
		setName := request.EventId + eventWinnersSetSuffix
		isMember, _ := redisservice.GetInstance().SetIsMember(setName, request.UserId)
		// log.Default().Printf("Checking if %s is member of %s, result %t", request.UserId, setName, isMember)
		c.JSON(http.StatusOK, gin.H{
			"result": isMember,
		})
	}
}

func eventStarter() gin.HandlerFunc {
	return func(c *gin.Context) {
		var request EventStartRequest
		if err := c.ShouldBindBodyWith(&request, binding.JSON); err != nil {
			c.AbortWithError(http.StatusBadRequest, err)
			return
		}
		if err := saveEventToCache(request); err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		if err := redisservice.GetInstance().ReserveBF(request.EventId + eventTokenBFSuffix); err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		ttlSeconds := request.RegisterStopTime - time.Now().Unix()
		time.AfterFunc(
			time.Second*time.Duration(ttlSeconds),
			func() {
				ballotProcessor(request.EventId, request.Capacity)
			},
		)
		c.Status(http.StatusAccepted)
	}
}

func saveEventToCache(event EventStartRequest) error {
	ttlSeconds := event.RegisterStopTime - time.Now().Unix()
	ttlDuration := time.Second * time.Duration(ttlSeconds)
	eventKey := event.EventId + eventMetadataSuffix
	serialisedObj, err := json.Marshal(event)
	if err != nil {
		return err
	}
	if err := redisservice.GetInstance().SetKey(eventKey, serialisedObj, ttlDuration); err != nil {
		return err
	}
	return nil
}

func retrieveEventMetadata(eventId string) (EventStartRequest, error) {
	eventKey := eventId + eventMetadataSuffix
	stringMetadata, err := redisservice.GetInstance().GetKey(eventKey)
	if err != nil {
		return EventStartRequest{}, err
	}
	var eventMetadata EventStartRequest
	err = json.Unmarshal([]byte(stringMetadata), &eventMetadata)
	if err != nil {
		return EventStartRequest{}, err
	}
	return eventMetadata, nil
}
