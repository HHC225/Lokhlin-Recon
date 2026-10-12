package mongodb

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"recon/internal/logger"
	"recon/internal/utils/helper"

	"go.mongodb.org/mongo-driver/mongo/options"
	"recon/internal/constants"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"recon/internal/config"
	"recon/internal/utils/random"
)

func printProgressBar(step, total int, action string) {
	width := 50
	progress := float64(step) / float64(total)
	bar := int(progress * float64(width))

	fmt.Printf("\r[%s%s] %d/%d %s\n",
		strings.Repeat("=", bar),
		strings.Repeat(" ", width-bar),
		step, total, action)
}

func CreateDatabase() error {
	client := Client
	if client == nil {
		return fmt.Errorf("MongoDB client is not initialized")
	}

	databases, err := client.ListDatabaseNames(context.Background(), bson.M{})
	if err != nil {
		return fmt.Errorf("failed to list databases: %v", err)
	}

	db := client.Database(config.GlobalConfig.MongoDB.Database)

	if !contains(databases, config.GlobalConfig.MongoDB.Database) {
		totalSteps := 13
		currentStep := 0

		collection := db.Collection("user")
		password, err := random.GeneratePassword(16)
		if err != nil {
			password = random.GenerateRandomString(16)
		}

		fmt.Println("Initial administrator credentials are stored in PASSWORD")

		if err := os.WriteFile(filepath.Join(config.GlobalConfig.System.ExeDir, "PASSWORD"), []byte(password), 0600); err != nil {
			return fmt.Errorf("failed to write password file: %v", err)
		}

		hashedPassword := helper.Sha256Hex(password)

		_, err = collection.InsertOne(context.Background(), bson.M{
			"username": "Recon",
			"password": hashedPassword,
		})
		if err != nil {
			return fmt.Errorf("failed to create user: %v", err)
		}
		currentStep++
		printProgressBar(currentStep, totalSteps, "Creating user")

		configCollection := db.Collection("config")

		_, err = configCollection.InsertOne(context.Background(), bson.M{
			"name":  "timezone",
			"value": "Asia/Shanghai",
			"type":  "system",
		})
		if err != nil {
			return fmt.Errorf("failed to insert timezone config: %v", err)
		}
		configCollection.InsertOne(context.Background(), bson.M{
			"name":  "ModulesConfig",
			"value": constants.ModulesConfig,
			"type":  "system",
		})

		configCollection.InsertOne(context.Background(), bson.M{
			"name":  "SubfinderApiConfig",
			"value": constants.SubfinderApiConfig,
			"type":  "subfinder",
		})

		configCollection.InsertOne(context.Background(), bson.M{
			"name":  "RadConfig",
			"value": constants.RadConfig,
			"type":  "rad",
		})

		configCollection.InsertOne(context.Background(), bson.M{
			"name":  "FingerVersion",
			"value": "2026-01-20 23:50",
			"type":  "finger",
		})

		currentStep++
		printProgressBar(currentStep, totalSteps, "Setting timezone")

		_, err = configCollection.InsertOne(context.Background(), bson.M{
			"name":                          "notification",
			"dirScanNotification":           true,
			"portScanNotification":          true,
			"sensitiveNotification":         true,
			"subdomainTakeoverNotification": true,
			"pageMonNotification":           true,
			"subdomainNotification":         true,
			"vulNotification":               true,
			"type":                          "notification",
		})
		if err != nil {
			return fmt.Errorf("failed to insert notification config: %v", err)
		}
		currentStep++
		printProgressBar(currentStep, totalSteps, "Setting notification config")

		scheduledTasksCollection := db.Collection("ScheduledTasks")
		_, err = scheduledTasksCollection.InsertOne(context.Background(), bson.M{
			"id":      "page_monitoring",
			"name":    "Page Monitoring",
			"hour":    24,
			"node":    []string{},
			"allNode": true,
			"type":    "Page Monitoring",
			"state":   true,
		})
		if err != nil {
			return fmt.Errorf("failed to insert scheduled task: %v", err)
		}
		currentStep++
		printProgressBar(currentStep, totalSteps, "Creating scheduled tasks")

		err = db.CreateCollection(context.Background(), "notification")
		if err != nil {
			return fmt.Errorf("failed to create notification collection: %v", err)
		}
		currentStep++
		printProgressBar(currentStep, totalSteps, "Creating notification collection")

		dictionaryCollection := db.Collection("dictionary")

		dirDict := constants.DirDict
		size := float64(len(dirDict)) / (1024 * 1024)
		result, err := dictionaryCollection.InsertOne(context.Background(), bson.M{
			"name":     "default",
			"category": "dir",
			"size":     fmt.Sprintf("%.2f", size),
		})
		if err != nil {
			return fmt.Errorf("failed to insert dir dictionary: %v", err)
		}

		if err := CreateGridFSFile(db, result.InsertedID.(primitive.ObjectID).Hex(), []byte(dirDict)); err != nil {
			return fmt.Errorf("failed to write dir dictionary: %v", err)
		}
		currentStep++
		printProgressBar(currentStep, totalSteps, "Creating dir dictionary")

		domainDict := constants.DomainDict
		size = float64(len(domainDict)) / (1024 * 1024)
		result, err = dictionaryCollection.InsertOne(context.Background(), bson.M{
			"name":     "default",
			"category": "subdomain",
			"size":     fmt.Sprintf("%.2f", size),
		})
		if err != nil {
			return fmt.Errorf("failed to insert subdomain dictionary: %v", err)
		}
		if err := CreateGridFSFile(db, result.InsertedID.(primitive.ObjectID).Hex(), []byte(domainDict)); err != nil {
			return fmt.Errorf("failed to write subdomain dictionary: %v", err)
		}
		currentStep++
		printProgressBar(currentStep, totalSteps, "Creating subdomain dictionary")

		sensitiveCollection := db.Collection("SensitiveRule")
		sensitiveData, _ := constants.GetSensitive()
		if len(sensitiveData) > 0 {
			_, err = sensitiveCollection.InsertMany(context.Background(), sensitiveData)
			if err != nil {
				return fmt.Errorf("failed to insert sensitive rules: %v", err)
			}
		}
		currentStep++
		printProgressBar(currentStep, totalSteps, "Creating sensitive rules")

		portCollection := db.Collection("PortDict")
		portData, _ := constants.GetPort()
		_, err = portCollection.InsertMany(context.Background(), portData)
		if err != nil {
			return fmt.Errorf("failed to insert port dictionary: %v", err)
		}
		currentStep++
		printProgressBar(currentStep, totalSteps, "Creating port dictionary")

		//pocCollection := db.Collection("PocList")
		//pocData := getPoc()
		//if len(pocData) > 0 {
		//	_, err = pocCollection.InsertMany(context.Background(), pocData)
		//	if err != nil {
		//		return fmt.Errorf("failed to insert POC: %v", err)
		//	}
		//}
		//currentStep++
		//printProgressBar(currentStep, totalSteps, "Creating POC")

		fingerprintCollection := db.Collection("FingerprintRules")
		fingerprint, _ := constants.GetFingerprintData()
		if len(fingerprint) > 0 {
			_, err = fingerprintCollection.InsertMany(context.Background(), fingerprint)
			if err != nil {
				return fmt.Errorf("failed to insert fingerprint rules: %v", err)
			}
		}

		currentStep++
		printProgressBar(currentStep, totalSteps, "Creating fingerprint rules")

		pluginsCollection := db.Collection("plugins")
		var plgDocs []interface{}
		for _, p := range constants.Plugins {
			plgDocs = append(plgDocs, p)
		}
		_, err = pluginsCollection.InsertMany(context.Background(), plgDocs)
		if err != nil {
			return fmt.Errorf("failed to insert plugins: %v", err)
		}
		currentStep++
		printProgressBar(currentStep, totalSteps, "Creating plugins")

		scanTemplateCollection := db.Collection("ScanTemplates")
		_, err = scanTemplateCollection.InsertOne(context.Background(), constants.ScanTemplateDefault)
		if err != nil {
			return fmt.Errorf("failed to insert scan template: %v", err)
		}
		currentStep++
		printProgressBar(currentStep, totalSteps, "Creating scan template")

		if err := createIndexes(db); err != nil {
			return fmt.Errorf("failed to create indexes: %v", err)
		}
		currentStep++
		printProgressBar(currentStep, totalSteps, "Creating indexes")

		fmt.Println()
		logger.Info("Project initialization successful")
	} else {

		configCollection := db.Collection("config")
		var result bson.M
		err := configCollection.FindOne(context.Background(), bson.M{"name": "timezone"}).Decode(&result)
		if err != nil {
			return fmt.Errorf("failed to get timezone config: %v", err)
		}

		scheduledTasksCollection := db.Collection("ScheduledTasks")
		var taskResult bson.M
		err = scheduledTasksCollection.FindOne(context.Background(), bson.M{"id": "page_monitoring"}).Decode(&taskResult)
		if err != nil {
			if err == mongo.ErrNoDocuments {

				_, err = scheduledTasksCollection.InsertOne(context.Background(), bson.M{
					"id":    "page_monitoring",
					"name":  "Page Monitoring",
					"hour":  24,
					"type":  "Page Monitoring",
					"state": true,
				})
				if err != nil {
					return fmt.Errorf("failed to insert default scheduled task: %v", err)
				}
			} else {
				return fmt.Errorf("failed to check scheduled task: %v", err)
			}
		}
	}

	return nil
}

