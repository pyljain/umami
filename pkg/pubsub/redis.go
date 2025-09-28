package pubsub

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

type redisClient struct {
	client *redis.Client
}

func NewRedis(address string) (*redisClient, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     address,
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	_, err := rdb.Ping(context.Background()).Result()
	if err != nil {
		return nil, err
	}

	ctx := context.Background()

	if err := rdb.ConfigSet(ctx, "notify-keyspace-events", "Ex").Err(); err != nil {
		log.Printf("WARN: couldn't enable keyevent notifications automatically: %v", err)
		log.Printf("      Make sure redis.conf has: notify-keyspace-events Ex")
	}

	channel := fmt.Sprintf("__keyevent@%d__:expired", 0)
	pubsub := rdb.PSubscribe(ctx, channel)
	// defer pubsub.Close()

	go func() {
		for msg := range pubsub.Channel() {
			key := msg.Payload // the expired key name
			if !strings.HasPrefix(key, "lock:") {
				continue
			}
			appID := strings.TrimPrefix(key, "lock:")
			if appID == "" {
				continue
			}

			// Put the task that was being processed back into the app queue as the first task to be processed
			taskId := rdb.Get(ctx, "processing:"+appID)
			if taskId.Err() == nil {
				log.Printf("Restoring the task that failed processing %s, %s", appID, taskId.Val())
				err = rdb.RPush(ctx, "q:"+appID, taskId.Val()).Err()
				if err != nil {
					log.Printf("Unable to restore the task that failed processing %s, %v, %s", appID, taskId.Val(), err)
				}
			}

			if err := rdb.ZAdd(ctx, "ready", redis.Z{
				Score:  1,
				Member: appID,
			}).Err(); err != nil {
				log.Printf("LPUSH %s %s failed: %v", "ready", appID, err)
				continue
			}
			log.Printf("Enqueued appID=%q to %q after %q expired", appID, "ready", key)
		}
		log.Printf("PubSub channel closed")
	}()

	return &redisClient{
		client: rdb,
	}, nil
}

func (r *redisClient) SendMessage(ctx context.Context, appID string, taskID string) error {
	appQueueName := fmt.Sprintf("q:%s", appID)

	log.Printf("SEND MESSAGE: Task Id being inserted is %s", taskID)
	res := r.client.LPush(ctx, appQueueName, taskID)
	if res.Err() != nil {
		return res.Err()
	}

	// Check lock for app
	if r.client.Get(ctx, "lock:"+appID).Err() != nil {
		// Ready is a set to ensure that the same appID is only added once to the data structure
		res := r.client.ZAdd(ctx, "ready", redis.Z{
			Score:  1,
			Member: appID,
		})
		if res.Err() != nil {
			return res.Err()
		}
	}
	return nil
}

// Called by workers when they need to BRPOP an appID and hence a task to process
func (r *redisClient) PullMessage(ctx context.Context) (string, error) {
	for {
		// Pop message from ready queue
		res := r.client.BZPopMin(ctx, 0, "ready")
		if res.Err() != nil {
			log.Printf("Redis.PullMessage Unable to pull message from redis %s", res.Err())
			time.Sleep(time.Second * 10)
			continue
		}

		appID := res.Val().Member.(string)

		// Check lock for app
		log.Printf("Redis.PullMessage Worker trying to lock %s", appID)
		log.Printf("LOCKING NOW: TASK ID IS %s", appID)
		if r.client.SetNX(ctx, "lock:"+appID, "", 30*time.Second).Err() != nil {
			log.Printf("Redis.PullMessage Worker unable to lock %s", appID)
			// App is locked
			continue
		}
		log.Printf("Redis.PullMessage Worker lock successful %s", appID)

		// Pop message from app queue
		log.Printf("Redis.PullMessage Getting message from app task queue %s", appID)
		appQueueName := fmt.Sprintf("q:%s", appID)
		task := r.client.RPop(ctx, appQueueName)
		if res.Err() != nil {
			log.Printf("Redis.PullMessage Unable to pull message from app task queue %s", res.Err())
			r.client.Del(ctx, "lock:"+appID)
			continue
		}

		taskId := task.Val()

		if taskId == "" {
			log.Printf("Redis.PullMessage Worker got no message %s", taskId)
			r.client.Del(ctx, "lock:"+appID)
			continue
		}

		log.Printf("Redis.PullMessage Worker got message %s", taskId)
		err := r.client.Set(ctx, "processing:"+appID, taskId, 0).Err()
		if err != nil {
			// TODO: if unable to set then remove lock
			r.client.Del(ctx, "lock:"+appID)
			continue
		}

		return taskId, nil
	}

}

func (r *redisClient) RenewLock(ctx context.Context, appID string) error {
	log.Printf("Trying to renew lock %s", appID)
	return r.client.Set(ctx, "lock:"+appID, "", 30*time.Second).Err()
}

func (r *redisClient) DeleteLock(ctx context.Context, appID string) error {

	// Remove the processing key once done working on the task
	err := r.client.Del(ctx, "processing:"+appID).Err()
	if err != nil {
		log.Printf("Error deleting processing key %s", appID)
	}

	log.Printf("Trying to delete lock %s", appID)
	err = r.client.Del(ctx, "lock:"+appID).Err()
	if err != nil {
		log.Printf("Error deleting lock %s", appID)
	}
	log.Printf("Successfully deleted lock %s", appID)

	log.Printf("Lock status for %s is %s", appID, r.client.Get(ctx, "lock:"+appID).Val())

	// Check if app queue depth is greater than 0
	l := r.client.LLen(ctx, "q:"+appID)
	if l.Err() != nil {
		return err
	}

	// Set app to ready
	if l.Val() > 0 {
		err = r.client.ZAdd(ctx, "ready", redis.Z{
			Score:  1,
			Member: appID,
		}).Err()
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *redisClient) GetAppPid(ctx context.Context, appID string) (int, error) {
	return r.client.Get(ctx, "pid:"+appID).Int()
}

func (r *redisClient) SetAppPid(ctx context.Context, appID string, pid int) error {
	return r.client.Set(ctx, "pid:"+appID, pid, 0).Err()
}
