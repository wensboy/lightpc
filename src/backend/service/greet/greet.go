package greet

import (
	domainService "github.com/wensboy/lightpc/backend/domain/service"
)

// GreetService 结构下通常只包含该服务职责范围内最小的字段数据集合, 作为一个示例, 类似 check service, 实际服务注册时可忽略.
// init 阶段从 service context 中获取必须的数据.
// deinit 阶段释放必须的资源, 通过 ebus 发送必要的事件.
type GreetService struct {
}

func NewGreetService() *GreetService {
	return &GreetService{}
}

func (s *GreetService) Group() string {
	return domainService.GroupGreet
}

func (s *GreetService) Init(sCtx domainService.ServiceContext) error {
	return nil
}

func (s *GreetService) Deinit() error {
	return nil
}

func (s *GreetService) Greet(name string) string {
	return "hello " + name
}
