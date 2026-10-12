// worker-------------------------------------
// @file      : screenshot.go
// @time      : 2025/11/2 17:58
// -------------------------------------------

package worker

import (
	"context"
	"encoding/base64"
	"fmt"
	"path"
	"recon/internal/config"
	"recon/internal/logger"
	"recon/internal/utils"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
	"recon/internal/repositories/common"
)

func ScreenshotHandle(ctx context.Context) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Worker error recovered:", r)

			select {
			case <-ctx.Done():
				fmt.Println("Task cancelled; ScreenshotHandle will not restart")
			default:
				go ScreenshotHandle(ctx)
			}
		}
	}()

	repo := common.NewRepository()

	for {
		select {
		case <-ctx.Done():
			fmt.Println("Task cancelled; ScreenshotHandle stopped")
			return
		default:

			limit := int64(100)
			opts := options.Find().SetLimit(limit)

			results, err := repo.FindMany(ctx, "screenshot", bson.M{}, opts)
			if err != nil {
				logger.Error(fmt.Sprintf("Failed to query screenshot data: %v\n", err))
				time.Sleep(10 * time.Second)
				continue
			}

			if len(results) == 0 {
				time.Sleep(60 * time.Second)
				continue
			}

			var ids []primitive.ObjectID

			for _, item := range results {

				hash, _ := item["hash"].(string)
				content, _ := item["content"].(string)

				if id, ok := item["_id"].(primitive.ObjectID); ok {
					ids = append(ids, id)
				}
				filepath := path.Join(config.GlobalConfig.System.ScreenshotDir, fmt.Sprintf("%v.png", hash))
				decoded, err := base64.StdEncoding.DecodeString(content)
				if err != nil {
					logger.Error(fmt.Sprintf("Failed to decode screenshot base64: %v %v\n", err, content))
					continue
				}
				err = utils.WriteFile(filepath, decoded)
				if err != nil {
					logger.Error(fmt.Sprintf("Failed to write screenshot: %v\n", err))
					continue
				}

			}

			if len(ids) > 0 {
				filter := bson.M{"_id": bson.M{"$in": ids}}
				_, err := repo.DeleteMany(ctx, "screenshot", filter)
				if err != nil {
					fmt.Printf("Failed to delete icon data: %v\n", err)
				}
			}

			time.Sleep(3 * time.Second)
		}
	}
}
