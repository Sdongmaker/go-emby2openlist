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

// GenerateAuthKey 生成 Type-A CDN 鉴权的 sign 参数
// 算法: sign = {timestamp}-{rand}-{uid}-{md5hash}
// 其中: md5hash = md5("{uri}-{timestamp}-{rand}-{uid}-{privateKey}")
// 完全按照 Python Reqable 脚本的实现逻辑
func GenerateAuthKey(uri string, privateKey string, ttl int64, uid string, useRandom bool) string {
	// 1. 生成时间戳 (当前时间 + TTL)
	// Python: expire_time = now_ts + ttl
	timestamp := time.Now().Unix() + ttl

	// 2. 生成随机字符串
	// Python: rand_str = "0"
	rand := "0"
	if useRandom {
		rand = randoms.RandomHex(32)
	}

	// 3. 构造签名字符串
	// Python: raw_str = f"{path}-{expire_time}-{rand_str}-{uid}-{secret_key}"
	// 格式: path-time-rand-uid-key
	sstring := fmt.Sprintf("%s-%d-%s-%s-%s", uri, timestamp, rand, uid, privateKey)

	// 4. 计算 MD5
	// Python: md5_signature = hashlib.md5(raw_str.encode('utf-8')).hexdigest()
	hash := md5.Sum([]byte(sstring))
	md5hash := fmt.Sprintf("%x", hash)

	// 5. 构造最终参数值
	// Python: auth_value = f"{expire_time}-{rand_str}-{uid}-{md5_signature}"
	// 格式: time-rand-uid-md5hash
	authKey := fmt.Sprintf("%d-%s-%s-%s", timestamp, rand, uid, md5hash)

	// 详细调试日志（与 Python 脚本输出格式一致）
	logs.Info("[Type A] path: %s", uri)
	logs.Info("[Type A] raw_str: %s", sstring)
	logs.Info("[Type A] md5: %s", md5hash)
	logs.Info("[Type A] sign: %s", authKey)

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

	// 3. 构建用于签名的路径（未编码）
	// 注意：这个路径必须与 CDN 实际接收到的请求路径完全一致
	signPath := ossPath
	if cfg.Bucket != "" {
		// 如果配置了 bucket，将其添加到路径前面
		signPath = "/" + cfg.Bucket + ossPath
	}

	// 4. 构建用于 URL 的路径（需要编码特殊字符）
	// URL 编码路径 (处理中文等特殊字符)
	encodedPath := url.PathEscape(signPath)
	// 修复: PathEscape 会将 / 也编码，需要还原
	encodedPath = strings.ReplaceAll(encodedPath, "%2F", "/")

	// 5. 构建基础 URL
	baseURL := cfg.Endpoint + encodedPath

	// 6. 如果启用 CDN 鉴权，添加 sign 参数
	if cfg.CdnAuth.Enable {
		// 重要：签名使用未编码的路径（signPath），而不是编码后的路径
		// 这与 Python 脚本中的 request.path 保持一致（纯路径，不包含 query）
		logs.Info("[Type A] 签名路径 (未编码): %s", signPath)
		logs.Info("[Type A] URL路径 (已编码): %s", encodedPath)

		authKey := GenerateAuthKey(
			signPath, // 签名使用未编码的路径
			cfg.CdnAuth.PrivateKey,
			cfg.CdnAuth.TTL,
			cfg.CdnAuth.UID,
			cfg.CdnAuth.UseRandom,
		)
		baseURL += "?sign=" + authKey
	}

	logs.Success("OSS URL 生成成功: %s", baseURL)
	return baseURL, nil
}

// MapPath 是 BuildURL 的辅助方法，仅用于路径映射测试
func MapPath(embyPath string) (string, error) {
	return config.C.Oss.MapPath(embyPath)
}
