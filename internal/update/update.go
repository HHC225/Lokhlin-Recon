// config-------------------------------------
// @file      : update.go
// @time      : 2025/5/10 20:32
// -------------------------------------------

package update

import (
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/mod/semver"
	"recon/internal/constants"
	"recon/internal/database/mongodb"
	"strings"
)

func normalizeVersion(version string) string {
	if !strings.HasPrefix(version, "v") {
		return "v" + version
	}
	return version
}

func Update() error {

	Update20()

	configColl := mongodb.DB.Collection("config")
	var result struct {
		Name    string `bson:"name"`
		Version string `bson:"version"`
		Update  bool   `bson:"update"`
	}
	err := configColl.FindOne(context.Background(), bson.M{"name": "version"}).Decode(&result)
	if errors.Is(err, mongo.ErrNoDocuments) {

		_, err = configColl.InsertOne(context.Background(), bson.M{
			"name":    "version",
			"version": constants.Version,
			"update":  false,
		})
		if err != nil {
			return err
		}
		result.Version = constants.Version
		result.Update = false
	} else if err != nil {
		return err
	}
	currentVersion := normalizeVersion(result.Version)
	if semver.Compare(currentVersion, normalizeVersion(constants.Version)) < 0 {
		result.Update = false
	}
	needUpdate := !result.Update
	if needUpdate {
		//if semver.Compare(currentVersion, "v1.4") < 0 {
		//	update14(db)
		//}
		//if semver.Compare(currentVersion, "v1.5.0") < 0 {
		//	update15(db)
		//}
		//if semver.Compare(currentVersion, "v1.6.0") < 0 {
		//	update16(db)
		//}
		//if semver.Compare(currentVersion, "v1.7.0") < 0 {
		//	update17(db)
		//}

		if semver.Compare(currentVersion, "v1.8.0") < 0 {
			Update18()
		}

		if semver.Compare(currentVersion, "v1.9.0") < 0 {
			Update19()
		}

		_, err := configColl.UpdateOne(context.Background(), bson.M{"name": "version"}, bson.M{
			"$set": bson.M{
				"version": constants.Version,
				"update":  true,
			},
		})
		if err != nil {
			return err
		}
	}
	return nil
}
