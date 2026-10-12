package plugins

import (
	"reflect"
	"strings"
	"testing"

	"github.com/traefik/yaegi/interp"
	"recon/internal/utils/random"
)

func TestStoredPluginImportsAreNormalizedWithoutChangingOtherSource(t *testing.T) {
	source := `package plugin
import (
  opts "example.com/previous-collection/internal/options"
  "recon/internal/models"
  "go.mongodb.org/mongo-driver/bson"
)
// Documentation: example.com/previous-collection/internal/options
const example = "example.com/previous-collection/internal/options"
`
	want := `package plugin
import (
  opts "recon/internal/options"
  "recon/internal/models"
  "go.mongodb.org/mongo-driver/bson"
)
// Documentation: example.com/previous-collection/internal/options
const example = "example.com/previous-collection/internal/options"
`
	got, err := normalizePluginImports(source)
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("import normalization modified the wrong source:\n%s", got)
	}
}

func TestNormalizedStoredPluginRunsInInterpreter(t *testing.T) {
	source := `package plugin
import oldrandom "example.com/previous-collection/internal/utils/random"
func Execute() string { return oldrandom.GenerateString(7) }
`
	code, err := normalizePluginImports(source)
	if err != nil {
		t.Fatal(err)
	}
	i := interp.New(interp.Options{})
	if err := i.Use(interp.Exports{
		"recon/internal/utils/random/random": {"GenerateString": reflect.ValueOf(random.GenerateString)},
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := i.Eval(code); err != nil {
		t.Fatalf("normalized stored plugin did not load: %v", err)
	}
	v, err := i.Eval("plugin.Execute()")
	if err != nil {
		t.Fatal(err)
	}
	if len(v.String()) != 7 {
		t.Fatalf("wrong plugin execution result: %q", v.String())
	}
}

func TestUnexportedAndThirdPartyPluginImportsAreNotRedirected(t *testing.T) {
	source := `package plugin
import (
 "example.com/previous-collection/internal/private"
 "example.com/library/utils/random"
 "go.mongodb.org/mongo-driver/bson"
)
`
	got, err := normalizePluginImports(source)
	if err != nil {
		t.Fatal(err)
	}
	if got != source {
		t.Fatalf("unrelated imports were redirected: %s", got)
	}
}

func TestMalformedStoredPluginImportsAreRejected(t *testing.T) {
	_, err := normalizePluginImports("package plugin\nimport \"unterminated")
	if err == nil || strings.TrimSpace(err.Error()) == "" {
		t.Fatal("malformed plugin imports were accepted")
	}
}
