package oss

import (
	"crypto/md5"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/AmbitiousJun/go-emby2openlist/v2/internal/config"
	"github.com/AmbitiousJun/go-emby2openlist/v2/internal/util/logs"
	"github.com/AmbitiousJun/go-emby2openlist/v2/internal/util/randoms"
)

// GenerateAuthKey 生成 Type-A CDN 鉴权的 auth_key
// 算法: auth_key = {timestamp}-{rand}-{uid}-{md5hash}
// 其中: md5hash = md5("{uri}-{timestamp}-{rand}-{uid}-{privateKey}")
func GenerateAuthKey(uri string, privateKey string, ttl int64, uid string, useRandom bool) string {
	// 1. 生成时间戳 (当前时间 + TTL)
	timestamp := time.Now().Unix() + ttl

	// 2. 生成随机字符串 (UUID格式, 无中划线)
	rand := "0"
	if useRandom {
		rand = randoms.RandomHex(32)
	}

	// 3. 构造签名字符串
	// 格式: URI-Timestamp-rand-uid-PrivateKey
	sstring := fmt.Sprintf("%s-%d-%s-%s-%s", uri, timestamp, rand, uid, privateKey)

	// 4. 计算 MD5
	hash := md5.Sum([]byte(sstring))
	md5hash := fmt.Sprintf("%x", hash)

	// 5. 构造 auth_key
	authKey := fmt.Sprintf("%d-%s-%s-%s", timestamp, rand, uid, md5hash)

	logs.Info("Type-A 鉴权生成: uri=%s, timestamp=%d, rand=%s, uid=%s, md5hash=%s", uri, timestamp, rand, uid, md5hash)

	return authKey
}

// BuildURL 根据 Emby 路径构建完整的 OSS URL (带 CDN 鉴权)
func BuildURL(embyPath string) (string, error) {
	cfg := config.C.Oss
	if !cfg.Enable {
		return "", fmt.Errorf("OSS 功能未启用")
	}

	// 1. 映射 Emby 路径到 OSS 路径
	ossPath, err := cfg.MapPath(embyPath)
	if err != nil {
		return "", fmt.Errorf("路径映射失败: %v", err)
	}

	// 2. 确保路径以 / 开头
	if !strings.HasPrefix(ossPath, "/") {
		ossPath = "/" + ossPath
	}

	// 3. 如果配置了 bucket，将其添加到路径前面
	fullPath := ossPath
	if cfg.Bucket != "" {
		fullPath = "/" + cfg.Bucket + ossPath
	}

	// 4. URL 编码路径 (处理中文等特殊字符)
	encodedPath := url.PathEscape(fullPath)
	// 修复: PathEscape 会将 / 也编码，需要还原
	encodedPath = strings.ReplaceAll(encodedPath, "%2F", "/")

	// 5. 构建基础 URL
	baseURL := cfg.Endpoint + encodedPath

	// 6. 如果启用 CDN 鉴权，添加 auth_key 参数
	if cfg.CdnAuth.Enable {
		authKey := GenerateAuthKey(
			fullPath, // 注意: 签名使用完整路径(含bucket)，不是编码后的
			cfg.CdnAuth.PrivateKey,
			cfg.CdnAuth.TTL,
			cfg.CdnAuth.UID,
			cfg.CdnAuth.UseRandom,
		)
		baseURL += "?auth_key=" + authKey
	}

	logs.Success("OSS URL 生成成功: %s", baseURL)
	return baseURL, nil
}

// MapPath 是 BuildURL 的辅助方法，仅用于路径映射测试
func MapPath(embyPath string) (string, error) {
	return config.C.Oss.MapPath(embyPath)
}
