package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"umami/pkg/db"
	"umami/pkg/pubsub"
	"umami/pkg/routes"
	"umami/pkg/storage"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{}

func main() {
	// Start Server that receives execute task signal
	// POST /api/v1/apps

	// POST /api/v1/apps/<id>/task
	// {
	//		"description": "I want to X"
	// }
	// Add item to redis
	// 1. Worker pick up task from redis
	// 1.1 Workers clones repo
	// 2. Worker starts claude code with task
	// 3. Worker execute and monitors task
	// 4. Checkpoint in git
	// 5. Push to repo

	ctx := context.Background()

	router := http.NewServeMux()
	mongoDb, err := db.NewMongoDB("mongodb://localhost:27017")
	if err != nil {
		log.Fatalf("Unable to connect to database %s", err)
	}

	storageClient, err := storage.NewGCS(ctx)
	if err != nil {
		log.Fatalf("Unable to connect to storage %s", err)
	}

	pubsubClient, err := pubsub.NewRedis("localhost:6379")
	if err != nil {
		log.Fatalf("Unable to connect to pubsub %s", err)
	}

	router.HandleFunc("/api/v1/apps", routes.ManageApps(mongoDb, storageClient))
	router.HandleFunc("/api/v1/apps/{id}/tasks", routes.ManageTasks(mongoDb, pubsubClient))
	router.HandleFunc("/api/v1/apps/{id}/download", routes.Download(mongoDb))
	router.HandleFunc("/api/v1/apps/{id}/tasks/{taskId}", routes.ManageTasks(mongoDb, pubsubClient))
	router.HandleFunc("/api/v1/apps/{id}/tasks/{taskId}/logs", routes.FetchLogs(mongoDb))
	router.HandleFunc("/api/v1/apps/{id}/tasks/{taskId}/logs/ws", func(w http.ResponseWriter, r *http.Request) {

		taskID := r.PathValue("taskId")

		// Update to Web Sockets
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Print("upgrade:", err)
			return
		}
		defer c.Close()

		// Start Stream with database
		for logSnapshot := range mongoDb.StartLogStream(ctx, taskID) {
			logBytes, err := json.Marshal(logSnapshot)
			if err != nil {
				log.Println(err)
				break
			}
			err = c.WriteMessage(websocket.TextMessage, logBytes)
			if err != nil {
				log.Println("write:", err)
				break
			}
		}
	})
	router.HandleFunc("/apps/{id}", routes.StartApp(mongoDb, pubsubClient))

	err = http.ListenAndServe(":9808", router)
	if err != nil {
		log.Fatalln(err)
	}
}
