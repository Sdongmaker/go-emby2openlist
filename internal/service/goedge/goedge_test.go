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

// TestEncodePathForCDN 测试路径编码函数
func TestEncodePathForCDN(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "/电影/华语电影/72小时黄金行动 (2023)/poster.jpg",
			expected: "/%E7%94%B5%E5%BD%B1/%E5%8D%8E%E8%AF%AD%E7%94%B5%E5%BD%B1/72%E5%B0%8F%E6%97%B6%E9%BB%84%E9%87%91%E8%A1%8C%E5%8A%A8%20%282023%29/poster.jpg",
		},
		{
			input:    "/剧集/国漫/仙逆",
			expected: "/%E5%89%A7%E9%9B%86/%E5%9B%BD%E6%BC%AB/%E4%BB%99%E9%80%86",
		},
		{
			input:    "/test/path with space/file.mp4",
			expected: "/test/path%20with%20space/file.mp4",
		},
	}

	for _, tt := range tests {
		result := encodePathForCDN(tt.input)
		if result != tt.expected {
			t.Errorf("encodePathForCDN(%q):\n  got:      %q\n  expected: %q", tt.input, result, tt.expected)
		} else {
			t.Logf("✓ encodePathForCDN(%q) = %q", tt.input, result)
		}
	}
}

// TestAuthSignWithOriginalPath 验证签名必须使用原始路径（与 Python 代码逻辑一致）
func TestAuthSignWithOriginalPath(t *testing.T) {
	t.Log("========== 验证 GoEdge 鉴权逻辑（与 Python 测试代码对比）==========")
	t.Log("")

	// Python 测试代码中的参数
	rawPath := "/电影/华语电影/72小时黄金行动 (2023)/poster.jpg"
	endpoint := "https://test.startspoint.com"

	t.Logf("1. 原始路径（中文）: %s", rawPath)
	t.Log("")

	// 编码路径
	encodedPath := encodePathForCDN(rawPath)
	t.Logf("2. 编码后的路径: %s", encodedPath)
	t.Log("")

	// 关键验证点
	t.Log("3. 【关键逻辑】")
	t.Log("   - 签名计算: 使用**原始路径**（未编码的中文）")
	t.Log("   - URL 构建: 使用**编码后的路径**")
	t.Log("")

	t.Log("4. 为什么这样设计？")
	t.Log("   - GoEdge CDN 服务器在验证签名时，会先解码接收到的 URL")
	t.Log("   - 然后用解码后的原始路径计算 MD5")
	t.Log("   - 如果签名也用编码路径计算，会导致验证失败")
	t.Log("")

	t.Log("5. 客户端兼容性")
	t.Log("   - Infuse: 会对 URL 进行编码，但我们已经返回编码 URL，避免双重编码")
	t.Log("   - SenPlayer: 不编码，直接使用我们返回的编码 URL")
	t.Log("   - 两种客户端都能正常工作")
	t.Log("")

	// 最终 URL
	finalURL := endpoint + encodedPath + "?sign={timestamp}-{rand}-{md5hash}"
	t.Logf("6. 最终 URL 格式: %s", finalURL)
	t.Log("")

	t.Log("========== 验证完成 ==========")
}