func createIndexes(db *mongo.Database) error {

	pageMonitoring := db.Collection("PageMonitoring")
	_, err := pageMonitoring.Indexes().CreateOne(context.Background(), mongo.IndexModel{
		Keys:    bson.D{{"url", 1}},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		return fmt.Errorf("failed to create PageMonitoring index: %v", err)
	}

	pageMonitoringBody := db.Collection("PageMonitoringBody")
	_, err = pageMonitoringBody.Indexes().CreateOne(context.Background(), mongo.IndexModel{
		Keys:    bson.D{{"md5", 1}},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		return fmt.Errorf("failed to create PageMonitoringBody index: %v", err)
	}

	asset := db.Collection("asset")
	indexes := []mongo.IndexModel{
		{Keys: bson.D{{"time", -1}}},
		{Keys: bson.D{{"url", 1}}},
		{Keys: bson.D{{"ip", 1}}},
		{Keys: bson.D{{"port", 1}}},
		{Keys: bson.D{{"host", 1}, {"port", 1}}, Options: options.Index().SetUnique(true)},
		{Keys: bson.D{{"project", 1}, {"time", -1}}},
		{Keys: bson.D{{"project", 1}}},
		{Keys: bson.D{{"taskName", 1}}},
		{Keys: bson.D{{"rootDomain", 1}}},

		{Keys: bson.D{{"tags", 1}}},
		{Keys: bson.D{{"technologies", 1}}},
		{Keys: bson.D{{"faviconmmh3", 1}}},
		{Keys: bson.D{{"bodyhash", 1}}},

		{Keys: bson.D{{"host", 1}}},

		{Keys: bson.D{{"rootDomain", 1}, {"time", -1}}},
	}
	_, err = asset.Indexes().CreateMany(context.Background(), indexes)
	if err != nil {
		return fmt.Errorf("failed to create asset indexes: %v", err)
	}

	ipAssetTmp := db.Collection("IPAssetTmp")
	indexes = []mongo.IndexModel{
		{Keys: bson.D{{"ip", 1}, {"port", 1}, {"domain", 1}}, Options: options.Index().SetUnique(true)},
	}
	_, err = ipAssetTmp.Indexes().CreateMany(context.Background(), indexes)
	if err != nil {
		return fmt.Errorf("failed to create ipAssetTmp indexes: %v", err)
	}

	ipAsset := db.Collection("IPAsset")
	indexes = []mongo.IndexModel{
		{Keys: bson.D{{"time", -1}}},
		{Keys: bson.D{{"ports.server.domain", 1}}},
		{Keys: bson.D{{"ports.port", 1}}},
		{Keys: bson.D{{"ip", 1}}, Options: options.Index().SetUnique(true)},
		{Keys: bson.D{{"ports.server.service", 1}}},
		{Keys: bson.D{{"ports.server.webServer", 1}}},
		{Keys: bson.D{{"rootDomain", 1}}},
		{Keys: bson.D{{"ports.server.technologies", 1}}},
		{Keys: bson.D{{"project", 1}}},
		{Keys: bson.D{{"taskName", 1}}},
	}
	_, err = ipAsset.Indexes().CreateMany(context.Background(), indexes)
	if err != nil {
		return fmt.Errorf("failed to create ipAsset indexes: %v", err)
	}

	icon := db.Collection("icon")
	indexes = []mongo.IndexModel{
		{Keys: bson.D{{"fav3", 1}}, Options: options.Index().SetUnique(true)},
	}
	_, err = icon.Indexes().CreateMany(context.Background(), indexes)
	if err != nil {
		return fmt.Errorf("failed to create icon indexes: %v", err)
	}

	httpBody := db.Collection("HttpBody")
	indexes = []mongo.IndexModel{
		{Keys: bson.D{{"hash", 1}}, Options: options.Index().SetUnique(true)},
	}
	_, err = httpBody.Indexes().CreateMany(context.Background(), indexes)
	if err != nil {
		return fmt.Errorf("failed to create httpBody indexes: %v", err)
	}

	httpScreenshot := db.Collection("screenshot")
	indexes = []mongo.IndexModel{
		{Keys: bson.D{{"hash", 1}}, Options: options.Index().SetUnique(true)},
	}
	_, err = httpScreenshot.Indexes().CreateMany(context.Background(), indexes)
	if err != nil {
		return fmt.Errorf("failed to create httpBody indexes: %v", err)
	}

	subdomain := db.Collection("subdomain")
	indexes = []mongo.IndexModel{
		{Keys: bson.D{{"project", 1}}},
		{Keys: bson.D{{"taskName", 1}}},
		{Keys: bson.D{{"rootDomain", 1}}},
		{Keys: bson.D{{"time", 1}}},
	}
	_, err = subdomain.Indexes().CreateMany(context.Background(), indexes)
	if err != nil {
		return fmt.Errorf("failed to create subdomain indexes: %v", err)
	}

	urlScan := db.Collection("UrlScan")
	indexes = []mongo.IndexModel{
		{Keys: bson.D{{"project", 1}}},
		{Keys: bson.D{{"taskName", 1}}},
		{Keys: bson.D{{"rootDomain", 1}}},
	}
	_, err = urlScan.Indexes().CreateMany(context.Background(), indexes)
	if err != nil {
		return fmt.Errorf("failed to create UrlScan indexes: %v", err)
	}

	SensitiveBody := db.Collection("SensitiveBody")
	_, err = SensitiveBody.Indexes().CreateOne(context.Background(), mongo.IndexModel{
		Keys: bson.D{{"md5", 1}},
	})
	if err != nil {
		fmt.Errorf("failed to create SensitiveBody index: %v", err)
	}

	crawler := db.Collection("crawler")
	indexes = []mongo.IndexModel{
		{Keys: bson.D{{"project", 1}}},
		{Keys: bson.D{{"taskName", 1}}},
		{Keys: bson.D{{"rootDomain", 1}}},
	}
	_, err = crawler.Indexes().CreateMany(context.Background(), indexes)
	if err != nil {
		return fmt.Errorf("failed to create crawler indexes: %v", err)
	}

	sensitiveResult := db.Collection("SensitiveResult")
	indexes = []mongo.IndexModel{
		{Keys: bson.D{{"project", 1}}},
		{Keys: bson.D{{"taskName", 1}}},
		{Keys: bson.D{{"rootDomain", 1}}},
	}
	_, err = sensitiveResult.Indexes().CreateMany(context.Background(), indexes)
	if err != nil {
		return fmt.Errorf("failed to create SensitiveResult indexes: %v", err)
	}

	dirScanResult := db.Collection("DirScanResult")
	indexes = []mongo.IndexModel{
		{Keys: bson.D{{"project", 1}}},
		{Keys: bson.D{{"taskName", 1}}},
		{Keys: bson.D{{"rootDomain", 1}}},
	}
	_, err = dirScanResult.Indexes().CreateMany(context.Background(), indexes)
	if err != nil {
		return fmt.Errorf("failed to create DirScanResult indexes: %v", err)
	}

	vulnerability := db.Collection("vulnerability")
	indexes = []mongo.IndexModel{
		{Keys: bson.D{{"project", 1}}},
		{Keys: bson.D{{"taskName", 1}}},
		{Keys: bson.D{{"rootDomain", 1}}},
		{Keys: bson.D{{"hash", 1}}},
	}
	_, err = vulnerability.Indexes().CreateMany(context.Background(), indexes)
	if err != nil {
		return fmt.Errorf("failed to create vulnerability indexes: %v", err)
	}

	vulnerabilityDetail := db.Collection("vulnerabilityDetail")
	indexes = []mongo.IndexModel{
		{Keys: bson.D{{"hash", 1}}},
	}
	_, err = vulnerabilityDetail.Indexes().CreateMany(context.Background(), indexes)
	if err != nil {
		return fmt.Errorf("failed to create vulnerabilityDetail indexes: %v", err)
	}

	rootDomain := db.Collection("RootDomain")
	indexes = []mongo.IndexModel{
		{Keys: bson.D{{"project", 1}}},
		{Keys: bson.D{{"taskName", 1}}},
		{Keys: bson.D{{"domain", 1}}, Options: options.Index().SetUnique(true)},
		{Keys: bson.D{{"time", 1}}},
	}
	_, err = rootDomain.Indexes().CreateMany(context.Background(), indexes)
	if err != nil {
		return fmt.Errorf("failed to create RootDomain indexes: %v", err)
	}

	app := db.Collection("app")
	indexes = []mongo.IndexModel{
		{Keys: bson.D{{"project", 1}}},
		{Keys: bson.D{{"taskName", 1}}},
		{Keys: bson.D{{"time", 1}}},
		{Keys: bson.D{{"name", 1}}},
	}
	_, err = app.Indexes().CreateMany(context.Background(), indexes)
	if err != nil {
		return fmt.Errorf("failed to create app indexes: %v", err)
	}

	mp := db.Collection("mp")
	indexes = []mongo.IndexModel{
		{Keys: bson.D{{"project", 1}}},
		{Keys: bson.D{{"taskName", 1}}},
		{Keys: bson.D{{"time", 1}}},
		{Keys: bson.D{{"name", 1}}},
	}
	_, err = mp.Indexes().CreateMany(context.Background(), indexes)
	if err != nil {
		return fmt.Errorf("failed to create mp indexes: %v", err)
	}

	task := db.Collection("task")
	indexes = []mongo.IndexModel{
		{Keys: bson.D{{"creatTime", -1}}},
		{Keys: bson.D{{"name", 1}}},
		{Keys: bson.D{{"progress", 1}}},
	}
	_, err = task.Indexes().CreateMany(context.Background(), indexes)
	if err != nil {
		return fmt.Errorf("failed to create task indexes: %v", err)
	}

	return nil
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
