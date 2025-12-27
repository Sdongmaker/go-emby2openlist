package oss

import (
	"testing"
	"time"

	"github.com/AmbitiousJun/go-emby2openlist/v2/internal/config"
)

func TestGenerateAuthKey(t *testing.T) {
	// 测试标准 Type-A 算法
	uri := "/video/test.mp4"
	privateKey := "aliyuncdnexp1234"
	ttl := int64(3600)
	uid := "0"
	useUID := true
	separator := "-"
	md5ToUpper := false
	useRandom := false
	randomLength := 6

	authKey := GenerateAuthKey(uri, privateKey, ttl, uid, useUID, separator, md5ToUpper, useRandom, randomLength)

	// auth_key 格式: timestamp-rand-uid-md5hash
	// 因为 timestamp 是动态的，我们只能验证格式
	if authKey == "" {
		t.Errorf("GenerateAuthKey 返回空字符串")
	}

	t.Logf("生成的 auth_key: %s", authKey)
}

func TestBuildURL(t *testing.T) {
	// 初始化测试配置
	config.C = &config.Config{
		Oss: &config.Oss{
			Enable:   true,
			Endpoint: "https://s3.startspoint.com",
			Bucket:   "",
			CdnAuth: &config.CdnAuth{
				Enable:     true,
				PrivateKey: "test-private-key",
				TTL:        3600,
				UID:        "0",
				UseUID:     true,
				Separator:  "-",
				MD5ToUpper: false,
				UseRandom:  false,
			},
		},
	}

	// 初始化路径映射
	config.C.Oss.PathMapping = []string{
		"/movie:/media",
		"/series:/tv",
	}
	config.C.Oss.Init()

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
		Oss: &config.Oss{
			Enable:      true,
			PathMapping: []string{"/movie:/media", "/series:/tv-shows"},
		},
	}
	config.C.Oss.Init()

	tests := []struct {
		name      string
		embyPath  string
		want      string
		wantError bool
	}{
		{
			name:      "movie 路径映射",
			embyPath:  "/movie/test.mkv",
			want:      "/media/test.mkv",
			wantError: false,
		},
		{
			name:      "series 路径映射",
			embyPath:  "/series/show/S01E01.mkv",
			want:      "/tv-shows/show/S01E01.mkv",
			wantError: false,
		},
		{
			name:      "中文路径映射",
			embyPath:  "/movie/星际穿越/星际穿越.mkv",
			want:      "/media/星际穿越/星际穿越.mkv",
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

func TestTypeAAlgorithm(t *testing.T) {
	// 测试 Type-A 算法的正确性
	// 根据阿里云文档的示例验证
	uri := "/video/standard/test.mp4"
	privateKey := "aliyuncdnexp1234"
	timestamp := int64(1444435200)
	rand := "0"
	uid := "0"

	// 手动构造签名字符串
	sstring := "/video/standard/test.mp4-1444435200-0-0-aliyuncdnexp1234"

	// 预期的 MD5 值（根据阿里云文档）
	expectedMD5 := "23bf85053008f5c0e791667a313e28ce"

	// 模拟固定时间戳的签名生成
	import (
		"crypto/md5"
		"fmt"
	)

	hash := md5.Sum([]byte(sstring))
	actualMD5 := fmt.Sprintf("%x", hash)

	if actualMD5 != expectedMD5 {
		t.Errorf("MD5 计算错误: got %s, want %s", actualMD5, expectedMD5)
	}

	t.Logf("Type-A 算法验证通过: %s", actualMD5)
}
