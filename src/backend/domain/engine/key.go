package engine

const (
	GroupLog    = "log"
	NameDirect  = "direct"
	NameFanout  = "fanout"
	NamePattern = "pattern"

	GroupSetting = "setting"
	NameSetting  = "setting"

	GroupStore = "store"
	NameFs     = "fs"
	NameDb     = "db"
)

var (
	LogDirect  = EngineDep(GroupLog, NameDirect)
	LogFanout  = EngineDep(GroupLog, NameFanout)
	LogPattern = EngineDep(GroupLog, NamePattern)

	Setting = EngineDep(GroupSetting, NameSetting)

	StoreFs = EngineDep(GroupStore, NameFs)
	StoreDb = EngineDep(GroupStore, NameDb)
)
