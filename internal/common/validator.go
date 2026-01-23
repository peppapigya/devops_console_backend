package common

import (
	"reflect"
	"strings"
	"sync"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	zhtranslations "github.com/go-playground/validator/v10/translations/zh"
)

type DefaultValidator struct {
	once     sync.Once
	validate *validator.Validate
	trans    ut.Translator
}

var _ binding.StructValidator = &DefaultValidator{}

func (v *DefaultValidator) ValidateStruct(obj interface{}) error {

	if kindOfData(obj) == reflect.Struct {

		v.lazyInit()

		if err := v.validate.Struct(obj); err != nil {
			return err
		}
	}

	return nil
}

func (v *DefaultValidator) Engine() interface{} {
	v.lazyInit()
	return v.validate
}

// 初始化
func (v *DefaultValidator) lazyInit() {
	v.once.Do(func() {
		v.validate = validator.New()
		v.validate.SetTagName("validate")
		// 如果没有validate标签的时候，使用gin默认的binding
		v.validate.RegisterTagNameFunc(registerTagNameFunc)
		// 集成中文翻译
		zhCN := zh.New()
		uni := ut.New(zhCN, zhCN)
		v.trans, _ = uni.GetTranslator("zh")
		_ = zhtranslations.RegisterDefaultTranslations(v.validate, v.trans)
	})
}

func registerTagNameFunc(fld reflect.StructField) string {
	// 优先使用label
	if label := fld.Tag.Get("label"); label != "" {
		return label
	}
	// 使用comment
	if comment := fld.Tag.Get("comment"); comment != "" {
		return comment
	}

	// 如果使用的是gorm
	gormTag := fld.Tag.Get("gorm")
	if gormTag != "" {
		if comment := extractCommentFromGormTag(gormTag); comment != "" {
			return comment
		}
	}
	// 使用json
	if jsonTag := fld.Tag.Get("json"); jsonTag != "" {
		return jsonTag
	}

	return fld.Name
}

func extractCommentFromGormTag(gormTag string) string {
	parts := strings.Split(gormTag, ";")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(part, "comment:") {
			return strings.TrimPrefix(part, "comment:")
		}
	}
	return ""
}

func kindOfData(data interface{}) reflect.Kind {

	value := reflect.ValueOf(data)
	valueType := value.Kind()

	if valueType == reflect.Ptr {
		valueType = value.Elem().Kind()
	}
	return valueType
}

func GetValidator() (*validator.Validate, bool) {
	if defaultValidator, ok := binding.Validator.(*DefaultValidator); ok {
		engine := defaultValidator.Engine()
		if validate, ok := engine.(*validator.Validate); ok {
			return validate, true
		}
	}
	return nil, false
}
func (v *DefaultValidator) GetTrans() ut.Translator {
	v.lazyInit()
	return v.trans
}
func GetTranslator() ut.Translator {
	if defaultValidator, ok := binding.Validator.(*DefaultValidator); ok {
		return defaultValidator.GetTrans()
	}
	return nil
}
