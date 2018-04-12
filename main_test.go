package main

import "testing"

func TestDirname(t *testing.T) {
	tests := []struct {
		desc   string
		url    string
		prefix string
		want   string
	}{
		{
			desc:   "basic",
			url:    "http://example.com",
			prefix: "aaa",
			want:   "aaa/example.com",
		},
		{
			desc:   "with 1 subdirectory",
			url:    "http://example.com/bbb",
			prefix: "aaa",
			want:   "aaa/example.com",
		},
		{
			desc:   "with 1 subdirectory ended by a slash",
			url:    "http://example.com/bbb/",
			prefix: "aaa",
			want:   "aaa/example.com/bbb",
		},
	}
	for _, tt := range tests {
		dir, err := dirname(tt.url, tt.prefix)
		if err != nil {
			t.Fatalf("dirname(%v): %v", tt.desc, err)
		}
		if dir != tt.want {
			t.Errorf("dirname(%v): got %q, but want %q", tt.desc, dir, tt.want)
		}
	}
}
