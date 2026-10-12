// plugin-------------------------------------
// @file      : scheduler.go
// @time      : 2025/12/20 17:50
// -------------------------------------------

package plugin

import (
	"fmt"

	schedulerCore "recon/internal/scheduler"
)

func AddJob(hash string, schedule string) error {
	job := &schedulerCore.Job{
		ID:        hash,
		Name:      "server_plugin",
		Handler:   "server_plugin",
		Params:    hash,
		CycleType: "cron",
		Schedule:  schedule,
	}

	err := schedulerCore.GetGlobalScheduler().RemoveJob(job.ID)
	if err != nil {
		return err
	}
	if err := schedulerCore.GetGlobalScheduler().AddJob(job); err != nil {
		return fmt.Errorf("failed to add scheduled task: %w", err)
	}
	return nil
}
