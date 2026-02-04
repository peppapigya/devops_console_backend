package mapper

import (
	"devops-console-backend/internal/dal/model"
	"devops-console-backend/internal/dal/query"

	"gorm.io/gorm"
)

type ProjectMapper struct {
	db    *gorm.DB
	query *query.Query
}

func NewProjectMapper(db *gorm.DB) *ProjectMapper {
	return &ProjectMapper{
		db:    db,
		query: query.Use(db),
	}
}

func (p *ProjectMapper) GetProjectById(id uint32) (*model.Project, error) {
	return p.query.Project.Where(p.query.Project.ID.Eq(id)).First()
}

func (p *ProjectMapper) UpdateProject(project *model.Project) error {
	_, err := p.query.Project.Where(p.query.Project.ID.Eq(project.ID)).Updates(project)
	return err
}

func (p *ProjectMapper) CreateProject(project *model.Project) error {
	return p.query.Project.Create(project)
}

func (p *ProjectMapper) DeleteProject(id uint32) error {
	_, err := p.query.Project.Where(p.query.Project.ID.Eq(id)).Delete()
	return err
}

// GetPageProjects 分页获取数据
func (p *ProjectMapper) GetPageProjects(pageNum int, pageSize int) ([]*model.Project, int64, error) {
	find, err := p.query.Project.Limit(pageSize).Offset((pageNum - 1) * pageSize).Find()
	total, err := p.query.Project.Count()
	return find, total, err
}

func (p *ProjectMapper) GetProjectsByName(name string) ([]*model.Project, error) {
	project := p.query.Project
	var po query.IProjectDo
	if name != "" {
		po = project.Where(project.Name.Like(name))
	}
	return po.Find()
}
