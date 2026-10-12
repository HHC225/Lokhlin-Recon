// update-------------------------------------
// @file      : update181.go
// @time      : 2025/11/23 17:17
// -------------------------------------------

package update

import (
	"context"
	"fmt"
	"recon/internal/constants"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"recon/internal/database/mongodb"
	"recon/internal/logger"
)

func Update19() {

	asset := mongodb.DB.Collection("asset")
	indexes := []mongo.IndexModel{
		{Keys: bson.D{{"host", 1}}},
		{Keys: bson.D{{"rootDomain", 1}, {"time", -1}}},
	}
	_, err := asset.Indexes().CreateMany(context.Background(), indexes)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to create asset indexes: %v", err))
	}

	ctx := context.Background()
	fingerprintCollection := mongodb.DB.Collection("FingerprintRules")
	backupPipeline := []bson.M{
		{"$out": "FingerprintRules_bak"},
	}
	cursor, err := fingerprintCollection.Aggregate(ctx, backupPipeline)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to backup FingerprintRules: %v", err))
		return
	}
	defer cursor.Close(ctx)
	logger.Info("FingerprintRules backed up to FingerprintRules_bak successfully")

	fingerprintCollection = mongodb.DB.Collection("FingerprintRules")
	fingerprint, _ := constants.GetFingerprintData()
	if len(fingerprint) > 0 {
		_, err = fingerprintCollection.InsertMany(context.Background(), fingerprint)
		if err != nil {
			logger.Error(fmt.Sprintf("Failed to create FingerprintRules: %v", err))
		}
	}

	configCollection := mongodb.DB.Collection("config")
	configCollection.InsertOne(context.Background(), bson.M{
		"name":  "FingerVersion",
		"value": "2026-01-20 23:50",
		"type":  "finger",
	})
}
