package service

import (
	"context"
	"github.com/Duke1616/ecmdb/internal/relation/internal/domain"
	"github.com/Duke1616/ecmdb/internal/relation/internal/repository"
	"golang.org/x/sync/errgroup"
)

type RelationModelService interface {
	CreateModelRelation(ctx context.Context, req domain.ModelRelation) (int64, error)
	ListModelRelation(ctx context.Context, offset, limit int64) ([]domain.ModelRelation, int64, error)
	ListModelUidRelation(ctx context.Context, offset, limit int64, modelUid string) ([]domain.ModelRelation, int64, error)

	// ListSrcModelByUid 根据源模型UID 查询所有的关联的模型
	ListSrcModelByUid(ctx context.Context, sourceUId string) ([]domain.ModelDiagram, error)
	ListDstModelByUid(ctx context.Context, sourceUId string) ([]domain.ModelDiagram, error)

	ListSrcModelByUIDs(ctx context.Context, srcUids []string) ([]domain.ModelDiagram, error)
}

type modelService struct {
	repo repository.RelationModelRepository
}

func NewRelationModelService(repo repository.RelationModelRepository) RelationModelService {
	return &modelService{
		repo: repo,
	}
}

func (s *modelService) CreateModelRelation(ctx context.Context, req domain.ModelRelation) (int64, error) {
	return s.repo.CreateModelRelation(ctx, req)
}

func (s *modelService) ListModelRelation(ctx context.Context, offset, limit int64) ([]domain.ModelRelation, int64, error) {
	relation, err := s.repo.ListModelRelation(ctx, offset, limit)
	if err != nil {
		return nil, 0, err
	}

	return relation, 0, nil
}

func (s *modelService) ListModelUidRelation(ctx context.Context, offset, limit int64, modelUid string) ([]domain.ModelRelation, int64, error) {
	var (
		eg        errgroup.Group
		relations []domain.ModelRelation
		total     int64
	)
	eg.Go(func() error {
		var err error
		relations, err = s.repo.ListRelationByModelUid(ctx, offset, limit, modelUid)
		return err
	})

	eg.Go(func() error {
		var err error
		total, err = s.repo.TotalByModelUid(ctx, modelUid)
		return err
	})
	return relations, total, eg.Wait()
}

func (s *modelService) ListSrcModelByUid(ctx context.Context, modelId string) ([]domain.ModelDiagram, error) {
	return s.repo.ListSrcModelByUid(ctx, modelId)
}

func (s *modelService) ListDstModelByUid(ctx context.Context, modelId string) ([]domain.ModelDiagram, error) {
	return s.repo.ListSrcModelByUid(ctx, modelId)
}

func (s *modelService) ListSrcModelByUIDs(ctx context.Context, srcUids []string) ([]domain.ModelDiagram, error) {
	return s.repo.ListSrcModelByUIDs(ctx, srcUids)
}
