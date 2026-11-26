package serviceGreetImpl

import (
	"fmt"
	"log/slog"

	domainEngine "github.com/wensboy/lightpc/backend/domain/engine"
	engineLog "github.com/wensboy/lightpc/backend/domain/engine/log"
	engineNotice "github.com/wensboy/lightpc/backend/domain/engine/notice"
	domainService "github.com/wensboy/lightpc/backend/domain/service"
	"github.com/wensboy/lightpc/backend/internal/cast"
)

// GreetService 结构下通常只包含该服务职责范围内最小的字段数据集合, 作为一个示例, 类似 check service, 实际服务注册时可忽略.
// init 阶段从 service context 中获取必须的数据.
// deinit 阶段释放必须的资源, 通过 ebus 发送必要的事件.
type GreetService struct {
	direct  engineLog.DirectClient
	fanout  engineLog.FanoutClient
	pattern engineLog.PatternClient
	notice  engineNotice.Client
}

func NewGreetService() *GreetService {
	return &GreetService{}
}

func (s *GreetService) Group() string {
	return domainService.GroupGreet
}

// Init resolves the log engines from the service context so Greet can exercise every sink.
func (s *GreetService) Init(sCtx domainService.ServiceContext) error {
	if engine, ok := sCtx.Engines[domainEngine.GroupLog+"."+domainEngine.NameDirect]; ok {
		s.direct, _ = cast.As[engineLog.DirectClient](engine)
	}
	if engine, ok := sCtx.Engines[domainEngine.GroupLog+"."+domainEngine.NameFanout]; ok {
		s.fanout, _ = cast.As[engineLog.FanoutClient](engine)
	}
	if engine, ok := sCtx.Engines[domainEngine.GroupLog+"."+domainEngine.NamePattern]; ok {
		s.pattern, _ = cast.As[engineLog.PatternClient](engine)
	}
	if engine, ok := sCtx.Engines[domainEngine.GroupNotice+"."+domainEngine.NameNotice]; ok {
		s.notice, _ = cast.As[engineNotice.Client](engine)
	}
	return nil
}

func (s *GreetService) Deinit() error {
	return nil
}

// EmitLogs writes one record per level through the database entry.
//
// It exists so the console has something to show on a fresh start: the boundary of the log service is the moment the
// service came up, which is later than the records produced during start up, so the panel would otherwise be empty.
// Each call writes four rows, one per level, which is what makes the level menu observable.
func (s *GreetService) EmitLogs() int {
	if s.fanout == nil {
		return 0
	}
	logger, ok := s.fanout.Logger("dev.db")
	if !ok {
		return 0
	}

	written := 0
	for _, entry := range []struct {
		emit  func(msg string, args ...any)
		level string
	}{
		{logger.Debug, "debug"},
		{logger.Info, "info"},
		{logger.Warn, "warn"},
		{logger.Error, "error"},
	} {
		entry.emit("console sample record",
			"level", entry.level,
			"origin", "greet",
			"detail", "this row exists so the panel has something to show",
		)
		written++
	}

	// One notice per batch, so opening the console proves the whole way through: the engine records it, announces it
	// and the interface turns the announcement into a toast. A failure here is reported rather than swallowed, because
	// a silent notice is exactly the kind of thing this check exists to catch.
	if s.notice != nil {
		if _, err := s.notice.Notify(engineNotice.Input{
			Level:   engineNotice.LevelSuccess,
			Title:   "log sample written",
			Content: fmt.Sprintf("%d records were added to the console", written),
			Source:  "greet",
		}); err != nil {
			slog.Default().Warn("greet notice failed", "err", err)
		}
	}
	return written
}

// Greet writes one record through every log route before answering.
//
// Per invocation the expected outcome is: one console record, two records in debug.log (slog default
// and the named direct logger), one record each in process.log and error.log, two records in
// network_debug_up.log and one in network_warn_up.engineLog.
func (s *GreetService) Greet(name string) string {
	// Direct: the slog default resolves to the first console handler of the direct group.
	slog.Info("greet via slog default", "name", name)

	if s.direct != nil {
		if logger, ok := s.direct.Logger("dev.file"); ok {
			logger.Info("greet via direct", "name", name)
		}
	}

	if s.fanout != nil {
		if logger, ok := s.fanout.Logger("dev.test"); ok {
			logger.Info("greet via fanout", "name", name)
		}
	}

	if s.pattern != nil {
		// 只向 debug 文件写 log
		if logger, ok := s.pattern.Logger("network.debug"); ok {
			logger.Info("greet via pattern debug", "name", name)
		}
		if logger, ok := s.pattern.Logger("network.*"); ok {
			logger.Info("greet via pattern", "name", name)
		}
	}

	return "hello " + name
}
