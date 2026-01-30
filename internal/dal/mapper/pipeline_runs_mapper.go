package mapper

import (
	"context"
	"devops-console-backend/internal/dal/model"
	"devops-console-backend/internal/dal/query"

	"gorm.io/gorm"
)

type PipelineRunMapper struct {
	db    *gorm.DB
	query *query.Query
}

func NewPipelineRunMapper(db *gorm.DB) *PipelineRunMapper {
	return &PipelineRunMapper{
		db:    db,
		query: query.Use(db),
	}
}

func (p *PipelineRunMapper) GetPipelineRunById(id uint64) (*model.PipelineRun, error) {
	return p.query.PipelineRun.WithContext(context.Background()).Where(p.query.PipelineRun.ID.Eq(id)).First()
}

func (p *PipelineRunMapper) UpdatePipelineRun(pipelineRun *model.PipelineRun) error {
	_, err := p.query.PipelineRun.WithContext(context.Background()).Where(p.query.PipelineRun.ID.Eq(pipelineRun.ID)).Updates(pipelineRun)
	return err
}

func (p *PipelineRunMapper) CreatePipelineRun(pipelineRun *model.PipelineRun) error {
	return p.query.PipelineRun.WithContext(context.Background()).Create(pipelineRun)
}
func (p *PipelineRunMapper) DeletePipelineRun(id uint64) error {
	_, err := p.query.PipelineRun.WithContext(context.Background()).Where(p.query.PipelineRun.ID.Eq(id)).Delete()
	return err
}
func (p *PipelineRunMapper) GetPagePipelineRuns(pageNum int, pageSize int) ([]*model.PipelineRun, int64, error) {
	info, err := p.query.PipelineRun.WithContext(context.Background()).Limit(pageSize).Offset((pageNum - 1) * pageSize).Find()
	count, err := p.query.PipelineRun.Count()
	return info, count, err
}
