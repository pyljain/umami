package db

import (
	"context"
	"iter"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// type LogIterFunc func(yield func(Log) bool)

type DB interface {
	CreateAppDatabase(ctx context.Context, name string) (databaseName string, username string, password string, err error) // Create an app database and user
	CreateApp(ctx context.Context, app *App) (string, error)                                                               // Create an app entry in Umami database
	CreateTask(ctx context.Context, appId string, title string, description string) (id string, err error)
	GetApp(ctx context.Context, appId string) (*App, error)
	GetApps(ctx context.Context) ([]*App, error)
	GetTasks(ctx context.Context, appId string) ([]*Task, error)
	GetTask(ctx context.Context, taskId string) (*Task, error)
	UpdateTask(ctx context.Context, appId, taskId string, title, description, status string) error
	InsertLog(ctx context.Context, taskId string, messages []map[string]string) error
	FetchLog(ctx context.Context, taskId string) (*Log, error)
	StartLogStream(ctx context.Context, taskId string) iter.Seq[Log]
}

type App struct {
	Id          bson.ObjectID `bson:"_id" json:"id"`
	Name        string        `bson:"name" json:"name"`
	Description string        `bson:"description" json:"description"`
	User        string        `bson:"user" json:"-"`
	Password    string        `bson:"password" json:"-"`
	Database    string        `bson:"database" json:"-"`
	Created     time.Time     `bson:"created" json:"created"`
	Status      string        `bson:"status" json:"status"`
}

type Task struct {
	Title       string        `json:"title" bson:"title"`
	Description string        `json:"description" bson:"description"`
	AppId       bson.ObjectID `json:"appId" bson:"appId"`
	Id          bson.ObjectID `json:"id" bson:"_id"`
	Status      string        `json:"status" bson:"status"`
	Created     time.Time     `json:"created" bson:"created"`
}

type Log struct {
	Id       bson.ObjectID       `json:"id" bson:"_id"`
	TaskID   bson.ObjectID       `json:"taskId" bson:"taskId"`
	Messages []map[string]string `json:"messages" bson:"messages"`
}

const TaskStatusAuthoring = "authoring"
const TaskStatusInProgress = "in-progress"
const TaskStatusCompleted = "completed"

const AppStatusActive = "active"
