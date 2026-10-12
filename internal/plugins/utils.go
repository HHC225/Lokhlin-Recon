// plugins-------------------------------------
// @file      : utils.go
// @time      : 2025/12/20 20:15
// -------------------------------------------

package plugins

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	mongoOptions "go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
	"recon/internal/database/mongodb"
	"recon/internal/logger"
	plgOption "recon/internal/options"
	"recon/internal/utils"
)

func SetStringVariable(key, value string) {
	ctx := context.Background()
	collection := mongodb.DB.Collection("variable")

	filter := bson.M{"key": key}
	update := bson.M{"$set": bson.M{"key": key, "value": value}}

	opts := options.Update().SetUpsert(true)
	_, err := collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		logger.Error(fmt.Sprintf("Error updating variable %s: %v", key, err))
	}
}

func GetStringVariable(key string) string {
	ctx := context.Background()
	collection := mongodb.DB.Collection("variable")

	var result bson.M
	err := collection.FindOne(ctx, bson.M{"key": key}).Decode(&result)
	if err != nil {
		logger.Error(fmt.Sprintf("Error getting variable %s: %v", key, err))
		return ""
	}

	if value, ok := result["value"].(string); ok {
		return value
	}
	return ""
}

func RunPluginOnce(hash string) {

	var result PluginInfo
	ctx := context.Background()
	collection := mongodb.DB.Collection("plugins")
	opts := mongoOptions.FindOne().
		SetProjection(bson.M{
			"status": 1,
		})
	err := collection.FindOne(ctx, bson.M{"hash": hash}, opts).Decode(&result)
	if err != nil {
		logger.Error(err.Error())
		return
	}
	if !result.Status {
		return
	}
	plg, flag := GlobalPluginManager.GetPlugin(hash)
	if flag {
		go func() {

			flag := utils.HasContext(hash)
			if flag {
				utils.CancelContext(hash)
				time.Sleep(time.Second * 5)
			}
			if plg.Cycle() == "1" {
				err := plg.Execute(plgOption.NewOption(hash))
				if err != nil {
					logger.Error("failed to execute plugin", zap.Error(err))
					return
				}
			}
		}()
	} else {
		logger.Error(fmt.Sprintf("%v plugin not found", hash))
	}
}
