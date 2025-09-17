package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIntegration_CompleteWorkflow(t *testing.T) {
	tests := []struct {
		name           string
		inputContent   string
		configContent  string
		expectedResult string
		checkPoints    []string // Strings that should exist in result
		notContains    []string // Strings that should NOT exist in result
	}{
		{
			name: "basic replacement workflow",
			inputContent: `TITLE: サンプル『タイトル』です
CATEGORY: Test
AUTHOR: テスト著者

これはテスト文書です。
oldtextを含む行です。
もう一つoldtextがあります。

別のセクション
CATEGORY: Keep
この部分は保持されます。`,
			configContent: `patterns:
  - name: "brackets"
    pattern: '『([^』]*)』'
    description: "Remove brackets"
    replacement: '$1'
  - name: "oldtext"
    pattern: 'oldtext'
    description: "Replace oldtext"
    replacement: 'newtext'`,
			checkPoints: []string{
				"タイトル",
				"newtext",
				"CATEGORY: Keep",
				"この部分は保持されます",
			},
			notContains: []string{
				"『タイトル』",
				"oldtext",
			},
		},
		{
			name: "HTML tag removal",
			inputContent: `<p style="color: red;">赤いテキスト</p>
<div class="container">コンテンツ</div>
<span style="font-size: 14px;">サイズ指定</span>`,
			configContent: `patterns:
  - name: "remove styles"
    pattern: ' style="[^"]*"'
    description: "Remove style attributes"
    replacement: ''
  - name: "remove class"
    pattern: ' class="[^"]*"'
    description: "Remove class attributes"
    replacement: ''`,
			checkPoints: []string{
				"<p>赤いテキスト</p>",
				"<div>コンテンツ</div>",
				"<span>サイズ指定</span>",
			},
			notContains: []string{
				"style=",
				"class=",
			},
		},
		{
			name: "line deletion pattern",
			inputContent: `行1
削除対象行: これを削除
行2
削除対象行: これも削除
行3`,
			configContent: `patterns:
  - name: "delete lines"
    pattern: '削除対象行: [^\n]*\n'
    description: "Delete specific lines"
    replacement: ''`,
			checkPoints: []string{
				"行1",
				"行2",
				"行3",
			},
			notContains: []string{
				"削除対象行",
				"これを削除",
				"これも削除",
			},
		},
		{
			name: "capture group replacement",
			inputContent: `The price is $100
The price is $250
The price is $99`,
			configContent: `patterns:
  - name: "price format"
    pattern: '\$(\d+)'
    description: "Format prices"
    replacement: '¥$1'`,
			checkPoints: []string{
				"¥100",
				"¥250",
				"¥99",
			},
			notContains: []string{
				"$100",
				"$250",
				"$99",
			},
		},
		{
			name: "multiple sequential replacements",
			inputContent: `Step1 -> Step2 -> Step3`,
			configContent: `patterns:
  - name: "first"
    pattern: 'Step1'
    replacement: 'Phase1'
  - name: "second"
    pattern: 'Step2'
    replacement: 'Phase2'
  - name: "third"
    pattern: 'Step3'
    replacement: 'Phase3'`,
			expectedResult: `Phase1 -> Phase2 -> Phase3`,
			notContains: []string{
				"Step1",
				"Step2",
				"Step3",
			},
		},
		{
			name: "empty pattern skipped",
			inputContent: `Original content here`,
			configContent: `patterns:
  - name: "empty"
    pattern: ''
    description: "Empty pattern"
    replacement: 'should not replace'
  - name: "valid"
    pattern: 'content'
    replacement: 'text'`,
			expectedResult: `Original text here`,
			notContains: []string{
				"should not replace",
				"content",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary directory for test
			tmpDir := t.TempDir()

			// Create input file
			inputFile := filepath.Join(tmpDir, "input.txt")
			err := os.WriteFile(inputFile, []byte(tt.inputContent), 0644)
			require.NoError(t, err)

			// Create config file
			configFile := filepath.Join(tmpDir, "config.yaml")
			err = os.WriteFile(configFile, []byte(tt.configContent), 0644)
			require.NoError(t, err)

			// Load config
			config, err := loadConfig(configFile)
			require.NoError(t, err)

			// Read input
			content, err := os.ReadFile(inputFile)
			require.NoError(t, err)

			// Perform replacements
			result := performReplacements(string(content), config)

			// Check expected result if provided
			if tt.expectedResult != "" {
				require.Equal(t, tt.expectedResult, result)
			}

			// Check that required strings are present
			for _, checkPoint := range tt.checkPoints {
				require.Contains(t, result, checkPoint, "Expected to find: %s", checkPoint)
			}

			// Check that unwanted strings are not present
			for _, notContain := range tt.notContains {
				require.NotContains(t, result, notContain, "Should not contain: %s", notContain)
			}

			// Test output file generation
			outputFile := generateOutputFileName(inputFile)
			expectedOutput := filepath.Join(tmpDir, "input_replaced.txt")
			require.Equal(t, expectedOutput, outputFile)
		})
	}
}

