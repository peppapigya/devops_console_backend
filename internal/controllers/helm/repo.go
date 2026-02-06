package helm

import (
	"devops-console-backend/internal/dal"
	"devops-console-backend/internal/dal/request/helm"
	helmService "devops-console-backend/internal/services/helm"
	"devops-console-backend/pkg/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RepoController Helm仓库控制器
type RepoController struct {
	db          *gorm.DB
	repoService *helmService.RepoService
}

// NewRepoController 创建仓库控制器实例
func NewRepoController(db *gorm.DB) *RepoController {
	return &RepoController{
		db:          db,
		repoService: helmService.NewRepoService(db),
	}
}

// GetRepoList 获取仓库列表
func (c *RepoController) GetRepoList(ctx *gin.Context) {
	helper := utils.NewResponseHelper(ctx)

	var repos []dal.HelmRepo
	if err := c.db.Order("created_at DESC").Find(&repos).Error; err != nil {
		helper.DatabaseError("获取仓库列表失败")
		return
	}

	// 转换为响应格式
	repoList := make([]helm.RepoListItem, 0, len(repos))
	for _, repo := range repos {
		repoList = append(repoList, helm.RepoListItem{
			ID:        repo.ID,
			Name:      repo.Name,
			URL:       repo.URL,
			Username:  repo.Username,
			CreatedAt: repo.CreatedAt.Unix(),
			UpdatedAt: repo.UpdatedAt.Unix(),
		})
	}

	helper.SuccessWithData("获取成功", "repoList", repoList)
}

// CreateRepo 创建仓库
func (c *RepoController) CreateRepo(ctx *gin.Context) {
	helper := utils.NewResponseHelper(ctx)

	var req helm.RepoCreateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		helper.ValidationError(err.Error())
		return
	}

	// TODO: 加密密码
	repo := dal.HelmRepo{
		Name:     req.Name,
		URL:      req.URL,
		Username: req.Username,
		Password: req.Password,
	}

	if err := c.db.Create(&repo).Error; err != nil {
		helper.DatabaseError("创建仓库失败")
		return
	}

	helper.SuccessWithData("创建成功", "repo", repo)
}

// UpdateRepo 更新仓库
func (c *RepoController) UpdateRepo(ctx *gin.Context) {
	helper := utils.NewResponseHelper(ctx)

	var id uint
	utils.GetParam(ctx, "id", &id, nil)

	var req helm.RepoUpdateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		helper.ValidationError(err.Error())
		return
	}

	updates := make(map[string]interface{})
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.URL != "" {
		updates["url"] = req.URL
	}
	if req.Username != "" {
		updates["username"] = req.Username
	}
	if req.Password != "" {
		// TODO: 加密密码
		updates["password"] = req.Password
	}

	if err := c.db.Model(&dal.HelmRepo{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		helper.DatabaseError("更新仓库失败")
		return
	}

	helper.Success("更新成功")
}

// DeleteRepo 删除仓库
func (c *RepoController) DeleteRepo(ctx *gin.Context) {
	helper := utils.NewResponseHelper(ctx)

	var id uint
	utils.GetParam(ctx, "id", &id, nil)

	// 删除仓库及关联的 Charts
	err := c.db.Transaction(func(tx *gorm.DB) error {
		// 删除关联 Charts
		if err := tx.Where("repo_id = ?", id).Delete(&dal.HelmChart{}).Error; err != nil {
			return err
		}
		// 删除仓库
		if err := tx.Delete(&dal.HelmRepo{}, id).Error; err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		helper.DatabaseError("删除仓库失败")
		return
	}

	helper.Success("删除成功")
}

// SyncRepo 同步仓库
func (c *RepoController) SyncRepo(ctx *gin.Context) {
	helper := utils.NewResponseHelper(ctx)

	var id int
	utils.GetParam(ctx, "id", &id, nil)

	if err := c.repoService.SyncRepo(uint(id)); err != nil {
		helper.InternalError("同步仓库失败: " + err.Error())
		return
	}

	helper.Success("同步成功")
}
