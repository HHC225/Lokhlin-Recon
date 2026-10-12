// worker-------------------------------------
// @file      : iconhandle.go
// @time      : 2025/11/1 18:25
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

func IconHandle(ctx context.Context) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Worker error recovered:", r)

			select {
			case <-ctx.Done():
				fmt.Println("Task cancelled; IconHandle will not restart")
			default:
				go IconHandle(ctx)
			}
		}
	}()

	repo := common.NewRepository()

	for {
		select {
		case <-ctx.Done():
			fmt.Println("Task cancelled; IconHandle stopped")
			return
		default:

			limit := int64(100)
			opts := options.Find().SetLimit(limit)

			results, err := repo.FindMany(ctx, "icon", bson.M{}, opts)
			if err != nil {
				logger.Error(fmt.Sprintf("Failed to query icon data: %v\n", err))
				time.Sleep(10 * time.Second)
				continue
			}

			if len(results) == 0 {
				time.Sleep(60 * time.Second)
				continue
			}

			var ids []primitive.ObjectID

			for _, item := range results {

				fav3, _ := item["fav3"].(string)
				content, _ := item["content"].(string)
				filepath := path.Join(config.GlobalConfig.System.IconDir, fmt.Sprintf("%v.png", fav3))
				decoded, err := base64.StdEncoding.DecodeString(content)
				if err != nil {
					logger.Error(fmt.Sprintf("Failed to decode icon base64: %v %v\n", err, content))
					continue
				}
				err = utils.WriteFile(filepath, decoded)
				if err != nil {
					logger.Error(fmt.Sprintf("Failed to write icon: %v\n", err))
					continue
				}

				if id, ok := item["_id"].(primitive.ObjectID); ok {
					ids = append(ids, id)
				}
			}

			if len(ids) > 0 {
				filter := bson.M{"_id": bson.M{"$in": ids}}
				_, err := repo.DeleteMany(ctx, "icon", filter)
				if err != nil {
					fmt.Printf("Failed to delete icon data: %v\n", err)
				}
			}

			time.Sleep(3 * time.Second)
		}
	}
}
