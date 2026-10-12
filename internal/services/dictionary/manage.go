package dictionary

import (
	"context"
	"fmt"

	"recon/internal/models"
	dictrepo "recon/internal/repositories/dictionary"
	nodessvc "recon/internal/services/node"
)

type ManageService interface {
	List(ctx context.Context) ([]models.DictionaryMeta, error)
	Create(ctx context.Context, name, category string, content []byte) error
	Download(ctx context.Context, id string) (filename string, data []byte, err error)
	Delete(ctx context.Context, ids []string) error
	Save(ctx context.Context, id string, content []byte) error
}

type manageService struct {
	repo        dictrepo.ManageRepository
	nodeService nodessvc.Service
}

func NewManageService() ManageService {
	return &manageService{
		repo:        dictrepo.NewManageRepository(),
		nodeService: nodessvc.NewService(),
	}
}

func (s *manageService) List(ctx context.Context) ([]models.DictionaryMeta, error) {
	return s.repo.List(ctx)
}

func (s *manageService) Create(ctx context.Context, name, category string, content []byte) error {

	exists, err := s.repo.ExistsByNameCategory(ctx, name, category)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("duplication file name")
	}

	id, err := s.repo.InsertMeta(ctx, name, category, fmt.Sprintf("%.2f", float64(len(content))/(1024*1024)))
	if err != nil {
		return err
	}
	if err := s.repo.UploadFile(ctx, id, content); err != nil {
		return err
	}

	_ = s.nodeService.RefreshConfig(ctx, models.Message{
		Name:    "all",
		Type:    "dictionary",
		Content: fmt.Sprintf("add:%s", id),
	})
	return nil
}

func (s *manageService) Download(ctx context.Context, id string) (string, []byte, error) {
	data, err := s.repo.DownloadFile(ctx, id)
	if err != nil {
		return "", nil, err
	}
	return id, data, nil
}

func (s *manageService) Delete(ctx context.Context, ids []string) error {
	if err := s.repo.DeleteMeta(ctx, ids); err != nil {
		return err
	}

	for _, id := range ids {
		if err := s.repo.DeleteFile(ctx, id); err != nil {
			return err
		}
		_ = s.nodeService.RefreshConfig(ctx, models.Message{
			Name:    "all",
			Type:    "dictionary",
			Content: fmt.Sprintf("delete:%s", id),
		})
	}
	return nil
}

func (s *manageService) Save(ctx context.Context, id string, content []byte) error {

	if err := s.repo.ReplaceFile(ctx, id, content); err != nil {
		return err
	}

	if err := s.repo.UpdateMetaSize(ctx, id, fmt.Sprintf("%.2f", float64(len(content))/(1024*1024))); err != nil {
		return err
	}
	_ = s.nodeService.RefreshConfig(ctx, models.Message{
		Name:    "all",
		Type:    "dictionary",
		Content: fmt.Sprintf("add:%s", id),
	})
	return nil
}
