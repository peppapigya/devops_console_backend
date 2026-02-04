package common

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ReturnData struct {
	Status  int                    `json:"status"`
	Message string                 `json:"message"`
	Data    map[string]interface{} `json:"data"`
}

// NewReturnData 构造函数
func NewReturnData() ReturnData {
	returnData := ReturnData{}
	returnData.Status = 200
	data := make(map[string]interface{})
	returnData.Data = data
	return returnData
}

type Response struct {
	Status  int         `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func Success(c *gin.Context, data interface{}) {
	c.JSON(
		http.StatusOK,
		Response{
			Status:  200,
			Message: "success",
			Data:    data,
		},
	)
}

func Fail(c *gin.Context, code *ErrorCode) {
	c.JSON(
		http.StatusOK,
		Response{
			Status:  code.Code,
			Message: code.Msg,
			Data:    nil,
		},
	)
}

func FailWithError(c *gin.Context, err error) {
	FailWithMsg(c, err.Error())
}

func FailWithMsg(c *gin.Context, msg string) {
	c.JSON(
		http.StatusOK,
		Response{
			Status:  500,
			Message: msg,
			Data:    nil,
		},
	)
}

// PageInfoResponse 分页信息
type PageInfoResponse[T any] struct {
	// 当前页码
	PageNum int `json:"pageNum"`
	// 页面数量
	PageSize int `json:"pageSize"`
	// 数据总数
	Total int64 `json:"total"`
	// 数据
	Data []T `json:"data"`
}

// ValidateFail 参数验证失败
func ValidateFail(c *gin.Context, msg string) {
	FailWithMsg(c, msg)
}

// BusinessFail 业务失败
func BusinessFail(c *gin.Context, msg string) {
	FailWithMsg(c, msg)
}

// StrToInt64 字符串转int64
func StrToInt64(s string) int64 {
	i, _ := strconv.ParseInt(s, 10, 64)
	return i
}
