package mapper

import (
	"context"
	"devops-console-backend/internal/dal/model"
	"devops-console-backend/internal/dal/query"

	"gorm.io/gorm"
)

type PipelinesMapper struct {
	db    *gorm.DB
	query *query.Query
}

func NewPipelinesMapper(db *gorm.DB) *PipelinesMapper {
	return &PipelinesMapper{
		db:    db,
		query: query.Use(db),
	}
}

func (p *PipelinesMapper) GetPipelineById(id uint32) (*model.Pipeline, error) {
	return p.query.Pipeline.WithContext(context.Background()).Where(p.query.Pipeline.ID.Eq(id)).First()
}

func (p *PipelinesMapper) UpdatePipeline(pipeline *model.Pipeline) error {
	_, err := p.query.Pipeline.WithContext(context.Background()).Where(p.query.Pipeline.ID.Eq(pipeline.ID)).Updates(pipeline)
	return err
}
func (p *PipelinesMapper) CreatePipeline(pipeline *model.Pipeline) error {
	return p.query.Pipeline.WithContext(context.Background()).Create(pipeline)
}
func (p *PipelinesMapper) DeletePipeline(id uint32) error {
	_, err := p.query.Pipeline.WithContext(context.Background()).Where(p.query.Pipeline.ID.Eq(id)).Delete()
	return err
}
func (p *PipelinesMapper) GetPagePipelines(pageNum int, pageSize int) ([]*model.Pipeline, int64, error) {
	data, err := p.query.Pipeline.WithContext(context.Background()).Limit(pageSize).Offset((pageNum - 1) * pageSize).Find()
	count, err := p.query.Pipeline.WithContext(context.Background()).Count()
	return data, count, err
}
