package routes

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"umami/pkg/db"
	"umami/pkg/pubsub"
)

func ManageTasks(dbConn db.DB, pubsubClient pubsub.PubSub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		appId := r.PathValue("id")
		switch r.Method {
		case http.MethodPost:
			t := db.Task{}
			err := json.NewDecoder(r.Body).Decode(&t)
			if err != nil {
				log.Printf("Unable to unmarshal task request %s", err)
			}

			// Create task in database
			id, err := dbConn.CreateTask(r.Context(), appId, t.Title, t.Description)
			if err != nil {
				http.Error(w, fmt.Sprintf("Unable to create task: %s", err), http.StatusInternalServerError)
				return
			}

			// Add Task to queue
			// err = pubsubClient.SendMessage(r.Context(), "tasks", id)
			// if err != nil {
			// 	http.Error(w, fmt.Sprintf("Unable to add task to queue: %s", err), http.StatusInternalServerError)
			// 	return
			// }

			w.WriteHeader(http.StatusCreated)
			err = json.NewEncoder(w).Encode(map[string]string{
				"id": id,
			})
			if err != nil {
				log.Printf("Unable to marshal task response %s", err)
			}
		case http.MethodPatch:
			t := db.Task{}
			taskId := r.PathValue("taskId")
			err := json.NewDecoder(r.Body).Decode(&t)
			if err != nil {
				log.Printf("Unable to unmarshal task request %s", err)
			}

			// Create task in database
			err = dbConn.UpdateTask(r.Context(), appId, taskId, t.Title, t.Description, t.Status)
			if err != nil {
				http.Error(w, fmt.Sprintf("Unable to update task: %s", err), http.StatusInternalServerError)
				return
			}

			if t.Status == db.TaskStatusInProgress {
				// Add Task to queue
				err = pubsubClient.SendMessage(r.Context(), appId, taskId)
				if err != nil {
					http.Error(w, fmt.Sprintf("Unable to add task to queue: %s", err), http.StatusInternalServerError)
					return
				}
			}

			w.WriteHeader(http.StatusOK)

		case http.MethodGet:
			tasks, err := dbConn.GetTasks(r.Context(), appId)
			if err != nil {
				http.Error(w, fmt.Sprintf("Unable to get tasks for the app %s to queue: %s", appId, err), http.StatusInternalServerError)
				return
			}

			if tasks == nil {
				tasks = []*db.Task{}
			}

			w.WriteHeader(http.StatusOK)
			err = json.NewEncoder(w).Encode(tasks)
			if err != nil {
				log.Printf("Unable to marshal tasks response %s", err)
			}

		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}

	}
}
