// helper-------------------------------------
// @file      : time.go
// @time      : 2025/5/25 15:04
// -------------------------------------------

package helper

import (
	"recon/internal/config"
	"time"
)

func GetNowTimeString() string {
	loc, err := time.LoadLocation(config.GlobalConfig.System.Timezone)
	if err != nil {

		loc = time.UTC
	}
	now := time.Now().In(loc)
	return now.Format("2006-01-02 15:04:05")
}

func GetNowTime() time.Time {
	loc, err := time.LoadLocation(config.GlobalConfig.System.Timezone)
	if err != nil {

		loc = time.UTC
	}
	return time.Now().In(loc)
}

func FormatTime(t time.Time) string {
	loc, err := time.LoadLocation(config.GlobalConfig.System.Timezone)
	if err != nil {
		loc = time.UTC
	}
	return t.In(loc).Format("2006-01-02 15:04:05")
}
