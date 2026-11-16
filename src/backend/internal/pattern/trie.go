// Package pattern provides a separator aware prefix tree for pattern style lookups.
package pattern

import (
	"strings"
	"sync"
)

// Trie stores values of type T keyed by separator segmented keys. Every non empty segment becomes a
// node, so a key such as "app.http.server" is reachable in three hops when Sep is ".".
type Trie[T any] struct {
	sep  string
	root *node[T]

	mu    sync.RWMutex
	count int
}

// node is one segment of a key. A node may be a pure path node or carry a value.
type node[T any] struct {
	children map[string]*node[T]
	value    T
	hasValue bool
}

// NewTrie builds an empty trie splitting keys by sep. An empty sep falls back to ".".
func NewTrie[T any](sep string) *Trie[T] {
	if sep == "" {
		sep = "."
	}
	return &Trie[T]{sep: sep, root: newNode[T]()}
}

// Sep returns the separator used to split keys.
func (t *Trie[T]) Sep() string { return t.sep }

// Put inserts value at key, replacing any value already stored there.
func (t *Trie[T]) Put(key string, value T) {
	segments := splitKey(key, t.sep)

	t.mu.Lock()
	defer t.mu.Unlock()
	current := t.root
	for _, segment := range segments {
		child, found := current.children[segment]
		if !found {
			child = newNode[T]()
			current.children[segment] = child
		}
		current = child
	}
	if !current.hasValue {
		t.count++
	}
	current.value = value
	current.hasValue = true
}

// Get returns the value stored exactly at key.
func (t *Trie[T]) Get(key string) (T, bool) {
	segments := splitKey(key, t.sep)

	t.mu.RLock()
	defer t.mu.RUnlock()
	current := t.root
	for _, segment := range segments {
		child, found := current.children[segment]
		if !found {
			var zero T
			return zero, false
		}
		current = child
	}
	return current.value, current.hasValue
}

// Delete removes the value stored exactly at key and reports whether it existed. Branches left empty
// by the removal are pruned so the tree does not keep dead paths.
func (t *Trie[T]) Delete(key string) bool {
	segments := splitKey(key, t.sep)

	t.mu.Lock()
	defer t.mu.Unlock()
	path := make([]*node[T], 0, len(segments)+1)
	current := t.root
	path = append(path, current)
	for _, segment := range segments {
		child, found := current.children[segment]
		if !found {
			return false
		}
		current = child
		path = append(path, current)
	}
	if !current.hasValue {
		return false
	}

	var zero T
	current.value = zero
	current.hasValue = false
	t.count--

	for index := len(path) - 1; index > 0; index-- {
		pruned := path[index]
		if pruned.hasValue || len(pruned.children) > 0 {
			break
		}
		delete(path[index-1].children, segments[index-1])
	}
	return true
}

// DeletePrefix removes prefix and every value stored below it, reporting how many values were removed.
// An empty prefix clears the whole tree.
func (t *Trie[T]) DeletePrefix(prefix string) int {
	segments := splitKey(prefix, t.sep)

	t.mu.Lock()
	defer t.mu.Unlock()
	path := make([]*node[T], 0, len(segments)+1)
	current := t.root
	path = append(path, current)
	for _, segment := range segments {
		child, found := current.children[segment]
		if !found {
			return 0
		}
		current = child
		path = append(path, current)
	}

	removed := countValues(current)
	t.count -= removed
	if len(segments) == 0 {
		t.root = newNode[T]()
		return removed
	}
	delete(path[len(path)-2].children, segments[len(segments)-1])
	return removed
}

// Values returns every value stored under prefix, including the value stored at prefix itself.
func (t *Trie[T]) Values(prefix string) []T {
	segments := splitKey(prefix, t.sep)

	t.mu.RLock()
	defer t.mu.RUnlock()
	current := t.root
	for _, segment := range segments {
		child, found := current.children[segment]
		if !found {
			return nil
		}
		current = child
	}

	values := make([]T, 0, countValues(current))
	collect(current, &values)
	return values
}

// Match returns every value whose key matches pattern. A "*" segment matches exactly one segment;
// any other segment is compared literally. Only branches that exist are visited.
func (t *Trie[T]) Match(pattern string) []T {
	segments := splitKey(pattern, t.sep)

	t.mu.RLock()
	defer t.mu.RUnlock()
	values := make([]T, 0)
	match(t.root, segments, &values)
	return values
}

// Len reports how many values are stored.
func (t *Trie[T]) Len() int {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.count
}

func newNode[T any]() *node[T] {
	return &node[T]{children: map[string]*node[T]{}}
}

// splitKey splits key by sep and drops empty segments so leading, trailing or doubled separators are
// tolerated.
func splitKey(key, sep string) []string {
	parts := strings.Split(key, sep)
	segments := parts[:0]
	for _, part := range parts {
		if part != "" {
			segments = append(segments, part)
		}
	}
	return segments
}

func collect[T any](n *node[T], values *[]T) {
	if n.hasValue {
		*values = append(*values, n.value)
	}
	for _, child := range n.children {
		collect(child, values)
	}
}

func match[T any](n *node[T], segments []string, values *[]T) {
	if len(segments) == 0 {
		if n.hasValue {
			*values = append(*values, n.value)
		}
		return
	}
	if segments[0] == "*" {
		for _, child := range n.children {
			match(child, segments[1:], values)
		}
		return
	}
	if child, found := n.children[segments[0]]; found {
		match(child, segments[1:], values)
	}
}

func countValues[T any](n *node[T]) int {
	count := 0
	if n.hasValue {
		count++
	}
	for _, child := range n.children {
		count += countValues(child)
	}
	return count
}
