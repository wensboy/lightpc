package serviceLogImpl

import (
	"fmt"
	"log/slog"
	"sync"
	"time"

	domainEngine "github.com/wensboy/lightpc/backend/domain/engine"
	engineStore "github.com/wensboy/lightpc/backend/domain/engine/store"
	domainService "github.com/wensboy/lightpc/backend/domain/service"
	serviceLog "github.com/wensboy/lightpc/backend/domain/service/log"
	"github.com/wensboy/lightpc/backend/internal/cast"
	"github.com/wensboy/lightpc/backend/internal/config"
	interError "github.com/wensboy/lightpc/backend/internal/error"
)

var (
	_ domainService.Service = (*LogService)(nil)
	_ serviceLog.LogClient  = (*LogService)(nil)
)

// LogService serves the console log view: it reads the records of one source from that source's boundary onwards,
// runs them through the filter pipeline, and hands the result over.
//
// The service owns three things the interface never sees: the boundary of each source, the filters, and the reader
// that turns a boundary into rows. None of them is exposed, so the interface only sends the source it picked and the
// parameters its toolbar holds.
type LogService struct {
	mu    sync.RWMutex
	store engineStore.DBStore
	// boundaries hold the instant each source is read from, keyed by name. A source appears here the first time it
	// is asked about, so the table stays the registry and nothing has to be declared twice.
	boundaries map[string]int64
	// filters run in order; each one narrows the list the previous step produced.
	filters []serviceLog.Filter
	// offsets convert a source specific instant onto the millisecond clock the table stores.
	now   func() int64
	limit int

	state domainEngine.EngineState
}

// NewLogService builds the service. Init resolves the database and registers the sources.
func NewLogService() *LogService {
	return &LogService{
		boundaries: map[string]int64{},
		state:      domainEngine.EnginePending,
		now:        func() int64 { return time.Now().UnixMilli() },
		limit:      defaultLimit,
	}
}

// Group implements domainService.Service.
func (s *LogService) Group() string { return domainService.GroupLog }

// Deps implements domainService.Service.
func (s *LogService) Deps() []string { return []string{domainEngine.StoreDb} }

// State implements domainService.Service.
func (s *LogService) State() domainEngine.EngineState {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state
}

// Init implements domainService.Service.
//
// The boundary of every source starts at the moment the service comes up. Nothing is read from before that instant,
// which is what makes a restart show only the logs of the current process while the table keeps the full history.
func (s *LogService) Init(sCtx domainService.ServiceContext) error {
	if sCtx.AppCtx != nil {
		interError.RegisterAll(sCtx.AppCtx.ErrorContext, serviceLog.Errors...)
	}

	engineKey := domainEngine.EngineDep(domainEngine.GroupStore, domainEngine.NameDb)
	engine, found := sCtx.Engines[engineKey]
	if !found {
		return interError.Retrieve(serviceLog.ErrKeyDatabaseMissing).
			Wrap(fmt.Errorf("engine %q is not ready", engineKey))
	}

	registry, ok := cast.As[engineStore.DBRegistry](engine)
	if !ok {
		return interError.Retrieve(serviceLog.ErrKeyDatabaseMissing).
			Wrap(fmt.Errorf("engine %q is not a database registry", engineKey))
	}

	key := ""
	expiration := defaultLimit
	if sCtx.AppCtx != nil && sCtx.AppCtx.ConfigContext != nil {
		cfg := sCtx.AppCtx.ConfigContext
		if value, found := cfg.Lookup(databaseKey); found {
			key, _ = cast.To[string](value)
		}
		if value, found := cfg.Lookup(limitKey); found {
			if text, _ := cast.To[string](value); text != "" {
				if parsed, err := config.ParseCount(text); err == nil {
					expiration = parsed
				}
			}
		}
	}
	if key == "" {
		keys := registry.DBKeys()
		if len(keys) == 0 {
			return interError.Retrieve(serviceLog.ErrKeyDatabaseMissing)
		}
		key = keys[0]
	}

	store, found := registry.DB(key)
	if !found {
		return interError.Retrieve(serviceLog.ErrKeyDatabaseMissing).
			Wrap(fmt.Errorf("database %q is not registered", key))
	}

	start := s.now()

	s.mu.Lock()
	s.store = store
	s.limit = expiration
	s.boundaries = map[string]int64{}
	s.filters = []serviceLog.Filter{&levelFilter{}, &patternFilter{}}
	s.mu.Unlock()

	// Every source that already wrote a row starts at the same instant: a restart shows only what this process
	// produces, while the table keeps the older rows for a reader that asks for a wider range.
	if names, err := s.listeners(store); err == nil {
		s.mu.Lock()
		for _, name := range names {
			s.boundaries[name] = start
		}
		s.mu.Unlock()
	}

	s.state = domainEngine.EngineReady
	slog.Default().Warn("log service ready", "database", key, "limit", s.limit)
	return nil
}

