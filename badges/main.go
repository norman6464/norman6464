package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	AtCoder struct {
		Username string `yaml:"username"`
	} `yaml:"atcoder"`
	Paiza struct {
		Rank string `yaml:"rank"`
	} `yaml:"paiza"`
}

type AtCoderContest struct {
	NewRating int `json:"NewRating"`
}

type AtCoderData struct {
	Username string
	Rating   int
	Rank     string
	Color    string
}

type PaizaData struct {
	Rank  string
	Color string
}

type BadgeData struct {
	AtCoder   AtCoderData
	Paiza     PaizaData
	UpdatedAt string
}

func main() {
	configPath := getConfigPath()
	config, err := loadConfig(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "設定ファイルの読み込みに失敗しました: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("=== 競技プログラミングランク バッジ生成ツール ===")

	// AtCoderデータ取得
	fmt.Printf("AtCoderユーザー '%s' のデータを取得中...\n", config.AtCoder.Username)
	atcoderData, err := fetchAtCoderData(config.AtCoder.Username)
	if err != nil {
		fmt.Fprintf(os.Stderr, "AtCoderデータの取得に失敗しました: %v\n", err)
		fmt.Println("デフォルト値を使用します")
		atcoderData = AtCoderData{
			Username: config.AtCoder.Username,
			Rating:   0,
			Rank:     "Unrated",
			Color:    "#808080",
		}
	}
	fmt.Printf("  レーティング: %d (%s)\n", atcoderData.Rating, atcoderData.Rank)

	// paizaデータ設定
	paizaData := getPaizaData(config.Paiza.Rank)
	fmt.Printf("paizaランク: %s\n", paizaData.Rank)

	// バッジ生成
	badgeData := BadgeData{
		AtCoder:   atcoderData,
		Paiza:     paizaData,
		UpdatedAt: time.Now().Format("2006-01-02"),
	}

	outputDir := getOutputDir()
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "出力ディレクトリの作成に失敗しました: %v\n", err)
		os.Exit(1)
	}

	if err := generateAtCoderBadge(badgeData, outputDir); err != nil {
		fmt.Fprintf(os.Stderr, "AtCoderバッジの生成に失敗しました: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("AtCoderバッジを生成しました")

	if err := generatePaizaBadge(badgeData, outputDir); err != nil {
		fmt.Fprintf(os.Stderr, "paizaバッジの生成に失敗しました: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("paizaバッジを生成しました")

	fmt.Println("=== バッジ生成完了 ===")
}

func getConfigPath() string {
	if p := os.Getenv("CONFIG_PATH"); p != "" {
		return p
	}
	return "config.yaml"
}

func getOutputDir() string {
	if p := os.Getenv("OUTPUT_DIR"); p != "" {
		return p
	}
	return "output"
}

func loadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("ファイル読み込みエラー: %w", err)
	}
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("YAML解析エラー: %w", err)
	}
	return &config, nil
}

func fetchAtCoderData(username string) (AtCoderData, error) {
	url := fmt.Sprintf("https://atcoder.jp/users/%s/history/json", username)
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return AtCoderData{}, fmt.Errorf("HTTPリクエスト失敗: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return AtCoderData{}, fmt.Errorf("HTTPステータス: %d", resp.StatusCode)
	}

	var contests []AtCoderContest
	if err := json.NewDecoder(resp.Body).Decode(&contests); err != nil {
		return AtCoderData{}, fmt.Errorf("JSONデコードエラー: %w", err)
	}

	rating := 0
	if len(contests) > 0 {
		rating = contests[len(contests)-1].NewRating
	}

	rank, color := getAtCoderRankAndColor(rating)

	return AtCoderData{
		Username: username,
		Rating:   rating,
		Rank:     rank,
		Color:    color,
	}, nil
}

func getAtCoderRankAndColor(rating int) (string, string) {
	switch {
	case rating >= 2800:
		return "赤", "#FF0000"
	case rating >= 2400:
		return "橙", "#FF8000"
	case rating >= 2000:
		return "黄", "#C0C000"
	case rating >= 1600:
		return "青", "#0000FF"
	case rating >= 1200:
		return "水", "#00C0C0"
	case rating >= 800:
		return "緑", "#008000"
	case rating >= 400:
		return "茶", "#804000"
	case rating > 0:
		return "灰", "#808080"
	default:
		return "Unrated", "#000000"
	}
}

func getPaizaData(rank string) PaizaData {
	rank = strings.ToUpper(strings.TrimSpace(rank))
	colorMap := map[string]string{
		"S":  "#1a73e8",
		"A":  "#e53935",
		"B":  "#43a047",
		"C":  "#fb8c00",
		"D":  "#8e24aa",
		"E":  "#757575",
	}
	color, ok := colorMap[rank]
	if !ok {
		color = "#757575"
	}
	return PaizaData{Rank: rank, Color: color}
}

