/*
@Time : 2019/10/17 16:04
@Author : nickqnxie
@File : zh.go
*/

package i18n

const (
	ErrParam  string = "参数错误"
	ErrServer string = "服务器忙碌，请稍后再试"
	TimLayOut string = "2006-01-02 15:04:05"
)

var ZhMessage = map[string]string{
	"AddSetClusterRequest.Setname.required": "set名称不能为空",
	"DelSetClusterRequest.Setid.required":   "setid不能为空",
	"AddSetClusterRequest.Area.required":    "可用区不能为空",
	"AddSetClusterRequest.Region.required":  "地域不能为空",
	"AddSetClusterRequest.IdcType.required": "idc类型,自研/云",
}
