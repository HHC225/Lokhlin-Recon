package constants

import (
	"os"
	"strings"
)

// DisplayName centralizes the temporary product label without changing protocols.
func DisplayName() string {
	if name := strings.TrimSpace(os.Getenv("RECON_DISPLAY_NAME")); name != "" {
		return name
	}
	return "Recon"
}
