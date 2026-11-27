package serviceNotice

// NoticeClient is the surface the front end depends on.
type NoticeClient interface {
	// History returns the live notices, newest first, capped by the configured count.
	History() (NoticeHistoryDto, error)
	// Send raises a notice through the engine, so both sides of the application share one pipeline.
	Send(input NoticeSendInputDto) (NoticeItemDto, error)
	// Clear hides everything currently shown, and anything older, for the rest of the process.
	Clear() (int, error)
	// Remove drops the notices named by the request and reports how many were there.
	Remove(input NoticeRemoveInputDto) (int, error)
}
