package mapper

import (
	"devops-console-backend/internal/dal/model"
	"devops-console-backend/internal/dal/query"

	"gorm.io/gorm"
)

type PipelineStepsMapper struct {
	db    *gorm.DB
	query *query.Query
}

func NewPipelineStepsMapper(db *gorm.DB) *PipelineStepsMapper {
	return &PipelineStepsMapper{
		db:    db,
		query: query.Use(db),
	}
}

func (p *PipelineStepsMapper) GetPipelineStepById(id uint32) (*model.PipelineStep, error) {
	return p.query.PipelineStep.WithContext(p.db.Statement.Context).Where(p.query.PipelineStep.ID.Eq(id)).First()
}

func (p *PipelineStepsMapper) UpdatePipelineStep(pipelineStep *model.PipelineStep) error {
	_, err := p.query.PipelineStep.WithContext(p.db.Statement.Context).Where(p.query.PipelineStep.ID.Eq(pipelineStep.ID)).Updates(pipelineStep)
	return err
}

func (p *PipelineStepsMapper) CreatePipelineStep(pipelineStep *model.PipelineStep) error {
	return p.query.PipelineStep.WithContext(p.db.Statement.Context).Create(pipelineStep)
}

func (p *PipelineStepsMapper) DeletePipelineStep(id uint32) error {
	_, err := p.query.PipelineStep.WithContext(p.db.Statement.Context).Where(p.query.PipelineStep.ID.Eq(id)).Delete()
	return err
}

func (p *PipelineStepsMapper) GetPipelineStepByPipelineId(pipelineId uint32) ([]*model.PipelineStep, error) {
	return p.query.PipelineStep.WithContext(p.db.Statement.Context).
		Where(p.query.PipelineStep.PipelineID.Eq(pipelineId)).
		Order(p.query.PipelineStep.Sort).
		Find()
}
