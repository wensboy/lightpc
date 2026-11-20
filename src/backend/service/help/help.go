// Package help serves the documentation sets that ship with the application.
package serviceHelpImpl

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	domainEngine "github.com/wensboy/lightpc/backend/domain/engine"
	engineStore "github.com/wensboy/lightpc/backend/domain/engine/store"
	domainService "github.com/wensboy/lightpc/backend/domain/service"
	serviceHelp "github.com/wensboy/lightpc/backend/domain/service/help"
	"github.com/wensboy/lightpc/backend/internal/cast"
	interError "github.com/wensboy/lightpc/backend/internal/error"
	"github.com/wensboy/lightpc/backend/internal/path"
)

// helpKey names the configuration entry that lists the documentation directories.
const helpKey = "default.runtime.docs"

// metaName is the file every documentation directory carries.
const metaName = "meta.json"

// Service reads the documentation that ships with the application.
type Service struct {
	fileStore engineStore.FileStore
	roots     []string
}

// NewService builds an uninitialised service.
func NewService() *Service {
	return &Service{}
}

// Group implements domainService.Service.
func (s *Service) Group() string { return domainService.GroupHelp }

// Init implements domainService.Service: it resolves the file store and the documentation directories.
func (s *Service) Init(sCtx domainService.ServiceContext) error {
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
	value, found := sCtx.AppCtx.ConfigContext.Lookup(helpKey)
	if !found {
		slog.Default().Warn("help service found no documentation directories", "key", helpKey)
		return nil
	}
	list, ok := cast.To[[]any](value)
	if !ok {
		slog.Default().Warn("help service found a malformed documentation list", "key", helpKey)
		return nil
	}
	for _, entry := range list {
		if root, ok := cast.To[string](entry); ok && root != "" {
			s.roots = append(s.roots, root)
		}
	}
	slog.Default().Warn("help service ready", "roots", s.roots)
	return nil
}

// Deinit implements domainService.Service.
func (s *Service) Deinit() error { return nil }

// Payload implements serviceHelp.HelpClient: every documentation set, ordered by version.
func (s *Service) Payload() (serviceHelp.DocsPayload, error) {
	payload := serviceHelp.DocsPayload{Versions: make([]serviceHelp.DocsVersion, 0, len(s.roots))}
	if s.fileStore == nil {
		return payload, interError.Retrieve(engineStore.ErrKeyInvalid)
	}

	for _, root := range s.roots {
		data, err := s.fileStore.Get(context.Background(), path.Join(root, metaName))
		if err != nil {
			slog.Default().Warn("help meta unreadable", "path", path.Join(root, metaName), "err", err)
			continue
		}
		parsed, err := decodeMeta(data)
		if err != nil {
			slog.Default().Warn("help meta invalid", "path", path.Join(root, metaName), "err", err)
			continue
		}
		payload.Versions = append(payload.Versions, serviceHelp.DocsVersion{
			Version:    parsed.Version,
			Index:      parsed.Index,
			Deprecated: parsed.Deprecated,
			Tags:       parsed.Tags,
			Toc:        parsed.Toc,
		})
	}
	return payload, nil
}

// Read implements serviceHelp.HelpClient: the markdown of one document.
func (s *Service) Read(version, documentPath string) (string, error) {
	if s.fileStore == nil {
		return "", interError.Retrieve(engineStore.ErrKeyInvalid)
	}
	root, found := s.rootOf(version)
	if !found {
		return "", interError.Retrieve(serviceHelp.ErrKeyVersionUnknown)
	}
	// The path comes from the toc, but it still has to stay inside the set.
	if strings.Contains(documentPath, "..") || strings.HasPrefix(documentPath, "/") {
		return "", interError.Retrieve(serviceHelp.ErrKeyPathInvalid)
	}
	data, err := s.fileStore.Get(context.Background(), path.Join(root, documentPath))
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// rootOf returns the directory that carries the given version.
func (s *Service) rootOf(version string) (string, bool) {
	for _, root := range s.roots {
		data, err := s.fileStore.Get(context.Background(), path.Join(root, metaName))
		if err != nil {
			continue
		}
		parsed, err := decodeMeta(data)
		if err != nil {
			continue
		}
		if parsed.Version == version {
			return root, true
		}
	}
	return "", false
}

// meta is the shape of one meta.json, with the toc kept in declaration order.
type meta struct {
	Version    string    `json:"version"`
	Index      string    `json:"index"`
	Deprecated bool      `json:"deprecated"`
	Tags       []string  `json:"tags"`
	Toc        tocGroups `json:"toc"`
}

// decodeMeta reads one meta.json.
func decodeMeta(data []byte) (meta, error) {
	parsed := meta{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		return meta{}, fmt.Errorf("decode %s: %w", metaName, err)
	}
	return parsed, nil
}
