// config-------------------------------------
// @file      : initdata.go
// @time      : 2025/5/13 22:34
// -------------------------------------------
package bootstrap

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"recon/internal/database/mongodb"
)

var ProjectList = make(map[string]string)

func GetProjectList() error {
	collection := mongodb.DB.Collection("project")
	opts := options.Find().SetProjection(bson.M{"_id": 1, "name": 1})
	cursor, err := collection.Find(context.Background(), bson.M{}, opts)
	if err != nil {
		return err
	}
	defer cursor.Close(context.Background())

	for cursor.Next(context.Background()) {
		var result struct {
			ID   string `bson:"_id"`
			Name string `bson:"name"`
		}
		if err := cursor.Decode(&result); err != nil {
			return err
		}
		ProjectList[result.ID] = result.Name
	}

	return cursor.Err()
}
