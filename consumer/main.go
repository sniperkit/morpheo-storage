package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/DeepSee/dc-compute"
	common "github.com/DeepSee/dc-compute/common"
)

type Worker struct {
	executionBackend common.ExecutionBackend
	storageBackend   common.StorageBackend
}

func (w *Worker) HandleLearn(message []byte) (err error) {
	var task dccompute.LearnTask
	err = json.NewDecoder(bytes.NewReader(message)).Decode(&task)
	if err != nil {
		return common.NewHandlerFatalError(fmt.Errorf("Error un-marshaling train task: %s -- Body: %s", err, message))
	}

	if err = task.Check(); err != nil {
		return common.NewHandlerFatalError(fmt.Errorf("Error in train task: %s -- Body: %s", err, message))
	}

	// Let's pass the learn task to our execution backend
	score, err := w.executionBackend.Train(task.LearnUplet.Model, task.Data)
	if err != nil {
		return common.NewHandlerFatalError(fmt.Errorf("Error in train task: %s -- Body: %s", err, message))
	}

	// TODO: update the score (notify the orchestrator ?)
	log.Printf("Train finished with success. Score %f", score)

	return
}

func (w *Worker) HandleTest(message []byte) (err error) {
	var task dccompute.TestTask
	err = json.NewDecoder(bytes.NewReader(message)).Decode(&task)
	if err != nil {
		return fmt.Errorf("Error un-marshaling test task: %s -- Body: %s", err, message)
	}

	// Let's pass the test task to our execution backend
	score, err := w.executionBackend.Test(task.LearnUplet.Model, task.Data)
	if err != nil {
		return common.NewHandlerFatalError(fmt.Errorf("Error in test task: %s -- Body: %s", err, message))
	}

	// TODO: update the score (notify the orchestrator ?)
	log.Printf("Test finished with success. Score %f", score)

	return
}

func (w *Worker) HandlePred(message []byte) (err error) {
	var task dccompute.Preduplet
	err = json.NewDecoder(bytes.NewReader(message)).Decode(&task)
	if err != nil {
		return fmt.Errorf("Error un-marshaling pred-uplet: %s -- Body: %s", err, message)
	}

	// Let's pass the prediction task to our execution backend
	prediction, err := w.executionBackend.Predict(task.Model, task.Data)
	if err != nil {
		return common.NewHandlerFatalError(fmt.Errorf("Error in prediction task: %s -- Body: %s", err, message))
	}

	// TODO: send the prediction to the viewer, asynchronously
	log.Printf("Predicition completed with success. Predicition %f", prediction)

	return
}

func main() {
	// TODO: improve config and add a -container-backend flag and relevant opts
	// TODO: add NSQ consumer flags
	var (
		lookupUrls           dccompute.MultiStringFlag
		topic                string
		channel              string
		queuePollingInterval time.Duration
	)

	flag.Var(&lookupUrls, "lookup-urls", "The URLs of the Nsqlookupd instances to fetch our topics from.")
	flag.StringVar(&topic, "topic", "learn", "The topic of the Nsqd/Nsqlookupd instance to listen to.")
	flag.StringVar(&channel, "channel", "compute", "The channel to use (default: compute)")
	flag.DurationVar(&queuePollingInterval, "lookup-interval", time.Second, "The interval at which nsqlookupd will be polled")
	flag.Parse()

	// Config check
	if len(lookupUrls) == 0 {
		lookupUrls = append(lookupUrls, "nsqlookupd:6460")
	}

	if topic != dccompute.LearnTopic && topic != dccompute.PredictionTopic && topic != dccompute.TestTopic {
		log.Panicf("Unknown topic: %s, valid values are %s, %s and %s", topic, dccompute.LearnTopic, dccompute.TestTopic, dccompute.PredictionTopic)
	}

	// Let's connect with Storage (TODO: replace our mock with the real storage)
	storageBackend := common.NewStorageAPIMock()

	// Let's hook to our container backend and create a Worker instance containing
	// our message handlers TODO: put data folders in flags
	executionBackend, err := common.NewDockerBackend("/data")
	if err != nil {
		log.Panicf("Impossible to connect to Docker container backend: %s", err)
	}
	worker := Worker{
		executionBackend: executionBackend,
		storageBackend:   storageBackend,
	}

	// Let's hook with our consumer
	consumer := common.NewNSQConsumer(lookupUrls, channel, queuePollingInterval)

	// Wire our message handlers
	consumer.AddHandler(dccompute.LearnTopic, worker.HandleLearn, 1)
	// consumer.AddHandler(dccompute.TestTopic, worker.HandleTest, 1)
	// consumer.AddHandler(dccompute.PredictionTopic, worker.HandlePred, 1)

	// Let's connect to the for real and start pulling tasks
	consumer.ConsumeUntilKilled()

	log.Println("Consumer has been gracefully stopped... Bye bye!")
	return
}