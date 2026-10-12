package config

import "testing"

func TestGetMongoURIUsesDecimalPort(t *testing.T) {
	previous := GlobalConfig
	t.Cleanup(func() { GlobalConfig = previous })
	GlobalConfig.MongoDB = MongoDBConfig{IP: "mongodb://database.test", Port: 27017}
	if got := GetMongoURI(); got != "mongodb://database.test:27017" {
		t.Fatalf("MongoDB port was not rendered as decimal digits: %q", got)
	}
}
