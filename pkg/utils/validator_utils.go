package utils

import (
	"devops-console-backend/internal/common"

	zh "github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	zhtranslations "github.com/go-playground/validator/v10/translations/zh"
)

var trans ut.Translator

// 启用中文翻译器

func InitValidator() {
	if v, ok := common.GetValidator(); ok {
		zhCn := zh.New()
		uni := ut.New(zhCn, zhCn)
		trans, _ = uni.GetTranslator("zh")
		_ = zhtranslations.RegisterDefaultTranslations(v, trans)
	}
}

func GetTrans() ut.Translator {
	return trans
}
