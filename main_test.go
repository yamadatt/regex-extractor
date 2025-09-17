package main

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name         string
		configData   string
		wantErr      bool
		wantPatterns int
		errContains  string
	}{
		{
			name: "valid config with single pattern",
			configData: `patterns:
  - name: "test pattern"
    pattern: 'test'
    description: "test description"
    replacement: 'replaced'`,
			wantErr:      false,
			wantPatterns: 1,
		},
		{
			name:         "empty config",
			configData:   `patterns: []`,
			wantErr:      false,
			wantPatterns: 0,
		},
		{
			name: "multiple patterns",
			configData: `patterns:
  - name: "pattern1"
    pattern: 'test1'
    description: "desc1"
    replacement: 'rep1'
  - name: "pattern2"
    pattern: 'test2'
    description: "desc2"
    replacement: 'rep2'`,
			wantErr:      false,
			wantPatterns: 2,
		},
		{
			name:        "invalid yaml syntax",
			configData:  `invalid: yaml: data:`,
			wantErr:     true,
			errContains: "YAML解析エラー",
		},
		{
			name:        "missing patterns key",
			configData:  `not_patterns: []`,
			wantErr:     false,
			wantPatterns: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary config file
			tmpfile, err := os.CreateTemp("", "config*.yaml")
			require.NoError(t, err)
			defer os.Remove(tmpfile.Name())

			_, err = tmpfile.Write([]byte(tt.configData))
			require.NoError(t, err)
			require.NoError(t, tmpfile.Close())

			config, err := loadConfig(tmpfile.Name())

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					require.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				require.NoError(t, err)
				require.NotNil(t, config)
				require.Len(t, config.Patterns, tt.wantPatterns)
			}
		})
	}
}

func TestLoadConfig_FileErrors(t *testing.T) {
	tests := []struct {
		name        string
		filename    string
		errContains string
	}{
		{
			name:        "file not found",
			filename:    "non_existent_file.yaml",
			errContains: "設定ファイルの読み込みに失敗",
		},
		{
			name:        "directory instead of file",
			filename:    ".",
			errContains: "設定ファイルの読み込みに失敗",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := loadConfig(tt.filename)
			require.Error(t, err)
			require.Contains(t, err.Error(), tt.errContains)
		})
	}
}

