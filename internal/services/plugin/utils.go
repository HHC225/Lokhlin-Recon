// plugin-------------------------------------
// @file      : utils.go
// @time      : 2025/12/20 18:42
// -------------------------------------------

package plugin

import (
	"go.uber.org/zap"
	"recon/internal/logger"
	"recon/internal/plugins"
)

func RegisterPlugin(source, hash string) {
	loadPlugin, err := plugins.LoadPlugin(source, hash)
	if err != nil {
		logger.Error("failed to load plugin", zap.Error(err))
		return
	}
	plugins.GlobalPluginManager.RegisterPlugin(hash, loadPlugin)
	err = loadPlugin.Install()
	if err != nil {
		logger.Error("failed to install plugin", zap.Error(err))
		return
	}
	cycle := loadPlugin.Cycle()

	if cycle != "" {

		if cycle == "1" {
			return
		}
		err := AddJob(hash, loadPlugin.Cycle())
		if err != nil {
			return
		}
	}
}
