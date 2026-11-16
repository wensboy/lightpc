package pattern

import (
	"path"
	"strconv"
	"testing"
)

const benchName = "app.module7.http"

func buildPatterns(count int) map[string]int {
	patterns := make(map[string]int, count)
	for index := 0; index < count; index++ {
		patterns["app.module"+strconv.Itoa(index)+".*"] = index
	}
	return patterns
}

func BenchmarkGlobScan(b *testing.B) {
	for _, count := range []int{50, 500} {
		b.Run(strconv.Itoa(count), func(b *testing.B) {
			patterns := buildPatterns(count)
			b.ReportAllocs()
			for index := 0; index < b.N; index++ {
				found := 0
				for pattern := range patterns {
					if ok, err := path.Match(pattern, benchName); err == nil && ok {
						found++
					}
				}
				_ = found
			}
		})
	}
}

func BenchmarkTrieMatch(b *testing.B) {
	for _, count := range []int{50, 500} {
		b.Run(strconv.Itoa(count), func(b *testing.B) {
			trie := NewTrie[int](".")
			for index := 0; index < count; index++ {
				trie.Put("app.module"+strconv.Itoa(index)+".http", index)
			}
			b.ReportAllocs()
			for index := 0; index < b.N; index++ {
				trie.Match("app.module7.*")
			}
		})
	}
}

func BenchmarkTrieMatchDeepPattern(b *testing.B) {
	trie := NewTrie[int](".")
	for index := 0; index < 500; index++ {
		trie.Put("app.module"+strconv.Itoa(index)+".http.server", index)
	}
	b.ReportAllocs()
	for index := 0; index < b.N; index++ {
		trie.Match("app.*.*.server")
	}
}
