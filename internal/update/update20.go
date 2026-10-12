package update

import (
	"context"

	"recon/internal/repositories/apikey"
)

func Update20() {
	repo := apikey.NewRepository()
	_ = repo.EnsureIndexes(context.Background())
}
