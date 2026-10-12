package plugins

import (
	"go/parser"
	"go/token"
	"strconv"
	"strings"
)

// Match only project packages exported to plugins. This supports stored source
// from earlier module names without rewriting comments, literals or dependencies.
var pluginImportPackages = map[string]bool{
	"internal/models":          true,
	"internal/options":         true,
	"internal/utils":           true,
	"internal/utils/helper":    true,
	"internal/utils/random":    true,
	"internal/utils/ct":        true,
	"internal/utils/ct/models": true,
	"internal/utils/ct/types":  true,
}

func normalizePluginImports(source string) (string, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "plugin.go", source, parser.ImportsOnly)
	if err != nil {
		return "", err
	}
	var out strings.Builder
	position := 0
	for _, spec := range file.Imports {
		importPath, err := strconv.Unquote(spec.Path.Value)
		if err != nil {
			return "", err
		}
		separator := strings.LastIndex(importPath, "/internal/")
		if separator < 0 {
			continue
		}
		packagePath := importPath[separator+1:]
		if !pluginImportPackages[packagePath] || importPath == "recon/"+packagePath {
			continue
		}
		start := fset.Position(spec.Path.Pos()).Offset
		end := fset.Position(spec.Path.End()).Offset
		out.WriteString(source[position:start])
		out.WriteString(strconv.Quote("recon/" + packagePath))
		position = end
	}
	out.WriteString(source[position:])
	return out.String(), nil
}
