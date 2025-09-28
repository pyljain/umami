package main

import (
	"context"
	"log"
	"time"
	"umami/pkg/claude"
	"umami/pkg/db"
	"umami/pkg/pubsub"
	"umami/pkg/worker"
)

const (
	maxNumberOfSubProcesses = 3
)

func main() {
	workChan := make(chan *worker.Work, maxNumberOfSubProcesses)
	ctx := context.Background()
	// Initialise Redis connection
	// redisAddress := os.Getenv("REDIS_ADDRESS")
	redisClient, err := pubsub.NewRedis("localhost:6379")
	if err != nil {
		log.Fatalf("Unable to connect to redis %s", err)
	}

	// Initialise Mongo connection
	// mongoAddress := os.Getenv("MONGO_ADDRESS")
	mongoClient, err := db.NewMongoDB("mongodb://localhost:27017")
	if err != nil {
		log.Fatalf("Unable to connect to mongo %s", err)
	}

	// Infinite for to pull messages from Redis
	for {
		// Pull message from Redis
		log.Printf("Worker waiting for message...")
		taskId, err := redisClient.PullMessage(ctx)
		if err != nil {
			log.Printf("Unable to pull message from redis %s", err)
			continue
		}
		log.Printf("Worker got message %s", taskId)

		taskCtx, cancel := context.WithCancel(ctx)

		// Fetch task details from Mongo
		task, err := mongoClient.GetTask(ctx, taskId)
		if err != nil {
			log.Printf("Unable to pull task from the datastore %s", err)
			cancel()
			continue
		}

		app, err := mongoClient.GetApp(ctx, task.AppId.Hex())
		if err != nil {
			log.Printf("Unable to pull app from the datastore %s", err)
			cancel()
			continue
		}

		log.Printf("Executing task %+v for app %+v", task, app)
		taskInProgress := true

		go func() {
			for {
				select {
				case <-taskCtx.Done():
					log.Printf("Renew loop cancelled for Task %s for app %s cancelled", task.Id, task.AppId)
					return
				case <-time.After(time.Second * 15):
					if taskInProgress {
						redisClient.RenewLock(taskCtx, task.AppId.Hex())
					}
				}
			}
		}()

		go func() {
			w := <-workChan

			taskLogWriter := claude.NewLogWriter(mongoClient, task.Id.Hex())

			// Create a sub process
			err := w.Execute(taskCtx, taskLogWriter)
			if err != nil {
				log.Printf("Unable to complete work for task %s and app %s. Error: %s", w.Task.Id, w.Task.AppId, err)
			}
			// log.Printf("Processing task %s for app %s", w.Task.Id, w.Task.AppId)
			// // time.Sleep(time.Second * 30)
			// log.Printf("Completed task %s for app %s", w.Task.Id, w.Task.AppId)

			taskInProgress = false
			redisClient.DeleteLock(taskCtx, task.AppId.Hex())
			cancel()

			// Update task status
			err = mongoClient.UpdateTask(ctx, w.Task.AppId.Hex(), w.Task.Id.Hex(), w.Task.Title, w.Task.Description, "completed")
			if err != nil {
				log.Printf("Unable to update task status for task %s and app %s. Error: %s", w.Task.Id, w.Task.AppId, err)
			}

			// Kill subprocess
			// Clear port in redis
		}()

		// Spin up a sub process to run the task
		workChan <- &worker.Work{
			Task: task,
			App:  app,
		}
	}
}
