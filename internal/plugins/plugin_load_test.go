package plugins

import (
	"recon/internal/models"
	"recon/internal/options"
	"testing"
)

func TestStoredPluginLoadsWithCurrentExportedTypes(t *testing.T) {
	plugin, err := LoadPlugin(`package plugin
import (
 oldmodels "example.com/previous-collection/internal/models"
 oldoptions "example.com/previous-collection/internal/options"
)
func Execute(op oldoptions.PluginOption) error { return nil }
func TaskEnd(task oldmodels.Task, hash string) error { return nil }
func Cycle() string { return "1" }
func Install() error { return nil }
func GetName() string { return "compatibility-fixture" }
`, "compatibility-fixture")
	if err != nil {
		t.Fatalf("stored plugin failed to load: %v", err)
	}
	if plugin.GetName() != "compatibility-fixture" || plugin.Cycle() != "1" {
		t.Fatal("plugin callbacks were not resolved")
	}
	if err := plugin.Install(); err != nil {
		t.Fatal(err)
	}
	if err := plugin.Execute(options.PluginOption{}); err != nil {
		t.Fatal(err)
	}
	if err := plugin.TaskEnd(models.Task{}, "compatibility-fixture"); err != nil {
		t.Fatal(err)
	}
}
