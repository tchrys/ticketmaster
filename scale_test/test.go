package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/go-retryablehttp"
)

const (
	serverPrefix = "http://localhost:8080/"
	contentType  = "application/json"
)

var stdClient *http.Client

func createEvent(capacity, deadlineDelta int64) string {
	eventId := uuid.NewString()
	timeNow := time.Now().Unix()

	postBody, _ := json.Marshal(map[string]interface{}{
		"eventId":           eventId,
		"capacity":          capacity,
		"registerStartTime": timeNow,
		"registerStopTime":  timeNow + deadlineDelta,
	})
	body := bytes.NewBuffer(postBody)

	resp, err := stdClient.Post(serverPrefix+"event/start", contentType, body)
	if err != nil {
		log.Fatalf("An error occured %v", err)
	}
	defer resp.Body.Close()

	log.Print("Token Request finished with status ", resp.StatusCode)
	return eventId
}

func requestToken(userId, eventId string) string {
	postBody, _ := json.Marshal(map[string]string{
		"userId":  userId,
		"eventId": eventId,
	})
	body := bytes.NewBuffer(postBody)

	resp, err := stdClient.Post(serverPrefix+"token", contentType, body)
	if err != nil {
		log.Fatalf("An error occured %v", err)
	}
	defer resp.Body.Close()

	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	var tokenResponse struct {
		Token string `json:"token"`
	}
	err = json.Unmarshal(responseBody, &tokenResponse)
	if err != nil {
		log.Fatalln(err)
	}
	return tokenResponse.Token
}

func requestEventResult(userId, eventId string) bool {
	postBody, _ := json.Marshal(map[string]string{
		"userId":  userId,
		"eventId": eventId,
	})
	body := bytes.NewBuffer(postBody)
	resp, err := stdClient.Post(serverPrefix+"event/result", contentType, body)
	if err != nil {
		log.Fatalf("An error occured %v", err)
	}
	defer resp.Body.Close()

	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println(string(responseBody))

	var resultResponse struct {
		Result bool `json:"result"`
	}
	err = json.Unmarshal(responseBody, &resultResponse)
	if err != nil {
		log.Fatalln(err)
	}

	return resultResponse.Result
}

func registerRequest(userId, eventId, tokenId string) {
	postBody, _ := json.Marshal(map[string]string{
		"userId":  userId,
		"eventId": eventId,
		"tokenId": tokenId,
	})
	body := bytes.NewBuffer(postBody)

	resp, err := stdClient.Post(serverPrefix+"event/register", contentType, body)
	if err != nil {
		log.Fatalf("An error occured %v", err)
	}
	defer resp.Body.Close()

	log.Print("Register request finished with status ", resp.StatusCode)
}

func setupHttpClient() {
	myclient := retryablehttp.NewClient()
	myclient.RetryMax = 4
	myclient.RetryWaitMin = 500 * time.Millisecond
	myclient.RetryWaitMax = 8 * time.Second
	stdClient = myclient.StandardClient()
}

func main() {
	setupHttpClient()
	capacity := 10000
	deadlineDelta := 120
	demandRatio := 20
	users := capacity * demandRatio

	eventId := createEvent(int64(capacity), int64(deadlineDelta))
	fmt.Println("eventId = ", eventId)

	registeredUsers := make(chan string, users)

	for idx := 0; idx < users; idx++ {
		go func() {
			userId := uuid.NewString()
			fmt.Println("User id = ", userId)
			time.Sleep(time.Millisecond * time.Duration(rand.Intn(1000)*deadlineDelta*9/20))
			token := requestToken(userId, eventId)
			// fmt.Println("token = ", token)
			time.Sleep(time.Millisecond * time.Duration(rand.Intn(1000)*deadlineDelta*9/20))
			registerRequest(userId, eventId, token)
			registeredUsers <- userId
		}()
	}

	time.Sleep(time.Second * (time.Duration(deadlineDelta + 60)))

	winnersCount := 0

	for reqNo := 0; reqNo < users; reqNo++ {
		userId := <-registeredUsers
		registerResult := requestEventResult(userId, eventId)
		fmt.Println("registerResult for user ", userId, " : ", registerResult)
		if registerResult {
			winnersCount += 1
		}
	}

	fmt.Println("winnersCount : ", winnersCount)

}
