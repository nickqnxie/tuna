/*
@Time : 2019/6/13 16:51
@Author : nickqnxie
@File : res.go
*/

package baseclass

type ResCode int

const (
	ResCodeOk                    ResCode = 0
	ResCodeValidationError       ResCode = 2000
	ResCodeRequestParamsError    ResCode = 2100
	ResCodeInnerServerError      ResCode = 5000
	ResCodeDataInsertError       ResCode = 5001
	ResCodeDataQueryError        ResCode = 5002
	ResCodeDataUpdateError       ResCode = 5003
	ResCodeDataDeleteError       ResCode = 5004
	ResCodeBizError              ResCode = 6000
	ResCodeInterfaceRequestError ResCode = 6001
	ResCodePermissionError       ResCode = 6002
)

type Code struct {
	Val int
	Msg string
}

type Res struct {
	Code    ResCode     `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}
