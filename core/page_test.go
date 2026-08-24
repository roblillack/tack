package core

import "testing"

func TestTitleWords(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		// Ordinary slugs: first letter of each word, hyphens left in place.
		{"hello-world", "Hello-World"},
		{"index", "Index"},
		{"a-b-c", "A-B-C"},
		{"hello world", "Hello World"},

		// The remainder of a word is never touched.
		{"hELLO-wORLD", "HELLO-WORLD"},
		{"MiXeD-CaSe", "MiXeD-CaSe"},
		{"HTML-and-CSS", "HTML-And-CSS"},

		// Digits are word-internal, so "2nd" does not become "2Nd".
		{"2nd-place", "2nd-Place"},
		{"3rd-time", "3rd-Time"},
		{"4k-video", "4k-Video"},
		{"post-2024-01-01", "Post-2024-01-01"},
		{"x1-y2", "X1-Y2"},

		// Apostrophes are word-internal too -- strings.Title got these wrong.
		{"it's-a-test", "It's-A-Test"},
		{"don't-panic", "Don't-Panic"},
		{"rock'n'roll", "Rock'n'roll"},
		{"don’t-worry", "Don’t-Worry"},

		// Non-ASCII letters, including a title-case digraph.
		{"über-uns", "Über-Uns"},
		{"café-au-lait", "Café-Au-Lait"},
		{"ǳ-digraph", "ǲ-Digraph"},

		// Non-ASCII punctuation and symbols separate words too, which
		// strings.Title did not do.
		{"a–b", "A–B"},
		{"rock—and—roll", "Rock—And—Roll"},
		{"foo·bar", "Foo·Bar"},
		{"«zitat»", "«Zitat»"},
		{"a\u00a0b", "A\u00a0B"},
		{"pi≈three", "Pi≈Three"},

		// Edge cases.
		{"", ""},
		{"foo--bar", "Foo--Bar"},
		{"-leading", "-Leading"},
		{"trailing-", "Trailing-"},
		{"_under_score", "_Under_Score"},
		{"foo_bar", "Foo_Bar"},
		{"  spaced  ", "  Spaced  "},
	}

	for _, tt := range tests {
		if got := titleWords(tt.in); got != tt.want {
			t.Errorf("titleWords(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}
