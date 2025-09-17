# 正規表現抽出・置換ツール

YAMLで定義された正規表現パターンを使用してファイルから文字列を抽出・置換するGoアプリケーションです。HTMLファイルの不要なタグ削除、テキストファイルの一括置換、ログファイルのパターン抽出など、様々なテキスト処理タスクに活用できます。

## このアプリケーションでできること

### 主な用途
- **HTMLクリーニング**: 不要なHTMLタグや特定のスタイル属性を持つ要素を一括削除
- **テキスト抽出**: ログファイルやWebページから特定パターンの情報を抽出
- **一括置換**: 複数の正規表現パターンを使った高度な文字列置換
- **フォーマット変換**: 特定の形式から別の形式への変換処理
- **コード整形**: ソースコードから特定のコメントやパターンを削除・置換

### 具体的な活用例
1. **Webスクレイピング後のHTMLクリーニング**
   - 不要な広告タグの削除
   - 特定のCSSスタイルを持つ要素の除去
   - JavaScriptコードの削除

2. **ログファイル解析**
   - エラーメッセージの抽出
   - 特定のIPアドレスやURLのパターン抽出
   - タイムスタンプ形式の統一

3. **ドキュメント処理**
   - マークダウンからHTMLタグの削除
   - 特定のフォーマットへの変換
   - 不要なメタデータの削除

## 機能

- **抽出モード**: 正規表現パターンにマッチする文字列を抽出・表示
- **置換モード**: パターンにマッチした部分を指定した文字列に置換してファイル保存
- **複数行対応**: 改行をまたぐHTMLタグなどの複雑なパターンに対応
- **統計表示**: パターン別のマッチ数や置換数を表示
- **自動ファイル保存**: 置換結果を`元ファイル名_replaced.拡張子`として自動保存
- **複数パターン処理**: 一度の実行で複数の正規表現パターンを順次適用
- **柔軟な設定**: YAML形式で簡単にパターンを追加・編集可能

## 必要な環境

- Go 1.21以上
- gopkg.in/yaml.v2パッケージ

## インストール

```bash
git clone <repository-url>
cd remove_tag
go mod tidy
```

## 使用方法

### 基本的な使用方法

```bash
# 抽出モード（デフォルト）
go run main.go <入力ファイルパス> [設定ファイルパス]

# 置換モード
go run main.go <入力ファイルパス> [設定ファイルパス] --replace
```

### コマンドライン例

```bash
# デフォルト設定で抽出（マッチした内容を表示）
go run main.go input.txt

# カスタム設定ファイルで抽出
go run main.go input.txt custom_config.yaml

# 置換実行（自動で_replaced.txtファイルを生成）
go run main.go input.txt --replace

# 短縮オプションで置換
go run main.go input.txt -r

# HTMLファイルのクリーニング
go run main.go webpage.html html_clean.yaml --replace

# ログファイルからエラー抽出
go run main.go app.log error_patterns.yaml
```

### オプション

- `--replace`, `-r`: 置換モードで実行（ファイルを書き換えて保存）
- 設定ファイルが指定されない場合は`config.yaml`を使用

### 処理フロー

1. **設定ファイル読み込み**: YAMLファイルから正規表現パターンを読み込み
2. **入力ファイル読み込み**: 処理対象のテキストファイルを読み込み
3. **パターン適用**: 各パターンを順番に適用
4. **結果出力**:
   - 抽出モード: マッチした内容を画面に表示
   - 置換モード: 置換後の内容を新しいファイルに保存

## 設定ファイル（config.yaml）

設定ファイルはYAML形式で正規表現パターンを定義します。用途に応じて複数の設定ファイルを作成し、使い分けることができます。

### 基本的な設定例

```yaml
# 正規表現パターン設定ファイル
patterns:
  - name: "URL抽出"
    pattern: 'https?://[^\s<>"{}|\\^`\[\]]+'
    description: "URLを抽出"
    replacement: "[URL削除]"

  - name: "HTMLタグ削除"
    pattern: '<script[^>]*>.*?</script>'
    description: "scriptタグとその内容を削除"
    replacement: ""

  - name: "空白行削除"
    pattern: '^\s*$\n'
    description: "空白行を削除"
    replacement: ""
```

### 実用的な設定例

#### HTMLクリーニング用 (html_clean.yaml)

```yaml
patterns:
  - name: "広告削除"
    pattern: '<div[^>]*class="[^"]*ad[^"]*"[^>]*>.*?</div>'
    description: "広告関連のdivタグを削除"
    replacement: ""

  - name: "スタイル削除"
    pattern: 'style="[^"]*"'
    description: "インラインスタイルを削除"
    replacement: ""

  - name: "スクリプト削除"
    pattern: '<script[^>]*>.*?</script>'
    description: "JavaScriptを削除"
    replacement: ""
```

#### ログ解析用 (log_patterns.yaml)

```yaml
patterns:
  - name: "エラー抽出"
    pattern: '.*ERROR.*'
    description: "エラー行を抽出"
    replacement: ""

  - name: "IP抽出"
    pattern: '\b(?:\d{1,3}\.){3}\d{1,3}\b'
    description: "IPアドレスを抽出"
    replacement: ""

  - name: "タイムスタンプ"
    pattern: '\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}'
    description: "日時を抽出"
    replacement: ""
