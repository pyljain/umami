package routes

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"syscall"
	"umami/pkg/db"
	"umami/pkg/pubsub"
)

func StartApp(dbConn db.DB, pubsubClient pubsub.Cache) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Check if app valid
		appId := r.PathValue("id")

		app, err := dbConn.GetApp(r.Context(), appId)
		if err != nil {
			http.Error(w, fmt.Sprintf("Unable to get app: %s", err), http.StatusInternalServerError)
			return
		}

		// Check if redis has app to port mapping
		pid, err := pubsubClient.GetAppPid(r.Context(), appId)
		if err == nil {
			// Kill Process
			err := syscall.Kill(pid, syscall.SIGKILL)
			if err != nil {
				log.Printf("Could not kill running application with pid: %s. Error %s", pid, err)
			}
		}

		// If redis does not then create port mapping by assoicating random high port
		port := rand.Intn(65535-1024) + 1024

		wd, _ := os.Getwd()

		// Create log file
		appLog, err := os.OpenFile(path.Join(wd, "logs", appId+".log"), os.O_RDWR|os.O_CREATE, os.ModePerm)
		if err != nil {
			http.Error(w, fmt.Sprintf("Unable to start app: %s", err), http.StatusInternalServerError)
			return
		}
		defer appLog.Close()

		// Start app in right directory by running ./run.sh <port>
		appDir := filepath.Join("repository", app.Id.Hex())
		cmd := exec.Command("./run.sh", fmt.Sprintf("%d", port))
		cmd.Dir = appDir
		cmd.Stdout = appLog
		cmd.Stderr = appLog

		cmd.SysProcAttr = &syscall.SysProcAttr{
			Setsid: true,
		}

		err = cmd.Start()
		if err != nil {
			http.Error(w, fmt.Sprintf("Unable to start app: %s", err), http.StatusInternalServerError)
			return
		}

		// Add app to redis
		err = pubsubClient.SetAppPid(r.Context(), appId, cmd.Process.Pid)
		if err != nil {
			http.Error(w, fmt.Sprintf("Unable to set app port: %s", err), http.StatusInternalServerError)
			return
		}

		cmd.Process.Release()

		// Proxy to port
		url, err := url.Parse(fmt.Sprintf("http://localhost:%d", port))
		if err != nil {
			http.Error(w, fmt.Sprintf("Unable to start app: %s", err), http.StatusInternalServerError)
			return
		}

		// p := httputil.NewSingleHostReverseProxy(url)
		// nh := http.StripPrefix("/apps/"+appId, p)
		// nh.ServeHTTP(w, r)
		http.Redirect(w, r, url.String(), http.StatusTemporaryRedirect)
	}
}
