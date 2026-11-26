package serviceNoticeImpl

// The settings key that caps how many notices the sheet shows.
//
// It is read through the settings engine rather than from the document directly: the count is a user preference, so
// it may be overridden by any layer, and only the settings engine knows the resolved view.
const historyKey = "general.notification.history"

// defaultHistory is what the service shows when the preference is absent or unusable.
const defaultHistory = 32

// intValue narrows a resolved setting onto an int, reporting false when the value is of another shape.
//
// The settings engine hands back whatever the layer held, which for a number is a float64 once a document has been
// decoded. Both forms are accepted, since either one means the same thing to a reader.
func intValue(value any) (int, bool) {
	switch typed := value.(type) {
	case int:
		return typed, true
	case int64:
		return int(typed), true
	case float64:
		return int(typed), true
	default:
		return 0, false
	}
}
