package db

import (
	"context"
	"fmt"
	"iter"
	"log"
	"time"
	"umami/pkg/utils"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const (
	databaseName        = "umami"
	appsCollection      = "apps"
	tasksCollection     = "tasks"
	logStreamCollection = "logs"
)

type mongoDB struct {
	client              *mongo.Client
	appsCollection      *mongo.Collection
	tasksCollection     *mongo.Collection
	logStreamCollection *mongo.Collection
}

func NewMongoDB(connectionString string) (*mongoDB, error) {
	opts := options.Client().ApplyURI(connectionString).
		SetMaxPoolSize(100).
		SetMaxConnIdleTime(30 * time.Second)
	client, err := mongo.Connect(opts)
	if err != nil {
		return nil, err
	}

	// Create apps collection object
	ac := client.Database(databaseName).Collection(appsCollection)
	tc := client.Database(databaseName).Collection(tasksCollection)
	lc := client.Database(databaseName).Collection(logStreamCollection)

	return &mongoDB{
		client:              client,
		appsCollection:      ac,
		tasksCollection:     tc,
		logStreamCollection: lc,
	}, nil
}

func (m *mongoDB) CreateApp(ctx context.Context, app *App) (string, error) {

	app.Id = bson.NewObjectID()

	res, err := m.appsCollection.InsertOne(ctx, &app)
	if err != nil {
		return "", err
	}

	appId := res.InsertedID.(bson.ObjectID)

	return appId.Hex(), nil
}

func (m *mongoDB) CreateAppDatabase(ctx context.Context, name string) (databaseName string, username string, password string, err error) {
	password = utils.RandStringBytes(12)
	uuid := uuid.New().String()
	username = fmt.Sprintf("admin-%s", uuid)
	databaseName = fmt.Sprintf("%s-%s", utils.GetName(name), uuid)

	res := m.client.Database(databaseName).RunCommand(ctx,
		bson.D{
			{"createUser", username},
			{"pwd", password},
			{"roles", []bson.M{
				{"role": "dbOwner", "db": databaseName},
			}},
		})

	if res.Err() != nil {
		return "", "", "", res.Err()
	}

	return databaseName, username, password, nil
}

func (m *mongoDB) CreateTask(ctx context.Context, appId string, title, description string) (id string, err error) {

	appObjectId, err := bson.ObjectIDFromHex(appId)
	if err != nil {
		return "", err
	}

	t := Task{
		Title:       title,
		Description: description,
		AppId:       appObjectId,
		Id:          bson.NewObjectID(),
		Status:      TaskStatusAuthoring,
		Created:     time.Now(),
	}

	res, err := m.tasksCollection.InsertOne(ctx, &t, nil)
	if err != nil {
		return "", err
	}

	insertedTaskId := res.InsertedID.(bson.ObjectID)

	// Insert into logs
	_, err = m.logStreamCollection.InsertOne(ctx, Log{
		Id:       bson.NewObjectID(),
		TaskID:   insertedTaskId,
		Messages: []map[string]string{},
	})
	if err != nil {
		return "", err
	}

	return insertedTaskId.Hex(), nil
}

func (m *mongoDB) InsertLog(ctx context.Context, taskId string, messages []map[string]string) error {
	taskObjectId, err := bson.ObjectIDFromHex(taskId)
	if err != nil {
		return err
	}

	_, err = m.logStreamCollection.UpdateMany(ctx, bson.M{
		"taskId": taskObjectId,
	}, bson.M{
		"$push": bson.M{
			"messages": bson.M{
				"$each": messages,
			},
		},
	})
	if err != nil {
		return err
	}

	return nil
}

func (m *mongoDB) FetchLog(ctx context.Context, taskId string) (*Log, error) {
	taskObjectId, err := bson.ObjectIDFromHex(taskId)
	if err != nil {
		return nil, err
	}

	l := Log{}
	err = m.logStreamCollection.FindOne(ctx, bson.M{"taskId": taskObjectId}).Decode(&l)
	if err != nil {
		return nil, err
	}

	return &l, nil
}

func (m *mongoDB) UpdateTask(ctx context.Context, appId, taskId string, title, description, status string) error {
	taskObjectId, err := bson.ObjectIDFromHex(taskId)
	if err != nil {
		return err
	}

	appObjectId, err := bson.ObjectIDFromHex(appId)
	if err != nil {
		return err
	}

	_, err = m.tasksCollection.UpdateOne(ctx, bson.M{"_id": taskObjectId, "appId": appObjectId}, bson.M{
		"$set": bson.M{
			"status":      status,
			"title":       title,
			"description": description,
		},
	})
	if err != nil {
		return err
	}

	return nil
}

func (m *mongoDB) GetTask(ctx context.Context, taskId string) (*Task, error) {
	taskObjectId, err := bson.ObjectIDFromHex(taskId)
	if err != nil {
		return nil, err
	}

	var task Task
	err = m.tasksCollection.FindOne(ctx, bson.M{"_id": taskObjectId}).Decode(&task)
	if err != nil {
		return nil, err
	}

	return &task, nil
}

func (m *mongoDB) GetApp(ctx context.Context, appId string) (*App, error) {
	appObjectId, err := bson.ObjectIDFromHex(appId)
	if err != nil {
		return nil, err
	}

	var app App
	err = m.appsCollection.FindOne(ctx, bson.M{"_id": appObjectId}).Decode(&app)
	if err != nil {
		return nil, err
	}

	return &app, nil
}

func (m *mongoDB) GetApps(ctx context.Context) ([]*App, error) {

	var apps []*App
	cursor, err := m.appsCollection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}

	for cursor.Next(ctx) {
		var app App
		err := cursor.Decode(&app)
		if err != nil {
			log.Printf("Unable to decode app %s with error %s", app.Name, err)
			continue
		}
		apps = append(apps, &app)
	}

	return apps, nil
}

func (m *mongoDB) GetTasks(ctx context.Context, appId string) ([]*Task, error) {
	appObjectId, err := bson.ObjectIDFromHex(appId)
	if err != nil {
		return nil, err
	}

	var tasks []*Task
	cursor, err := m.tasksCollection.Find(ctx, bson.M{"appId": appObjectId})
	if err != nil {
		return nil, err
	}

	for cursor.Next(ctx) {
		var task Task
		err := cursor.Decode(&task)
		if err != nil {
			log.Printf("Unable to decode task %s with error %s", task.Description, err)
			continue
		}
		tasks = append(tasks, &task)
	}

	return tasks, nil
}

func (m *mongoDB) StartLogStream(ctx context.Context, taskId string) iter.Seq[Log] {
	return func(yield func(Log) bool) {
		taskObjectId, err := bson.ObjectIDFromHex(taskId)
		if err != nil {
			log.Printf("Unable to decode task %s with error %s", taskId, err)
			return
		}

		csOpts := options.ChangeStream().SetFullDocument(options.UpdateLookup)

		type changeDoc struct {
			FullDocument Log `bson:"fullDocument"`
		}

		stream, err := m.logStreamCollection.Watch(ctx, mongo.Pipeline{
			bson.D{
				{"$match", bson.D{
					{Key: "fullDocument.taskId", Value: taskObjectId},
				}},
			},
		}, csOpts)
		if err != nil {
			log.Printf("Unable to watch task %s with error %s", taskId, err)
			return
		}

		defer stream.Close(ctx)

		for stream.Next(ctx) {
			var ev changeDoc
			if err := stream.Decode(&ev); err != nil {
				log.Printf("Unable to decode log %s with error %s", taskId, err)
				return
			}

			if !yield(ev.FullDocument) {
				return
			}
		}
	}
}