func generateAtCoderBadge(data BadgeData, outputDir string) error {
	tmpl := `<svg xmlns="http://www.w3.org/2000/svg" width="320" height="120" viewBox="0 0 320 120">
  <defs>
    <linearGradient id="bg" x1="0%" y1="0%" x2="100%" y2="100%">
      <stop offset="0%" style="stop-color:#1a1a2e;stop-opacity:1" />
      <stop offset="100%" style="stop-color:#16213e;stop-opacity:1" />
    </linearGradient>
    <linearGradient id="accent" x1="0%" y1="0%" x2="100%" y2="0%">
      <stop offset="0%" style="stop-color:{{.AtCoder.Color}};stop-opacity:1" />
      <stop offset="100%" style="stop-color:{{.AtCoder.Color}};stop-opacity:0.6" />
    </linearGradient>
  </defs>
  <rect width="320" height="120" rx="10" fill="url(#bg)" />
  <rect x="0" y="0" width="6" height="120" rx="3" fill="url(#accent)" />
  <text x="24" y="32" font-family="Segoe UI, Arial, sans-serif" font-size="13" fill="#8892b0" font-weight="600">AtCoder</text>
  <text x="24" y="58" font-family="Segoe UI, Arial, sans-serif" font-size="16" fill="#ccd6f6" font-weight="700">{{.AtCoder.Username}}</text>
  <text x="24" y="86" font-family="Segoe UI, Arial, sans-serif" font-size="28" fill="{{.AtCoder.Color}}" font-weight="800">{{.AtCoder.Rating}}</text>
  <rect x="130" y="66" width="70" height="28" rx="14" fill="{{.AtCoder.Color}}" opacity="0.15" />
  <text x="165" y="85" font-family="Segoe UI, Arial, sans-serif" font-size="13" fill="{{.AtCoder.Color}}" font-weight="700" text-anchor="middle">{{.AtCoder.Rank}}</text>
  <text x="24" y="108" font-family="Segoe UI, Arial, sans-serif" font-size="10" fill="#495670">更新日: {{.UpdatedAt}}</text>
</svg>`

	t, err := template.New("atcoder").Parse(tmpl)
	if err != nil {
		return fmt.Errorf("テンプレート解析エラー: %w", err)
	}

	outputPath := filepath.Join(outputDir, "atcoder-badge.svg")
	f, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("ファイル作成エラー: %w", err)
	}
	defer f.Close()

	return t.Execute(f, data)
}

func generatePaizaBadge(data BadgeData, outputDir string) error {
	tmpl := `<svg xmlns="http://www.w3.org/2000/svg" width="320" height="120" viewBox="0 0 320 120">
  <defs>
    <linearGradient id="bg-paiza" x1="0%" y1="0%" x2="100%" y2="100%">
      <stop offset="0%" style="stop-color:#1a1a2e;stop-opacity:1" />
      <stop offset="100%" style="stop-color:#16213e;stop-opacity:1" />
    </linearGradient>
    <linearGradient id="accent-paiza" x1="0%" y1="0%" x2="100%" y2="0%">
      <stop offset="0%" style="stop-color:{{.Paiza.Color}};stop-opacity:1" />
      <stop offset="100%" style="stop-color:{{.Paiza.Color}};stop-opacity:0.6" />
    </linearGradient>
  </defs>
  <rect width="320" height="120" rx="10" fill="url(#bg-paiza)" />
  <rect x="0" y="0" width="6" height="120" rx="3" fill="url(#accent-paiza)" />
  <text x="24" y="32" font-family="Segoe UI, Arial, sans-serif" font-size="13" fill="#8892b0" font-weight="600">paiza</text>
  <text x="24" y="58" font-family="Segoe UI, Arial, sans-serif" font-size="16" fill="#ccd6f6" font-weight="700">スキルチェック</text>
  <text x="24" y="90" font-family="Segoe UI, Arial, sans-serif" font-size="36" fill="{{.Paiza.Color}}" font-weight="800">{{.Paiza.Rank}}ランク</text>
  <text x="24" y="108" font-family="Segoe UI, Arial, sans-serif" font-size="10" fill="#495670">更新日: {{.UpdatedAt}}</text>
</svg>`

	t, err := template.New("paiza").Parse(tmpl)
	if err != nil {
		return fmt.Errorf("テンプレート解析エラー: %w", err)
	}

	outputPath := filepath.Join(outputDir, "paiza-badge.svg")
	f, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("ファイル作成エラー: %w", err)
	}
	defer f.Close()

	return t.Execute(f, data)
}