```

### 設定項目

- `name`: パターンの名前（識別用、ログ出力で使用）
- `pattern`: 正規表現パターン（Goのregexpパッケージ準拠）
- `description`: パターンの説明（統計表示で使用）
- `replacement`: 置換文字列（抽出モードでは無視される）

### 置換文字列の指定方法

- **削除**: `replacement: ""`（空文字列で完全削除）
- **コメント化**: `replacement: "<!-- 削除されました -->"`
- **プレースホルダー**: `replacement: "[削除]"`
- **別のタグに置換**: `replacement: '<div class="new">新内容</div>'`
- **テキスト置換**: `replacement: "置換後のテキスト"`

## 実行例

### 抽出モード（パターンマッチング確認）

```bash
$ go run main.go webpage.html

=== 抽出結果 ===
総マッチ数: 15

[スクリプト削除] 行 12:
  → <script src="analytics.js"></script>

[広告削除] 行 45:
  → <div class="ad-banner">広告コンテンツ</div>

[URL抽出] 行 67:
  → https://example.com/api/v1/data

=== パターン別統計 ===
スクリプト削除     : 3件 (JavaScriptを削除)
広告削除         : 5件 (広告関連のdivタグを削除)
URL抽出         : 7件 (URLを抽出)
```

### 置換モード（ファイル処理）

```bash
$ go run main.go webpage.html html_clean.yaml --replace

[スクリプト削除] 3件置換しました
[広告削除] 5件置換しました
[スタイル削除] 42件置換しました
総置換数: 50件
置換結果を保存しました: webpage_replaced.html
```

### ログファイル解析例

```bash
$ go run main.go application.log log_patterns.yaml

=== 抽出結果 ===
総マッチ数: 156

[エラー抽出] 行 234:
  → 2024-01-15 10:23:45 ERROR: Database connection failed

[IP抽出] 行 567:
  → 192.168.1.105

=== パターン別統計 ===
エラー抽出       : 23件 (エラー行を抽出)
IP抽出         : 89件 (IPアドレスを抽出)
タイムスタンプ    : 44件 (日時を抽出)
```

## ファイル構成

```
remove_tag/
├── main.go              # メインアプリケーション
├── config.yaml          # デフォルト設定ファイル
├── go.mod              # Go依存関係管理
├── README.md           # このファイル
└── samples/            # サンプルファイル（任意）
    ├── test.txt
    └── sample.txt
```

## 技術仕様

### 正規表現エンジン

- Goの標準`regexp`パッケージを使用
- PCRE互換の正規表現をサポート
- 複数行マッチング対応（`(?s)`フラグ自動付与）

### ファイル処理

- **入力**: UTF-8エンコーディングのテキストファイル
- **出力**: 同じエンコーディングで保存
- **権限**: 出力ファイルは644権限で作成

### パフォーマンス

- ファイル全体をメモリに読み込み
- 複数パターンの順次処理
- 大きなファイル（数十MB）でも処理可能

## トラブルシューティング

### よくある問題

1. **YAML解析エラー**
   ```
   YAML解析エラー: yaml: line X: found character that cannot start any token
   ```
   → config.yamlの文法を確認してください（引用符のエスケープなど）

2. **正規表現エラー**
   ```
   正規表現エラー ('パターン名'): error parsing regexp: ...
   ```
   → 正規表現パターンの文法を確認してください

3. **ファイル読み込みエラー**
   ```
   ファイルの読み込みエラー: open filename: no such file or directory
   ```
   → ファイルパスが正しいか確認してください

### デバッグ方法

1. **設定ファイルの確認**
   ```bash
   # YAMLファイルの検証
   python3 -c "import yaml; yaml.safe_load(open('config.yaml'))"
   ```

2. **小さなファイルでテスト**
   ```bash
   # 小さなサンプルファイルで動作確認
   echo "test content" > test.txt
   go run main.go test.txt
   ```

## 開発情報

### ディレクトリ構造

```go
type Pattern struct {
    Name        string `yaml:"name"`
    Pattern     string `yaml:"pattern"`
    Description string `yaml:"description"`
    Replacement string `yaml:"replacement"`
}

type Config struct {
    Patterns []Pattern `yaml:"patterns"`
}
```

### 主要な関数

- `main()`: エントリーポイント、引数解析
- `loadConfig()`: YAML設定ファイルの読み込み
- `performReplacements()`: 置換処理の実行
- `generateOutputFileName()`: 出力ファイル名の生成
- `printResults()`: 抽出結果の表示

## ライセンス

このプロジェクトはMITライセンスの下で公開されています。

## 更新履歴

- v1.0.0: 初期リリース（抽出機能）
- v1.1.0: 置換機能追加
- v1.2.0: 自動ファイル保存機能追加
- v1.3.0: 複数行パターン対応強化

## 貢献

バグ報告や機能要求は、GitHubのIssuesでお知らせください。
プルリクエストも歓迎します。