// symbols-------------------------------------
// @file      : symbols.go
// @time      : 2025/12/6 15:54
// -------------------------------------------

//go:generate yaegi extract recon/internal/options
//go:generate yaegi extract recon/internal/utils
//go:generate yaegi extract recon/internal/utils/ct
//go:generate yaegi extract recon/internal/utils/ct/models
//go:generate yaegi extract recon/internal/utils/ct/types
//go:generate yaegi extract recon/internal/utils/random
//go:generate yaegi extract recon/internal/utils/helper
//go:generate yaegi extract recon/internal/models
//go:generate yaegi extract go.mongodb.org/mongo-driver/bson
//go:generate yaegi extract go.mongodb.org/mongo-driver/bson/primitive
//go:generate yaegi extract go.mongodb.org/mongo-driver/mongo
//go:generate yaegi extract go.mongodb.org/mongo-driver/mongo/options
package symbols

import (
	"github.com/traefik/yaegi/stdlib"
	"reflect"
)

var Symbols = map[string]map[string]reflect.Value{}

func init() {

	for pkg, symbols := range map[string]string{
		"os/exec": "os/exec",
	} {
		Symbols[pkg] = stdlib.Symbols[symbols]
	}
}
