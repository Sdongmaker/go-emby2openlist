package goedge

import (
	"testing"

	"github.com/AmbitiousJun/go-emby2openlist/v2/internal/config"
)

func TestGenerateAuthSign(t *testing.T) {
	// 测试 GoEdge 鉴权算法
	path := "/images/test.jpg"
	privateKey := "123456"
	ttl := int64(3600)
	useRandom := false
	randomLength := 16

	authSign := GenerateAuthSign(path, privateKey, ttl, useRandom, randomLength)

	// sign 格式: timestamp-rand-md5hash
	// 因为 timestamp 是动态的，我们只能验证格式
	if authSign == "" {
		t.Errorf("GenerateAuthSign 返回空字符串")
	}

	t.Logf("生成的 sign: %s", authSign)
}

func TestBuildURL(t *testing.T) {
	// 初始化测试配置
	config.C = &config.Config{
		GoEdge: &config.GoEdge{
			Enable:   true,
			Endpoint: "https://example.com",
			Auth: &config.GoEdgeAuth{
				Enable:       true,
				PrivateKey:   "123456",
				TTL:          3600,
				UseRandom:    true,
				RandomLength: 16,
			},
		},
	}

	// 初始化路径映射
	config.C.GoEdge.PathMapping = []string{
		"/movie:/images",
		"/series:/videos",
	}
	config.C.GoEdge.Init()

	tests := []struct {
		name      string
		embyPath  string
		wantError bool
	}{
		{
			name:      "普通英文路径",
			embyPath:  "/movie/test/video.mkv",
			wantError: false,
		},
		{
			name:      "中文路径",
			embyPath:  "/movie/星际穿越 (2014)/星际穿越 (2014) - 2160p.mkv",
			wantError: false,
		},
		{
			name:      "无效路径映射",
			embyPath:  "/invalid/path/video.mkv",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url, err := BuildURL(tt.embyPath)
			if (err != nil) != tt.wantError {
				t.Errorf("BuildURL() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if !tt.wantError {
				t.Logf("生成的 URL: %s", url)
			}
		})
	}
}

func TestMapPath(t *testing.T) {
	// 初始化测试配置
	config.C = &config.Config{
		GoEdge: &config.GoEdge{
			Enable:      true,
			PathMapping: []string{"/movie:/images", "/series:/videos"},
		},
	}
	config.C.GoEdge.Init()

	tests := []struct {
		name      string
		embyPath  string
		want      string
		wantError bool
	}{
		{
			name:      "movie 路径映射",
			embyPath:  "/movie/test.mkv",
			want:      "/images/test.mkv",
			wantError: false,
		},
		{
			name:      "series 路径映射",
			embyPath:  "/series/show/S01E01.mkv",
			want:      "/videos/show/S01E01.mkv",
			wantError: false,
		},
		{
			name:      "中文路径映射",
			embyPath:  "/movie/星际穿越/星际穿越.mkv",
			want:      "/images/星际穿越/星际穿越.mkv",
			wantError: false,
		},
		{
			name:      "无匹配前缀",
			embyPath:  "/music/song.mp3",
			want:      "",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := MapPath(tt.embyPath)
			if (err != nil) != tt.wantError {
				t.Errorf("MapPath() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if !tt.wantError && got != tt.want {
				t.Errorf("MapPath() = %v, want %v", got, tt.want)
			}
		})
	}
}
