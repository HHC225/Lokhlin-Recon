// interfaces-------------------------------------
// @file      : plugin.go
// @time      : 2025/12/6 16:08
// -------------------------------------------

package interfaces

import (
	"recon/internal/models"
	"recon/internal/options"
)

type Plugin interface {
	GetName() string
	Cycle() string
	Install() error
	Execute(op options.PluginOption) error
	TaskEnd(task models.Task, plgHash string) error
	Clone() Plugin
	Log(msg string, tp ...string)
	GetPluginId() string
}
