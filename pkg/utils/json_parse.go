package utils

import (
	"devops-console-backend/internal/common"
	"errors"
	"log"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

// BindAndValidate 解析参数并将参数绑定到obj
func BindAndValidate(c *gin.Context, obj interface{}) bool {
	contentType := c.GetHeader("Content-Type")
	if strings.Contains(contentType, "multipart/form-data") {
		if err := c.Request.ParseMultipartForm(32 << 20); err != nil { // 32MB
			log.Printf("解析multipart表单失败: %v", err)
			common.Fail(c, common.NewErrorCode(400, "表单解析失败"))
			return false
		}

		// 使用 ShouldBindWith 指定 FormData 绑定
		if err := c.ShouldBindWith(obj, binding.Form); err != nil {
			log.Printf("FormData参数绑定失败: %v", err)
			common.Fail(c, common.ServerError)
			return false
		}
	} else {
		if err := c.ShouldBindJSON(obj); err != nil {
			var errs validator.ValidationErrors
			if ok := errors.As(err, &errs); ok {
				log.Printf("参数校验失败: %v", errs)
				common.Fail(c, common.NewErrorCode(400, errs[0].Translate(common.GetTranslator())))
				c.Abort()
				return false
			}
			log.Printf("解析参数失败: %v", err)
			common.Fail(c, common.ServerError)
			return false
		}
	}

	return true
}

func BindQueryParam(c *gin.Context, obj interface{}) bool {
	if err := c.ShouldBindQuery(obj); err != nil {
		var errs validator.ValidationErrors
		if ok := errors.As(err, &errs); ok {
			log.Printf("参数校验失败: %v", errs)
			common.FailWithMsg(c, errs[0].Translate(trans))
			return false
		}
		log.Printf("解析参数失败: %v", err)
		common.Fail(c, common.ServerError)
		return false
	}
	return true
}

// GetParam 获取路径参数以及参数校验
func GetParam(c *gin.Context, key string, param interface{}, validate func(param interface{})) {
	var value string
	value = c.Query(key)
	if value == "" {
		value = c.Param(key)
	}
	if strParam, ok := param.(*string); ok {
		*strParam = value
	}
	if int64Param, ok := param.(*int64); ok {
		*int64Param, _ = strconv.ParseInt(value, 10, 64)
	}

	if validate != nil {
		validate(param)
	}
	return
}
