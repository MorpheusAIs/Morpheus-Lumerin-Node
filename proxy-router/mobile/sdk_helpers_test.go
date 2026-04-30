package mobile

import (
	"reflect"
	"testing"
)

func TestParseEthNodeURLs(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: nil,
		},
		{
			name:     "whitespace only",
			input:    "  ",
			expected: nil,
		},
		{
			name:     "single url",
			input:    "https://rpc.example.com",
			expected: []string{"https://rpc.example.com"},
		},
		{
			name:     "comma separated",
			input:    "https://a.com,https://b.com",
			expected: []string{"https://a.com", "https://b.com"},
		},
		{
			name:     "semicolon separated",
			input:    "https://a.com;https://b.com",
			expected: []string{"https://a.com", "https://b.com"},
		},
		{
			name:     "pipe separated",
			input:    "https://a.com|https://b.com",
			expected: []string{"https://a.com", "https://b.com"},
		},
		{
			name:     "newline separated",
			input:    "https://a.com\nhttps://b.com",
			expected: []string{"https://a.com", "https://b.com"},
		},
		{
			name:     "dedup",
			input:    "https://a.com,https://a.com",
			expected: []string{"https://a.com"},
		},
		{
			name:     "whitespace around urls",
			input:    "  https://a.com , https://b.com  ",
			expected: []string{"https://a.com", "https://b.com"},
		},
		{
			name:     "empty segments",
			input:    ",,,https://a.com,,,",
			expected: []string{"https://a.com"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseEthNodeURLs(tt.input)
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("parseEthNodeURLs(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}
