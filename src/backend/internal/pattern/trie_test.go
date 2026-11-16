package pattern

import (
	"slices"
	"strconv"
	"sync"
	"testing"
)

func TestTriePutGetAndUpdate(t *testing.T) {
	trie := NewTrie[int](".")
	trie.Put("app.http.server", 1)
	trie.Put("app.http.client", 2)

	if value, ok := trie.Get("app.http.server"); !ok || value != 1 {
		t.Fatalf("value = %v, %v", value, ok)
	}

	trie.Put("app.http.server", 3)
	if value, ok := trie.Get("app.http.server"); !ok || value != 3 {
		t.Fatalf("value = %v, %v", value, ok)
	}
	if trie.Len() != 2 {
		t.Fatalf("len = %d", trie.Len())
	}
}

func TestTrieSepAndEmptySegments(t *testing.T) {
	trie := NewTrie[int]("/")
	if trie.Sep() != "/" {
		t.Fatalf("sep = %q", trie.Sep())
	}

	trie.Put("/app//http/", 1)
	if value, ok := trie.Get("app/http"); !ok || value != 1 {
		t.Fatalf("value = %v, %v", value, ok)
	}
}

func TestTrieDefaultSep(t *testing.T) {
	trie := NewTrie[int]("")
	trie.Put("a.b", 1)

	if _, ok := trie.Get("a.b"); !ok {
		t.Fatal("the default separator must split keys")
	}
}

func TestTrieDeletePrunesEmptyBranch(t *testing.T) {
	trie := NewTrie[int](".")
	trie.Put("app.http.server", 1)

	if !trie.Delete("app.http.server") {
		t.Fatal("delete must report the removal")
	}
	if trie.Delete("app.http.server") {
		t.Fatal("a second delete must report false")
	}
	if trie.Len() != 0 {
		t.Fatalf("len = %d", trie.Len())
	}
	if len(trie.root.children) != 0 {
		t.Fatalf("children = %v", trie.root.children)
	}
}

func TestTrieDeleteKeepsSiblings(t *testing.T) {
	trie := NewTrie[int](".")
	trie.Put("app.http.server", 1)
	trie.Put("app.http.client", 2)

	trie.Delete("app.http.server")
	if _, ok := trie.Get("app.http.client"); !ok {
		t.Fatal("the sibling must survive")
	}
}

func TestTrieValuesByPrefix(t *testing.T) {
	trie := NewTrie[int](".")
	trie.Put("app", 1)
	trie.Put("app.http", 2)
	trie.Put("app.http.server", 3)
	trie.Put("worker", 4)

	values := trie.Values("app")
	slices.Sort(values)
	if !slices.Equal(values, []int{1, 2, 3}) {
		t.Fatalf("values = %v", values)
	}
	if values := trie.Values("missing"); values != nil {
		t.Fatalf("values = %v", values)
	}
}

func TestTrieMatchWildcard(t *testing.T) {
	trie := NewTrie[int](".")
	trie.Put("app.http", 1)
	trie.Put("app.ws", 2)
	trie.Put("app.http.server", 3)
	trie.Put("worker.http", 4)

	values := trie.Match("app.*")
	slices.Sort(values)
	if !slices.Equal(values, []int{1, 2}) {
		t.Fatalf("values = %v", values)
	}

	values = trie.Match("*.http")
	slices.Sort(values)
	if !slices.Equal(values, []int{1, 4}) {
		t.Fatalf("values = %v", values)
	}

	if values := trie.Match("app.*.server"); !slices.Equal(values, []int{3}) {
		t.Fatalf("values = %v", values)
	}
}

func TestTrieDeletePrefix(t *testing.T) {
	trie := NewTrie[int](".")
	trie.Put("app.http", 1)
	trie.Put("app.ws", 2)
	trie.Put("worker", 3)

	if removed := trie.DeletePrefix("app"); removed != 2 {
		t.Fatalf("removed = %d", removed)
	}
	if trie.Len() != 1 {
		t.Fatalf("len = %d", trie.Len())
	}
	if removed := trie.DeletePrefix(""); removed != 1 {
		t.Fatalf("removed = %d", removed)
	}
	if trie.Len() != 0 || len(trie.root.children) != 0 {
		t.Fatalf("len = %d, children = %v", trie.Len(), trie.root.children)
	}
}

func TestTrieConcurrentAccess(t *testing.T) {
	trie := NewTrie[int](".")
	var waitGroup sync.WaitGroup
	for index := 0; index < 8; index++ {
		waitGroup.Add(1)
		go func(index int) {
			defer waitGroup.Done()
			trie.Put("app.http."+strconv.Itoa(index), index)
			trie.Values("app")
			trie.Match("app.*.*")
		}(index)
	}
	waitGroup.Wait()

	if trie.Len() != 8 {
		t.Fatalf("len = %d", trie.Len())
	}
}