// Start implements domainService.Service.
func (s *LogService) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state = domainEngine.EngineRunning
	return nil
}

// Stop implements domainService.Service.
func (s *LogService) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state = domainEngine.EngineStopped
	return nil
}

// Deinit implements domainService.Service.
func (s *LogService) Deinit() error { return s.Stop() }

// Listeners implements serviceLog.LogClient.
//
// The table is the registry, so the answer changes as sources start writing. The title is the name: a source is
// identified by what it writes, and there is no second label to keep in step with it.
func (s *LogService) Listeners() ([]serviceLog.ListenerDto, error) {
	s.mu.RLock()
	store := s.store
	s.mu.RUnlock()

	if store == nil {
		return nil, interError.Retrieve(serviceLog.ErrKeyDatabaseMissing)
	}
	names, err := s.listeners(store)
	if err != nil {
		return nil, err
	}

	entries := make([]serviceLog.ListenerDto, 0, len(names))
	for _, name := range names {
		entries = append(entries, serviceLog.ListenerDto{Name: name, Title: name})
	}
	return entries, nil
}

// Records implements serviceLog.LogClient.
//
// The order is: read from the listener, then run the pipeline. Reading first keeps the filters free of any knowledge
// about where the records came from, and it lets a filtered read still be a single capped query.
func (s *LogService) Records(query serviceLog.LogQueryDto) (serviceLog.LogRecordsDto, error) {
	s.mu.RLock()
	store := s.store
	limit := s.limit
	filters := append([]serviceLog.Filter(nil), s.filters...)
	s.mu.RUnlock()

	if store == nil {
		return serviceLog.LogRecordsDto{}, interError.Retrieve(serviceLog.ErrKeyDatabaseMissing)
	}
	if query.Listener == "" {
		return serviceLog.LogRecordsDto{}, interError.Retrieve(serviceLog.ErrKeyListenerUnknown)
	}
	if query.Limit > 0 {
		limit = query.Limit
	}

	boundary := s.boundaryOf(query.Listener)
	records, err := s.latest(store, query.Listener, boundary, limit)
	if err != nil {
		return serviceLog.LogRecordsDto{}, err
	}

	// Each step returns the slice it was given when it lets everything through, so an untouched filter costs one
	// nil check rather than a copy.
	for _, filter := range filters {
		records = filter.Apply(records)
	}

	if records == nil {
		records = []serviceLog.LogRecordDto{}
	}
	return serviceLog.LogRecordsDto{Records: records, Boundary: boundary}, nil
}

// Clear implements serviceLog.LogClient.
//
// Only the boundary of the named source moves. The rows stay in the table, so a reader asking for a wider range still
// finds them; what changes is where the console starts looking.
func (s *LogService) Clear(listener string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if listener == "" {
		return interError.Retrieve(serviceLog.ErrKeyListenerUnknown)
	}
	s.boundaries[listener] = s.now()
	return nil
}

// SetFilter implements serviceLog.LogClient.
//
// The payload is handed to every filter as is: each one reads the part it owns and ignores the rest, which is what
// keeps one filter invisible to the others. A filter that changes nothing is still called, so a parameter reset
// takes effect on the next read without a second entry point.
func (s *LogService) SetFilter(params serviceLog.FilterParamsDto) error {
	s.mu.RLock()
	filters := append([]serviceLog.Filter(nil), s.filters...)
	s.mu.RUnlock()

	for _, filter := range filters {
		switch target := filter.(type) {
		case interface{ Set([]int64) }:
			target.Set(params.Levels)
		case interface{ Set(string) }:
			target.Set(params.Pattern)
		}
	}
	return nil
}
