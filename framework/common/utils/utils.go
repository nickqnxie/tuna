/*
@Time : 2019/10/23 15:23
@Author : nickqnxie
@File : utils.go
*/

package utils

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ddliu/go-httpclient"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net"
	"regexp"
	"strings"
	"time"
)

func GetIPs() (ips []string) {

	interfaceAddr, err := net.InterfaceAddrs()
	if err != nil {
		logrus.Errorf("fail to get net interface addrs: %v", err)
		return ips
	}

	for _, address := range interfaceAddr {
		ipNet, isValidIpNet := address.(*net.IPNet)
		if isValidIpNet && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil {
				ips = append(ips, ipNet.IP.String())
			}
		}
	}
	return ips
}

func GetConnLocalIP(dstAddress string) (string, error) {
	conn, err := net.Dial("tcp", dstAddress)
	if err != nil {
		return "127.0.0.1", err
	}
	defer conn.Close()
	localIP := strings.Split(conn.LocalAddr().String(), ":")[0]

	return localIP, nil
}

//获取前一分钟时间戳
func OneMinuteBefore() int64 {
	m, _ := time.ParseDuration("-1m")
	return time.Now().Add(m).Unix()
}

func Str2Time(formatTimeStr string) *time.Time {
	loc, _ := time.LoadLocation("Local")
	theTime, _ := time.ParseInLocation(time.RFC3339, formatTimeStr, loc) //使用模板在对应时区转化为time.time类型

	return &theTime

}

func BuildData(data interface{}) (res map[string]interface{}, err error) {

	bytes, err := json.Marshal(data)

	if err != nil {
		return
	}

	err = json.Unmarshal(bytes, &res)

	return
}

func GetRegionByip(ip string) (rip string, err error) {
	type apiresp struct {
		ErrorCode  int    `json:"errorCode"`
		ResultData string `json:"resultData"`
		ErrorInfo  string `json:"errorInfo"`
		Version    string `json:"version"`
	}

	var resp apiresp
	apiurl := "http://platformserver.cpo.tencentyun.com:8080/api/v1/platform_flow/area/query/ip"

	response, err := httpclient.PostJson(apiurl, map[string]interface{}{
		"ip": ip,
	})

	defer response.Body.Close()
	if err != nil {
		logrus.Warnf("GetRegionByip is error, err=%v", err)
		return
	}

	bytes, err := ioutil.ReadAll(response.Body)
	err = json.Unmarshal(bytes, &resp)

	if err != nil {
		logrus.Warnf("GetRegionByip is error, err=%v", err)
		return
	}

	rip = resp.ResultData
	return
}

func GetMd5(str string) string {
	h := md5.New()
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}

func GetEthxIp(ethx string) (err error, ip string) {
	filePath := fmt.Sprintf("/etc/sysconfig/network-scripts/ifcfg-%s", ethx)

	bytes, err := ioutil.ReadFile(filePath)

	if err != nil {
		return
	}

	reg := regexp.MustCompile(`IPADDR=(\d+\.\d+\.\d+\.\d+)`)

	allString := reg.FindAllString(string(bytes), -1)

	if len(allString) > 0 {
		ip = strings.Split(allString[0], "=")[1]
	} else {

		return errors.New("No match to IP"), ""
	}

	return

}
