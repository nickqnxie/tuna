/*
@Time : 2019/10/22 19:35
@Author : nickqnxie
@File : lb.go
*/

package common

import (
	"math/rand"
	"sync/atomic"
)

//负载均衡器接口
type Balancer interface {
	Next(hosts []*ServerInstance) *ServerInstance
}

var _ Balancer = new(RoundRobinBalancer)

//简单轮询负载均衡算法
type RoundRobinBalancer struct {
	ct uint32 //计数器
}

func (r *RoundRobinBalancer) Next(hosts []*ServerInstance) *ServerInstance {
	if len(hosts) == 0 {
		return nil
	}
	//自增
	count := atomic.AddUint32(&r.ct, 1)
	//取模计算索引
	index := int(count) % len(hosts)
	//按照索引取出实例
	instance := hosts[index]
	return instance
}

var _ Balancer = new(RandomBalancer)

//随机负载均衡算法
type RandomBalancer struct {
}

func (r *RandomBalancer) Next(hosts []*ServerInstance) *ServerInstance {
	if len(hosts) == 0 {
		return nil
	}
	//随机数
	count := rand.Int31()
	//取模计算索引
	index := int(count) % len(hosts)
	//按照索引取出实例
	instance := hosts[index]
	return instance
}

/*
	使用例子：
	//consulclinet, err := common.NewConsulClinet()
	//
	//if err != nil {
	//	log.Fatal("consul连接失败,", err)
	//}
	//获取所以可以实例
	//nodes, _ := consulclinet.GetServerInstance("cloudat.micro.namespace.srv.heartbeat")
	//
	//bytes, _ := json.Marshal(nodes)
	//创建负载均衡器
	//lb := common.RandomBalancer{}
	//
	//t := time.NewTimer(time.Second * 2)
	//defer t.Stop()
	//for {
	//	<-t.C
	//	next := lb.Next(nodes)
	//
	//	fmt.Printf("%s:%d\n", next.Address, next.Port)
	//
	//	t.Reset(time.Second * 2)
	//}
*/
