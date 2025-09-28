package worker

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path"
	"umami/pkg/db"
	"umami/pkg/utils"
)

type Work struct {
	Task *db.Task
	App  *db.App
}

func (w *Work) Execute(ctx context.Context, logWriter io.Writer) error {
	// Start a new sub process
	systemInstruction := `The app you generate will be spun up programmatically by the platform that manages these apps. Please ensure that
							you create a run.sh file in the project route with steps that run the web application or the API server. The port will be
							passed in as the first argument. Please remember that users will enhance apps that you build, so create the run.she when it does
							not exist, else update it as necessary.`
	taskBrief := fmt.Sprintf("Important Instructions\n%s\nTask Title: %s\n Task Description:%s", systemInstruction, w.Task.Title, w.Task.Description)
	cmd := exec.CommandContext(ctx, "claude", "-p", "--verbose", "--output-format", "stream-json", "--dangerously-skip-permissions", taskBrief)
	cmd.Env = []string{
		fmt.Sprintf("ANTHROPIC_API_KEY=%s", os.Getenv("ANTHROPIC_API_KEY")),
		fmt.Sprintf("MONGO_CONNECTION_STRING=mongodb://%s:%s@localhost:27017", w.App.User, w.App.Password),
		fmt.Sprintf("MONGO_DB_NAME=%s", w.App.Database),
		fmt.Sprintf("APP_BUCKET_NAME=%s", fmt.Sprintf("umami-bucket-%s", utils.GetName(w.App.Name))),
	}
	cmd.Dir = path.Join(".", "repository", w.App.Id.Hex())

	// cmd.SysProcAttr = &syscall.SysProcAttr{
	// 	Chroot: path.Join(".", "repository", w.App.Id.Hex()),
	// }
	cmd.Stdout = logWriter
	cmd.Stderr = os.Stderr

	log.Printf("Executing task: with claude %s", w.Task.Title)
	err := cmd.Run()
	if err != nil {
		return err
	}
	log.Printf("Completed Claude TASK execution: with claude %s", w.Task.Title)

	return nil
}
