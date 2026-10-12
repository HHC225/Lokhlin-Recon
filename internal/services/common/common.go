// common-------------------------------------
// @file      : common.go
// @time      : 2025/9/25 23:40
// -------------------------------------------

package common

import (
	"recon/internal/repositories/common"
)

type Service struct {
	//taskRepo         task.Repository
	//commonService    commonTask.Service
	//schedulerService schedulerSvc.Service
	//nodeService      node.Service
	Repo common.Repository
}

func NewService() Service {
	return Service{
		Repo: common.NewRepository(),
	}
}
