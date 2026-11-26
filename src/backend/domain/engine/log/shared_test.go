package engineLog

import "testing"

func TestEntriesPicksTheSubGroup(t *testing.T) {
	config := Config{
		Direct:  []LogEntry{{Name: "only.direct", Handler: "console"}},
		Fanout:  []LogEntry{{Name: "fan.out", Handler: "file"}, {Name: ""}},
		Pattern: []LogEntry{{Name: "pat", Handler: "console"}},
	}

	// The unnamed entry cannot be addressed, so it is dropped rather than producing an engine.
	if entries := config.Entries(SubFanout); len(entries) != 1 || entries[0].Name != "fan.out" {
		t.Fatalf("fanout = %+v", entries)
	}
	if entries := config.Entries(SubDirect); len(entries) != 1 || entries[0].Name != "only.direct" {
		t.Fatalf("direct = %+v", entries)
	}
	if entries := config.Entries(SubPattern); len(entries) != 1 || entries[0].Name != "pat" {
		t.Fatalf("pattern = %+v", entries)
	}
}

func TestEntriesOfAnUnknownSubGroupIsEmpty(t *testing.T) {
	config := Config{Direct: []LogEntry{{Name: "x"}}}

	// An empty slice rather than nil: a caller ranges over the answer without a nil check.
	if entries := config.Entries("unknown"); entries == nil || len(entries) != 0 {
		t.Fatalf("entries = %+v", entries)
	}
}

func TestEntriesKeepsTheOutputRaw(t *testing.T) {
	config := Config{Fanout: []LogEntry{{
		Name:    "dev.file",
		Handler: "file",
		Output:  []byte(`[{"path":"local://~/logs/a.log"}]`),
	}}}

	entries := config.Entries(SubFanout)
	if len(entries) != 1 || len(entries[0].Output) == 0 {
		t.Fatalf("entries = %+v", entries)
	}
}
