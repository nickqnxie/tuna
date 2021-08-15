/*
@Time : 2019/10/21 20:37
@Author : nickqnxie
@File : elastic.go
*/

package common

import (
	"context"
	"encoding/json"
	"github.com/sirupsen/logrus"
	"gopkg.in/olivere/elastic.v6"
)

type BasicAuth struct {
	UserName string
	Passwrod string
}

func NewElasticClient(url string, basicAuth ...BasicAuth) (*elastic.Client, error) {

	client, err := elastic.NewClient(elastic.SetURL(url),
		elastic.SetBasicAuth("elastic", "CloudCat@5566$*"),
		elastic.SetSniff(false),
		elastic.SetHealthcheck(false))

	if len(basicAuth) > 0 {
		client, err = elastic.NewClient(elastic.SetURL(url),
			elastic.SetBasicAuth(basicAuth[0].UserName, basicAuth[0].Passwrod),
			elastic.SetSniff(false),
			elastic.SetHealthcheck(false))
	}

	if err != nil {
		return nil, err
	}

	return client, nil
}

func NewElasticClientv2(url, username, password string) (*elastic.Client, error) {

	client, err := elastic.NewClient(elastic.SetURL(url),
		elastic.SetBasicAuth(username, password),
		elastic.SetSniff(false),
		elastic.SetHealthcheck(false))
	if err != nil {
		return nil, err
	}
	return client, nil
}

func IndexMappings(configMappings map[string]interface{}) string {
	// 加入all_search字段
	//configMappings["all_search"] = map[string]interface{}{
	//	"type": "text"}
	configMappingsJSON, _ := json.Marshal(map[string]interface{}{"mappings": map[string]interface{}{"_doc": map[string]interface{}{"properties": configMappings}}})
	return string(configMappingsJSON)
}

func IndexExists(client *elastic.Client, index string, mapping map[string]interface{}) {
	exists, err := client.IndexExists(index).Do(context.Background())

	if err != nil {
		logrus.Errorf("%v", err)
		return
	}

	if !exists {
		if mapping != nil {
			_, err = client.CreateIndex(index).BodyString(IndexMappings(mapping)).Do(context.Background())
		} else {
			_, err = client.CreateIndex(index).Do(context.Background())
		}

		if err != nil {
			logrus.Errorf("index创建失败,err=%v", err)
		}
		_, err = client.IndexPutSettings(index).
			BodyJson(map[string]interface{}{"refresh_interval": "100ms"}).
			Do(context.Background())

		if err != nil {
			logrus.Errorf("index刷新时间设置失败,err=%v", err)
		}
	}

}
