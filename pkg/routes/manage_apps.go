package routes

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path"
	"time"
	"umami/pkg/db"
	"umami/pkg/storage"

	"github.com/go-git/go-git/v6"
)

func ManageApps(dbConn db.DB, storageClient storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		if r.Method == http.MethodPost {
			app := db.App{}
			err := json.NewDecoder(r.Body).Decode(&app)
			if err != nil {
				http.Error(w, fmt.Sprintf("Bad Request %s", err), http.StatusBadRequest)
				return
			}

			// 2. Create database in mongo
			// 3. Create a user with access to that database only
			databaseName, username, password, err := dbConn.CreateAppDatabase(r.Context(), app.Name)
			if err != nil {
				http.Error(w, fmt.Sprintf("Unable to create a new database for the app: %s", err), http.StatusInternalServerError)
				return
			}

			app.Created = time.Now()
			app.Status = db.AppStatusActive
			app.User = username
			app.Password = password
			app.Database = databaseName

			// 1. Creates app in mongo
			appId, err := dbConn.CreateApp(r.Context(), &app)
			if err != nil {
				http.Error(w, fmt.Sprintf("Unable to create app: %s", err), http.StatusInternalServerError)
				return
			}

			// 4. Creates git repository
			repoDir := path.Join(".", "repository", appId)
			err = os.MkdirAll(repoDir, os.ModePerm)
			if err != nil {
				http.Error(w, fmt.Sprintf("Unable to create repository: %s", err), http.StatusInternalServerError)
				return
			}

			_, err = git.PlainInit(repoDir, false)
			if err != nil {
				http.Error(w, fmt.Sprintf("Unable to initialise repository: %s", err), http.StatusInternalServerError)
				return
			}

			// 5. Creates bucket in GCS
			err = storageClient.CreateBucket(r.Context(), app.Name)
			if err != nil {
				http.Error(w, fmt.Sprintf("Unable to create bucket: %s", err), http.StatusInternalServerError)
				return
			}

			err = json.NewEncoder(w).Encode(map[string]string{
				"id": appId,
			})
			if err != nil {
				http.Error(w, "Unable to respond", http.StatusInternalServerError)
				return
			}

		} else if r.Method == http.MethodGet {
			apps, err := dbConn.GetApps(r.Context())
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			err = json.NewEncoder(w).Encode(&apps)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

		} else {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
	}

}