func TestGenerateOutputFileName(t *testing.T) {
	tests := []struct {
		name      string
		inputFile string
		expected  string
	}{
		{
			name:      "simple filename with extension",
			inputFile: "/path/to/file.txt",
			expected:  "/path/to/file_replaced.txt",
		},
		{
			name:      "filename without extension",
			inputFile: "/path/to/file",
			expected:  "/path/to/file_replaced",
		},
		{
			name:      "filename with multiple dots",
			inputFile: "/path/to/file.test.txt",
			expected:  "/path/to/file.test_replaced.txt",
		},
		{
			name:      "relative path",
			inputFile: "test.txt",
			expected:  "test_replaced.txt",
		},
		{
			name:      "path with spaces",
			inputFile: "/path with spaces/file.txt",
			expected:  "/path with spaces/file_replaced.txt",
		},
		{
			name:      "deep nested path",
			inputFile: "/a/b/c/d/e/file.txt",
			expected:  "/a/b/c/d/e/file_replaced.txt",
		},
		{
			name:      "hidden file",
			inputFile: "/path/.hidden.txt",
			expected:  "/path/.hidden_replaced.txt",
		},
		{
			name:      "file with dash",
			inputFile: "my-file.txt",
			expected:  "my-file_replaced.txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generateOutputFileName(tt.inputFile)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestPerformReplacements(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		config   *Config
		expected string
	}{
		{
			name: "simple text replacement",
			text: "This is a test text with test words",
			config: &Config{
				Patterns: []Pattern{
					{
						Name:        "test pattern",
						Pattern:     "test",
						Description: "replace test",
						Replacement: "sample",
					},
				},
			},
			expected: "This is a sample text with sample words",
		},
		{
			name: "multiple patterns applied sequentially",
			text: "Hello world and goodbye world",
			config: &Config{
				Patterns: []Pattern{
					{
						Name:        "hello",
						Pattern:     "Hello",
						Description: "replace Hello",
						Replacement: "Hi",
					},
					{
						Name:        "goodbye",
						Pattern:     "goodbye",
						Description: "replace goodbye",
						Replacement: "farewell",
					},
				},
			},
			expected: "Hi world and farewell world",
		},
		{
			name: "regex with capture groups",
			text: "The price is $100 and $200",
			config: &Config{
				Patterns: []Pattern{
					{
						Name:        "price",
						Pattern:     `\$(\d+)`,
						Description: "capture price",
						Replacement: "USD $1",
					},
				},
			},
			expected: "The price is USD 100 and USD 200",
		},
		{
			name: "Japanese text with brackets",
			text: "タイトル『テスト』です",
			config: &Config{
				Patterns: []Pattern{
					{
						Name:        "brackets",
						Pattern:     `『([^』]*)』`,
						Description: "remove brackets",
						Replacement: "$1",
					},
				},
			},
			expected: "タイトルテストです",
		},
		{
			name: "multiline pattern with (?s) flag",
			text: "Start\nMiddle\nEnd",
			config: &Config{
				Patterns: []Pattern{
					{
						Name:        "multiline",
						Pattern:     `Start.*End`,
						Description: "match across lines",
						Replacement: "Replaced",
					},
				},
			},
			expected: "Replaced",
		},
		{
			name: "empty pattern skipped",
			text: "Original text",
			config: &Config{
				Patterns: []Pattern{
					{
						Name:        "empty",
						Pattern:     "",
						Description: "empty pattern",
						Replacement: "replacement",
					},
				},
			},
			expected: "Original text",
		},
		{
			name: "no matches found",
			text: "Original text",
			config: &Config{
				Patterns: []Pattern{
					{
						Name:        "no match",
						Pattern:     "notfound",
						Description: "won't match",
						Replacement: "replaced",
					},
				},
			},
			expected: "Original text",
		},
		{
			name:     "nil config",
			text:     "Original text",
			config:   nil,
			expected: "Original text",
		},
		{
			name: "complex HTML replacement",
			text: `<p style="color: red;">Text</p>`,
			config: &Config{
				Patterns: []Pattern{
					{
						Name:        "style",
						Pattern:     ` style="[^"]*"`,
						Description: "remove style",
						Replacement: "",
					},
				},
			},
			expected: `<p>Text</p>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := performReplacements(tt.text, tt.config)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestPerformReplacements_ErrorHandling(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		config   *Config
		expected string
	}{
		{
			name: "invalid regex pattern",
			text: "test text",
			config: &Config{
				Patterns: []Pattern{
					{
						Name:        "invalid",
						Pattern:     "[invalid(regex",
						Description: "invalid regex",
						Replacement: "replaced",
					},
				},
			},
			expected: "test text", // Should return original on error
		},
		{
			name: "pattern with invalid backreference",
			text: "test text",
			config: &Config{
				Patterns: []Pattern{
					{
						Name:        "invalid backref",
						Pattern:     "(test)",
						Description: "invalid replacement",
						Replacement: "$9", // Non-existent group
					},
				},
			},
			expected: " text", // Go replaces invalid backreferences with empty string
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := performReplacements(tt.text, tt.config)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestMatch_Structure(t *testing.T) {
	tests := []struct {
		name        string
		match       Match
		wantPattern string
		wantLine    int
		wantText    string
		wantMatches []string
	}{
		{
			name: "basic match structure",
			match: Match{
				PatternName: "test",
				Line:        10,
				Text:        "matched text",
				Matches:     []string{"match1", "match2"},
			},
			wantPattern: "test",
			wantLine:    10,
			wantText:    "matched text",
			wantMatches: []string{"match1", "match2"},
		},
		{
			name: "empty matches",
			match: Match{
				PatternName: "empty",
				Line:        1,
				Text:        "",
				Matches:     []string{},
			},
			wantPattern: "empty",
			wantLine:    1,
			wantText:    "",
			wantMatches: []string{},
		},
		{
			name: "single match",
			match: Match{
				PatternName: "single",
				Line:        100,
				Text:        "single match",
				Matches:     []string{"single"},
			},
			wantPattern: "single",
			wantLine:    100,
			wantText:    "single match",
			wantMatches: []string{"single"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.wantPattern, tt.match.PatternName)
			require.Equal(t, tt.wantLine, tt.match.Line)
			require.Equal(t, tt.wantText, tt.match.Text)
			require.Equal(t, tt.wantMatches, tt.match.Matches)
		})
	}
}

// Benchmark tests
func BenchmarkPerformReplacements(b *testing.B) {
	benchmarks := []struct {
		name   string
		text   string
		config *Config
	}{
		{
			name: "small text",
			text: "This is a test text with multiple test occurrences.",
			config: &Config{
				Patterns: []Pattern{
					{
						Name:        "test",
						Pattern:     "test",
						Description: "replace test",
						Replacement: "sample",
					},
				},
			},
		},
		{
			name: "medium text",
			text: strings.Repeat("This is a test text with multiple test occurrences. ", 100),
			config: &Config{
				Patterns: []Pattern{
					{
						Name:        "test",
						Pattern:     "test",
						Description: "replace test",
						Replacement: "sample",
					},
				},
			},
		},
		{
			name: "large text",
			text: strings.Repeat("This is a test text with multiple test occurrences. ", 1000),
			config: &Config{
				Patterns: []Pattern{
					{
						Name:        "test",
						Pattern:     "test",
						Description: "replace test",
						Replacement: "sample",
					},
				},
			},
		},
		{
			name: "multiple patterns",
			text: strings.Repeat("Hello world and goodbye world. ", 100),
			config: &Config{
				Patterns: []Pattern{
					{
						Name:        "hello",
						Pattern:     "Hello",
						Replacement: "Hi",
					},
					{
						Name:        "goodbye",
						Pattern:     "goodbye",
						Replacement: "farewell",
					},
					{
						Name:        "world",
						Pattern:     "world",
						Replacement: "universe",
					},
				},
			},
		},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				performReplacements(bm.text, bm.config)
			}
		})
	}
}

func BenchmarkLoadConfig(b *testing.B) {
	configSizes := []struct {
		name       string
		configData string
	}{
		{
			name: "small config",
			configData: `patterns:
  - name: "pattern1"
    pattern: 'test1'
    description: "description1"
    replacement: 'replace1'`,
		},
		{
			name: "medium config",
			configData: `patterns:
  - name: "pattern1"
    pattern: 'test1'
    description: "description1"
    replacement: 'replace1'
  - name: "pattern2"
    pattern: 'test2'
    description: "description2"
    replacement: 'replace2'
  - name: "pattern3"
    pattern: 'test3'
    description: "description3"
    replacement: 'replace3'
  - name: "pattern4"
    pattern: 'test4'
    description: "description4"
    replacement: 'replace4'
  - name: "pattern5"
    pattern: 'test5'
    description: "description5"
    replacement: 'replace5'`,
		},
	}

	for _, cs := range configSizes {
		b.Run(cs.name, func(b *testing.B) {
			tmpfile, err := os.CreateTemp("", "benchmark_config*.yaml")
			require.NoError(b, err)
			defer os.Remove(tmpfile.Name())

			_, err = tmpfile.Write([]byte(cs.configData))
			require.NoError(b, err)
			require.NoError(b, tmpfile.Close())

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _ = loadConfig(tmpfile.Name())
			}
		})
	}
}