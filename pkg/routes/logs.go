package routes

import (
	"encoding/json"
	"net/http"
	"umami/pkg/db"
)

func FetchLogs(database db.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		taskId := r.PathValue("taskId")
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		logEntries, err := database.FetchLog(r.Context(), taskId)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		err = json.NewEncoder(w).Encode(logEntries)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
}
