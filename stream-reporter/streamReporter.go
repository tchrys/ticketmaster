package streamreporter

import (
	"log"
	"sync"
)

type streamProgress struct {
	toProcess     int
	processed     int
	updateChannel chan int
}

type streamReporter struct {
	ongoingStreams map[string]*streamProgress
}

var (
	instance *streamReporter
	once     sync.Once
)

func GetInstance() *streamReporter {
	once.Do(func() {
		instance = &streamReporter{
			ongoingStreams: make(map[string]*streamProgress),
		}
	})
	return instance
}

func (reporter *streamReporter) AddStream(eventId string, toProcess int) (chan int, chan struct{}) {
	reporter.ongoingStreams[eventId] = &streamProgress{
		toProcess:     toProcess,
		processed:     0,
		updateChannel: make(chan int, toProcess/100+1),
	}

	stream := reporter.ongoingStreams[eventId]
	doneChannel := make(chan struct{})
	go func(eventId string) {
		for {
			batchSize := <-stream.updateChannel
			stream.processed += batchSize
			progressPercentage := 100.0 * float64(stream.processed) / float64(stream.toProcess)
			log.Default().Println("Computing winners progress for", eventId, ":", progressPercentage, "%")
			if stream.processed == stream.toProcess {
				reporter.cleanup(eventId)
				doneChannel <- struct{}{}
				close(doneChannel)
				return
			}
		}
	}(eventId)

	return stream.updateChannel, doneChannel
}

func (reporter *streamReporter) cleanup(eventId string) {
	close(reporter.ongoingStreams[eventId].updateChannel)
	delete(reporter.ongoingStreams, eventId)
}
