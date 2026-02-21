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
func (p *PipelineRunMapper) GetPagePipelineRuns(pageNum int, pageSize int, pipelineId uint64) ([]*model.PipelineRun, int64, error) {
	q := p.query.PipelineRun.WithContext(context.Background())
	cq := p.query.PipelineRun.WithContext(context.Background())
	if pipelineId > 0 {
		q = q.Where(p.query.PipelineRun.PipelineID.Eq(uint32(pipelineId)))
		cq = cq.Where(p.query.PipelineRun.PipelineID.Eq(uint32(pipelineId)))
	}
	info, err := q.Order(p.query.PipelineRun.ID.Desc()).Limit(pageSize).Offset((pageNum - 1) * pageSize).Find()
	if err != nil {
		return nil, 0, err
	}
	count, err := cq.Count()
	return info, count, err
}

func (p *PipelineRunMapper) GetLastPipelineRunByPipelineId(pipelineId uint64) (*model.PipelineRun, error) {
	run, err := p.query.PipelineRun.WithContext(context.Background()).Where(p.query.PipelineRun.PipelineID.Eq(uint32(pipelineId))).Order(p.query.PipelineRun.ID.Desc()).First()
	return run, err
}
