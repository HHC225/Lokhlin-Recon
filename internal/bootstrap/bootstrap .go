package bootstrap

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"log"
	"recon/internal/config"
	"recon/internal/constants"
	"recon/internal/database/mongodb"
	"time"
)

type StartupEvent func(ctx context.Context) error

var startupEvents []StartupEvent

func RegisterStartupEvent(event StartupEvent) {
	startupEvents = append(startupEvents, event)
}

func ExecuteStartupEvents() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	for _, event := range startupEvents {
		if err := event(ctx); err != nil {
			return err
		}
	}
	return nil
}

func init() {

	//RegisterStartupEvent(func(ctx context.Context) error {
	//	return config.Load()
	//})

	//RegisterStartupEvent(func(ctx context.Context) error {
	//	return logger.Init()
	//})

	//RegisterStartupEvent(func(ctx context.Context) error {
	//	return mongodb.Init()
	//})

	//RegisterStartupEvent(func(ctx context.Context) error {
	//	return redis.Init()
	//})

	RegisterStartupEvent(func(ctx context.Context) error {
		return Banner()
	})

	RegisterStartupEvent(func(ctx context.Context) error {
		return mongodb.CreateDatabase()
	})

	RegisterStartupEvent(func(ctx context.Context) error {
		if err := mongodb.FindAll("notification", bson.M{"state": true}, bson.M{"_id": 0, "method": 1, "url": 1, "contentType": 1, "data": 1, "state": 1}, &config.NotificationApi); err != nil {
			return err
		}
		return nil
	})

	err := ExecuteStartupEvents()
	if err != nil {
		log.Fatalf("Error during startup events: %v", err)
		return
	}
}

func Banner() error {
	fmt.Println(constants.DisplayName() + " collection service")
	return nil
}
