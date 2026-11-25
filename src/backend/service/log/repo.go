package serviceLogImpl

import (
	"context"

	engineStore "github.com/wensboy/lightpc/backend/domain/engine/store"
	serviceLog "github.com/wensboy/lightpc/backend/domain/service/log"
	interError "github.com/wensboy/lightpc/backend/internal/error"
)

// latest returns the newest records of one listener from since onwards.
//
// The query is written newest first because that is the direction a cap applies to: the reader wants the most recent
// rows, and asking the index for them lets sqlite stop after limit rows instead of collecting the whole range. The
// caller reverses the slice, so the order that reaches the interface is still oldest first.
//
// No ORDER BY on the result would be cheaper still, but the contract says a listener answers in time order and an
// index scan is only sorted by accident of the query plan.
func (s *LogService) latest(store engineStore.DBStore, listener string, since int64, limit int) ([]serviceLog.LogRecordDto, error) {
	if limit <= 0 {
		limit = defaultLimit
	}

	response, err := store.Dispatch(context.Background(), engineStore.QueryCommand{
		SQL: `SELECT id, listener, level, msg, detail, created_at FROM log
			WHERE listener = ? AND created_at > ?
			ORDER BY created_at DESC, id DESC
			LIMIT ?`,
		Args: []any{listener, since, limit},
	})
	if err != nil {
		return nil, err
	}
	query, ok := response.(engineStore.QueryResponse)
	if !ok || query.Rows == nil {
		return nil, interError.Retrieve(serviceLog.ErrKeyDatabaseMissing)
	}
	defer func() { _ = query.Rows.Close() }()

	records := make([]serviceLog.LogRecordDto, 0, limit)
	for query.Rows.Next() {
		var record serviceLog.LogRecordDto
		if err := query.Rows.Scan(&record.ID, &record.Listener, &record.Level, &record.Msg, &record.Detail, &record.CreatedAt); err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	if err := query.Rows.Err(); err != nil {
		return nil, err
	}

	// The rows arrive newest first; the interface renders top to bottom, so the slice is flipped once at the end
	// rather than sorted.
	for left, right := 0, len(records)-1; left < right; left, right = left+1, right-1 {
		records[left], records[right] = records[right], records[left]
	}
	return records, nil
}

// listeners returns every source that has written at least one row, ordered by name.
//
// The registry is the table itself: a source exists because it produced a record, so adding one is a matter of
// logging through it and nothing has to be declared twice.
func (s *LogService) listeners(store engineStore.DBStore) ([]string, error) {
	response, err := store.Dispatch(context.Background(), engineStore.QueryCommand{
		SQL: `SELECT DISTINCT listener FROM log ORDER BY listener`,
	})
	if err != nil {
		return nil, err
	}
	query, ok := response.(engineStore.QueryResponse)
	if !ok || query.Rows == nil {
		return nil, nil
	}
	defer func() { _ = query.Rows.Close() }()

	names := make([]string, 0)
	for query.Rows.Next() {
		var name string
		if err := query.Rows.Scan(&name); err != nil {
			return nil, err
		}
		names = append(names, name)
	}
	return names, query.Rows.Err()
}

// boundaryOf returns the instant a source is read from, creating it on first sight.
//
// A source discovered after start up begins at the moment it is first asked about, which is the same rule the service
// applies at Init: nothing before the service knew about it is shown.
func (s *LogService) boundaryOf(listener string) int64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	boundary, found := s.boundaries[listener]
	if !found {
		boundary = s.now()
		s.boundaries[listener] = boundary
	}
	return boundary
}
