// Package help serves the documentation sets that ship with the application.
package serviceHelpImpl

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	domainEngine "github.com/wensboy/lightpc/backend/domain/engine"
	engineStore "github.com/wensboy/lightpc/backend/domain/engine/store"
	domainService "github.com/wensboy/lightpc/backend/domain/service"
	serviceHelp "github.com/wensboy/lightpc/backend/domain/service/help"
	"github.com/wensboy/lightpc/backend/internal/cast"
	"github.com/wensboy/lightpc/backend/internal/config"
	interError "github.com/wensboy/lightpc/backend/internal/error"
)

// HelpService reads the documentation that ships with the application.
//
// The sets are embedded, so they never change while the process runs. Init reads every meta.json once and
// keeps both the ordered list the reader shows and the version to directory lookup that Read needs; resolving
// a version on every call would decode the same file again for no benefit.
type HelpService struct {
	fileStore engineStore.FileStore
	roots     []string
	versions  []serviceHelp.DocsVersionDto
	byVersion map[string]string
}

// NewHelpService builds an uninitialised service.
func NewHelpService() *HelpService {
	return &HelpService{byVersion: make(map[string]string, 2)}
}

// Group implements domainService.Service.
func (s *HelpService) Group() string { return domainService.GroupHelp }

// Deps implements domainService.Service: the documents live in the file store.
func (s *HelpService) Deps() []string { return []string{domainEngine.StoreFs} }

// Init implements domainService.Service: it resolves the file store and indexes the documentation sets.
func (s *HelpService) Init(sCtx domainService.ServiceContext) error {
	if sCtx.AppCtx != nil {
		interError.RegisterAll(sCtx.AppCtx, serviceHelp.Errors...)
	}

	engineKey := domainEngine.EngineDep(domainEngine.GroupStore, domainEngine.NameFs)
	engine, found := sCtx.Engines[engineKey]
	if !found {
		return interError.Retrieve(engineStore.ErrKeyInvalid).Wrap(fmt.Errorf("engine %q is not ready", engineKey))
	}
	fileStore, ok := cast.As[engineStore.FileStore](engine)
	if !ok {
		return interError.Retrieve(engineStore.ErrKeyInvalid).Wrap(fmt.Errorf("engine %q is not a file store", engineKey))
	}
	s.fileStore = fileStore

	if sCtx.AppCtx == nil || sCtx.AppCtx.ConfigContext == nil {
		slog.Default().Warn("help service without configuration, no documentation will be listed")
		return nil
	}
	value, found := sCtx.AppCtx.ConfigContext.Lookup(rootsKey)
	if !found {
		slog.Default().Warn("help service found no documentation directories", "key", rootsKey)
		return nil
	}
	list, ok := cast.To[[]any](value)
	if !ok {
		slog.Default().Warn("help service found a malformed documentation list", "key", rootsKey)
		return nil
	}
	for _, entry := range list {
		if root, ok := cast.To[string](entry); ok && root != "" {
			s.roots = append(s.roots, root)
		}
	}

	s.index()
	slog.Default().Warn("help service ready", "roots", s.roots, "versions", len(s.versions))
	return nil
}

// Deinit implements domainService.Service.
func (s *HelpService) Deinit() error { return nil }

// index reads every meta.json once and keeps the result.
//
// A set whose meta.json cannot be read or decoded is skipped with a warning: one broken directory must not hide
// the others, and the reader would have nothing to show for it anyway. The entity is turned into the dto right
// here rather than through a converter, so the mapping stays visible at the point it happens.
func (s *HelpService) index() {
	s.versions = make([]serviceHelp.DocsVersionDto, 0, len(s.roots))
	s.byVersion = make(map[string]string, len(s.roots))

	for _, root := range s.roots {
		location := config.MustResource(root).Join(metaName).String()

		data, err := s.fileStore.Get(context.Background(), location)
		if err != nil {
			slog.Default().Warn("help meta unreadable", "path", location, "err", err)
			continue
		}
		parsed, err := decodeMeta(data)
		if err != nil {
			slog.Default().Warn("help meta invalid", "path", location, "err", err)
			continue
		}
		if parsed.Version == "" {
			slog.Default().Warn("help meta without a version", "path", location)
			continue
		}

		s.versions = append(s.versions, serviceHelp.DocsVersionDto{
			Version:    parsed.Version,
			Index:      parsed.Index,
			Deprecated: parsed.Deprecated,
			Tags:       parsed.Tags,
			Toc:        parsed.Toc,
		})
		s.byVersion[parsed.Version] = root
	}
}

// Payload implements serviceHelp.HelpClient: every documentation set, ordered as the configuration lists them.
func (s *HelpService) Payload() (serviceHelp.DocsPayloadDto, error) {
	if s.fileStore == nil {
		return serviceHelp.DocsPayloadDto{Versions: []serviceHelp.DocsVersionDto{}}, interError.Retrieve(engineStore.ErrKeyInvalid)
	}
	return serviceHelp.DocsPayloadDto{Versions: s.versions}, nil
}

// Read implements serviceHelp.HelpClient: the markdown of one document.
func (s *HelpService) Read(version, documentPath string) (string, error) {
	if s.fileStore == nil {
		return "", interError.Retrieve(engineStore.ErrKeyInvalid)
	}

	root, found := s.byVersion[version]
	if !found {
		return "", interError.Retrieve(serviceHelp.ErrKeyVersionUnknown)
	}
	// The path comes from the toc, but it still has to stay inside the set.
	if documentPath == "" || strings.Contains(documentPath, "..") || strings.HasPrefix(documentPath, "/") {
		return "", interError.Retrieve(serviceHelp.ErrKeyPathInvalid)
	}

	data, err := s.fileStore.Get(context.Background(), config.MustResource(root).Join(documentPath).String())
	if err != nil {
		return "", err
	}
	return string(data), nil
}
