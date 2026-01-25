package mapper

import (
	"devops-console-backend/internal/dal/model"
	"devops-console-backend/internal/dal/query"

	"gorm.io/gorm"
)

type UserMapper struct {
	DB    *gorm.DB
	query *query.Query
}

func NewUserMapper(db *gorm.DB) *UserMapper {
	return &UserMapper{
		DB:    db,
		query: query.Use(db),
	}
}

func (user *UserMapper) GetUserByUsername(username string) (*model.SystemUser, error) {
	userQuery := user.query.SystemUser
	return userQuery.Where(userQuery.Username.Eq(username)).First()
}

// ===================  systemUserToken ===================

func (user *UserMapper) InsertSystemUserToken(userToken *model.SystemUserToken) error {
	return user.query.SystemUserToken.Create(userToken)
}
