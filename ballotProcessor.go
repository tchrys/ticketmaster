package main

import (
	"errors"
	redisservice "learning/ticketmaster/redis-service"
	streamreporter "learning/ticketmaster/stream-reporter"
	"log"
	"strconv"

	"github.com/go-redis/redis"
)

type WorkerMetadata struct {
	batchSize     int
	eventId       string
	streamName    string
	updateChannel chan int
	ballotSetName string
	capacity      int
}

func ballotProcessor(eventId string, capacity int64) {
	streamName := eventId + eventRegistrationStreamSuffix
	toProcess, err := redisservice.GetInstance().StreamLength(streamName)
	if err != nil {
		panic(err)
	}
	streamUpdateChannel, doneChannel := streamreporter.GetInstance().AddStream(eventId, int(toProcess))
	metadata := WorkerMetadata{
		batchSize:     100,
		ballotSetName: eventId + eventBallotSetSuffix,
		eventId:       eventId,
		streamName:    streamName,
		updateChannel: streamUpdateChannel,
		capacity:      int(capacity),
	}

	go ballotWorker(&metadata, "-")
	<-doneChannel
	go recordWinners(eventId, int64(capacity))
}

func ballotWorker(metadata *WorkerMetadata, start string) {
	events, err := redisservice.GetInstance().ReadFromStream(metadata.streamName, start, int64(metadata.batchSize))
	if err != nil {
		panic(err)
	}
	isLastBatch := len(events) < metadata.batchSize
	if !isLastBatch {
		lastId := events[len(events)-1].ID
		go ballotWorker(metadata, "("+lastId)
	}
	consumeStreamEvents(events, metadata.ballotSetName, isLastBatch)
	metadata.updateChannel <- len(events)
}

func consumeStreamEvents(events []redis.XMessage, ballotSetName string, isLastBatch bool) {
	for _, xmessage := range events {
		registrationEvent, err := convertRedisEvent(xmessage)
		if err != nil {
			log.Default().Println(err.Error())
		}
		// log.Default().Println("Consumed event for user " + registrationEvent.username)
		err = redisservice.GetInstance().AddToSortSet(
			ballotSetName,
			float64(registrationEvent.userTimestamp),
			registrationEvent.username,
		)
		// log.Default().Println("Event for username " + registrationEvent.username + " added to set " + ballotSetName)
		if err != nil {
			log.Default().Println(err.Error())
		}
	}
}

func recordWinners(eventId string, capacity int64) {
	// log.Default().Println("Extracting winners for event " + eventId)
	ballotSetName := eventId + eventBallotSetSuffix
	batchSize := 1000
	winners, err := redisservice.GetInstance().SortSetRange(ballotSetName, 0, capacity-1)
	if err != nil {
		log.Default().Println(err.Error())
	}
	totalSize := len(winners)
	batches := totalSize/batchSize + 1
	// log.Default().Printf("Size: %d, Batches: %d\n", totalSize, batches)
	for batch := 0; batch < batches; batch++ {
		go processBatch(batchSize, batch, totalSize, eventId+eventWinnersSetSuffix, winners)
	}
}

func processBatch(batchSize, batch, totalSize int, ballotSetName string, winners []string) {
	startIdx := batch * batchSize
	endIdx := (batch + 1) * batchSize
	if endIdx >= totalSize {
		endIdx = totalSize
	}
	for idx := startIdx; idx < endIdx; idx++ {
		// log.Default().Printf("Added %s to winners set %s", winners[idx], ballotSetName)
		redisservice.GetInstance().AddToSet(ballotSetName, winners[idx])
	}
}

func convertRedisEvent(event redis.XMessage) (RegistrationEvent, error) {
	username, usernameFound := event.Values["username"]
	timestamp, timestampFound := event.Values["userTimestamp"]
	if !usernameFound {
		return RegistrationEvent{}, errors.New("username not found")
	}
	if !timestampFound {
		return RegistrationEvent{}, errors.New("timestamp not found")
	}
	stringUsername, ok := username.(string)
	if !ok {
		return RegistrationEvent{}, errors.New("expected string for username field")
	}
	intTimestamp, err := strconv.ParseInt(timestamp.(string), 10, 64)
	if err != nil {
		return RegistrationEvent{}, errors.New("expected int for timestamp field")
	}
	return RegistrationEvent{stringUsername, intTimestamp}, nil
}
