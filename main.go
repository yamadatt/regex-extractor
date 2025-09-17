package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	yaml "gopkg.in/yaml.v2"
)

type Pattern struct {
	Name        string `yaml:"name"`
	Pattern     string `yaml:"pattern"`
	Description string `yaml:"description"`
	Replacement string `yaml:"replacement"`
}

type Config struct {
	Patterns []Pattern `yaml:"patterns"`
}

type Match struct {
	PatternName string
	Line        int
	Text        string
	Matches     []string
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("使用方法: go run main.go <入力ファイルパス> [設定ファイルパス] [オプション]")
		fmt.Println("例: go run main.go /home/yamadatt/git/ameblo_url_list/interi20250915.txt")
		fmt.Println("    go run main.go /home/yamadatt/git/ameblo_url_list/interi20250915.txt config.yaml")
		fmt.Println("    go run main.go /home/yamadatt/git/ameblo_url_list/interi20250915.txt config.yaml --replace")
		fmt.Println("")
		fmt.Println("オプション:")
		fmt.Println("  --replace, -r  : 抽出ではなく置換を実行し、結果を出力")
		os.Exit(1)
	}

	inputFile := os.Args[1]
	configFile := "config.yaml"
	replaceMode := false

	// 引数を解析
	for i := 2; i < len(os.Args); i++ {
		arg := os.Args[i]
		if arg == "--replace" || arg == "-r" {
			replaceMode = true
		} else if !strings.HasPrefix(arg, "-") {
			configFile = arg
		}
	}

	config, err := loadConfig(configFile)
	if err != nil {
		log.Fatalf("設定ファイルの読み込みエラー: %v", err)
	}

	content, err := os.ReadFile(inputFile)
	if err != nil {
		log.Fatalf("ファイルの読み込みエラー: %v", err)
	}

	text := string(content)

	if replaceMode {
		// 置換モード
		replacedText := performReplacements(text, config)

		// 出力ファイル名を生成（元ファイル名_replaced.拡張子）
		outputFile := generateOutputFileName(inputFile)

		// ファイルに保存
		err = os.WriteFile(outputFile, []byte(replacedText), 0644)
		if err != nil {
			log.Fatalf("ファイル保存エラー: %v", err)
		}

		fmt.Fprintf(os.Stderr, "置換結果を保存しました: %s\n", outputFile)
	} else {
		// 抽出モード（従来の動作）
		var allMatches []Match

		for _, pattern := range config.Patterns {
			if pattern.Pattern == "" {
				continue
			}

			regex, err := regexp.Compile("(?s)" + pattern.Pattern)
			if err != nil {
				fmt.Printf("正規表現エラー ('%s'): %v\n", pattern.Name, err)
				continue
			}

			matches := regex.FindAllString(text, -1)
			if len(matches) > 0 {
				// マッチした位置を特定して行番号を計算
				for _, match := range matches {
					lineNumber := 1
					index := strings.Index(text, match)
					if index >= 0 {
						lineNumber = strings.Count(text[:index], "\n") + 1
					}

					allMatches = append(allMatches, Match{
						PatternName: pattern.Name,
						Line:        lineNumber,
						Text:        match,
						Matches:     []string{match},
					})
				}
			}
		}

		printResults(allMatches, config)
	}
}

func loadConfig(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("設定ファイルの読み込みに失敗: %w", err)
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("YAML解析エラー: %w", err)
	}

	return &config, nil
}

func performReplacements(text string, config *Config) string {
	if config == nil {
		return text
	}

	result := text
	totalReplacements := 0

	for _, pattern := range config.Patterns {
		if pattern.Pattern == "" {
			continue
		}

		regex, err := regexp.Compile("(?s)" + pattern.Pattern)
		if err != nil {
			fmt.Fprintf(os.Stderr, "正規表現エラー ('%s'): %v\n", pattern.Name, err)
			continue
		}

		// 置換前のマッチ数をカウント
		matches := regex.FindAllString(result, -1)
		matchCount := len(matches)

		if matchCount > 0 {
			// 置換実行
			result = regex.ReplaceAllString(result, pattern.Replacement)
			totalReplacements += matchCount
			fmt.Fprintf(os.Stderr, "[%s] %d件置換しました\n", pattern.Name, matchCount)
		}
	}

	fmt.Fprintf(os.Stderr, "総置換数: %d件\n", totalReplacements)
	return result
}

func generateOutputFileName(inputFile string) string {
	dir := filepath.Dir(inputFile)
	base := filepath.Base(inputFile)
	ext := filepath.Ext(base)
	nameWithoutExt := strings.TrimSuffix(base, ext)

	outputFileName := nameWithoutExt + "_replaced" + ext
	return filepath.Join(dir, outputFileName)
}

func printResults(matches []Match, config *Config) {
	fmt.Printf("\n=== 抽出結果 ===\n")
	fmt.Printf("総マッチ数: %d\n\n", len(matches))

	patternStats := make(map[string]int)

	for _, match := range matches {
		patternStats[match.PatternName]++
		fmt.Printf("[%s] 行 %d:\n", match.PatternName, match.Line)
		for _, m := range match.Matches {
			fmt.Printf("  → %s\n", m)
		}
		fmt.Println()
	}

	fmt.Println("=== パターン別統計 ===")
	for _, pattern := range config.Patterns {
		if pattern.Pattern != "" {
			count := patternStats[pattern.Name]
			fmt.Printf("%-15s: %d件 (%s)\n", pattern.Name, count, pattern.Description)
		}
	}
}