func TestIntegration_ReplaceMode(t *testing.T) {
	tests := []struct {
		name              string
		inputContent      string
		configContent     string
		expectedInFile    []string
		notExpectedInFile []string
	}{
		{
			name: "replace mode creates output file",
			inputContent: `TITLE: これは『テスト』タイトルです
CATEGORY: Test
本文にoldtextが含まれています。`,
			configContent: `patterns:
  - name: "brackets"
    pattern: '『([^』]*)』'
    replacement: '$1'
  - name: "oldtext"
    pattern: 'oldtext'
    replacement: 'newtext'`,
			expectedInFile: []string{
				"テスト",
				"newtext",
			},
			notExpectedInFile: []string{
				"『テスト』",
				"oldtext",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary directory
			tmpDir := t.TempDir()

			// Create input file
			inputFile := filepath.Join(tmpDir, "test.txt")
			err := os.WriteFile(inputFile, []byte(tt.inputContent), 0644)
			require.NoError(t, err)

			// Create config file
			configFile := filepath.Join(tmpDir, "config.yaml")
			err = os.WriteFile(configFile, []byte(tt.configContent), 0644)
			require.NoError(t, err)

			// Load config and perform replacements
			config, err := loadConfig(configFile)
			require.NoError(t, err)

			content, err := os.ReadFile(inputFile)
			require.NoError(t, err)

			result := performReplacements(string(content), config)

			// Write result to output file
			outputFile := generateOutputFileName(inputFile)
			err = os.WriteFile(outputFile, []byte(result), 0644)
			require.NoError(t, err)

			// Verify output file exists
			require.FileExists(t, outputFile)

			// Read and verify content
			outputContent, err := os.ReadFile(outputFile)
			require.NoError(t, err)
			outputStr := string(outputContent)

			// Check expected content
			for _, expected := range tt.expectedInFile {
				require.Contains(t, outputStr, expected)
			}

			// Check not expected content
			for _, notExpected := range tt.notExpectedInFile {
				require.NotContains(t, outputStr, notExpected)
			}
		})
	}
}

func TestIntegration_ConfigValidation(t *testing.T) {
	tests := []struct {
		name          string
		configContent string
		wantErr       bool
		errContains   string
	}{
		{
			name: "valid config",
			configContent: `patterns:
  - name: "test"
    pattern: "test"
    description: "test"
    replacement: "replace"`,
			wantErr: false,
		},
		{
			name:          "invalid YAML",
			configContent: `invalid: : yaml: :`,
			wantErr:       true,
			errContains:   "YAML解析エラー",
		},
		{
			name:          "empty file",
			configContent: ``,
			wantErr:       false, // Empty config is valid, just has no patterns
		},
		{
			name: "missing required fields",
			configContent: `patterns:
  - name: "test"`,
			wantErr: false, // Missing fields are allowed, pattern will be empty
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			configFile := filepath.Join(tmpDir, "config.yaml")

			err := os.WriteFile(configFile, []byte(tt.configContent), 0644)
			require.NoError(t, err)

			config, err := loadConfig(configFile)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					require.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				require.NoError(t, err)
				require.NotNil(t, config)
			}
		})
	}
}

func TestIntegration_EdgeCases(t *testing.T) {
	tests := []struct {
		name          string
		inputContent  string
		configContent string
		description   string
	}{
		{
			name:         "empty input file",
			inputContent: "",
			configContent: `patterns:
  - name: "test"
    pattern: "test"
    replacement: "replace"`,
			description: "Should handle empty input gracefully",
		},
		{
			name:         "no matching patterns",
			inputContent: "This is some content",
			configContent: `patterns:
  - name: "no match"
    pattern: "xyz123"
    replacement: "replace"`,
			description: "Should return original when no matches",
		},
		{
			name:         "complex regex pattern",
			inputContent: "email@example.com",
			configContent: `patterns:
  - name: "email"
    pattern: "([a-zA-Z0-9]+)@([a-zA-Z0-9]+)\\.([a-zA-Z]+)"
    replacement: "$1 at $2 dot $3"`,
			description: "Should handle complex regex",
		},
		{
			name:         "unicode content",
			inputContent: "テスト 😀 テスト",
			configContent: `patterns:
  - name: "emoji"
    pattern: "😀"
    replacement: "😎"`,
			description: "Should handle unicode/emoji",
		},
		{
			name:         "newline patterns",
			inputContent: "line1\n\nline2",
			configContent: `patterns:
  - name: "double newline"
    pattern: '\n\n'
    replacement: '\n'`,
			description: "Should handle newline patterns",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			// Create files
			inputFile := filepath.Join(tmpDir, "input.txt")
			err := os.WriteFile(inputFile, []byte(tt.inputContent), 0644)
			require.NoError(t, err)

			configFile := filepath.Join(tmpDir, "config.yaml")
			err = os.WriteFile(configFile, []byte(tt.configContent), 0644)
			require.NoError(t, err)

			// Load and process
			config, err := loadConfig(configFile)
			require.NoError(t, err)

			content, err := os.ReadFile(inputFile)
			require.NoError(t, err)

			// Should not panic
			result := performReplacements(string(content), config)
			require.NotNil(t, result)
		})
	}
}