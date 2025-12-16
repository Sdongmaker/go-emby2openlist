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
// 完全按照腾讯云 Type-A 标准实现
// 注意：timestamp 是当前时间，不是过期时间！CDN 会自动加上有效时长判断
// randomLength: 随机数长度（腾讯云建议 6 位，阿里云建议 32 位）
func GenerateAuthKey(uri string, privateKey string, ttl int64, uid string, useRandom bool, randomLength int) string {
	// 1. 生成时间戳（当前时间，不是过期时间！）
	// 腾讯云文档：timestamp = 生成签名的时间
	// CDN 验证逻辑：timestamp + 控制台配置的有效时长 > 当前时间
	//
	// 临时方案：减去 8 小时 (28800 秒)
	// 原因：虽然 Docker 设置了 TZ=UTC，但时间戳仍然快了 8 小时
	timestamp := time.Now().Unix() - 28800

	// 2. 生成随机字符串
	// Python: rand_str = "0"
	// 腾讯云: 6位字母数字混合 (如 "q87NIR")
	// 阿里云: 32位十六进制 (如 "52df74f89e7ed1369ffbc0204fd1f9bc")
	rand := "0"
	if useRandom {
		if randomLength <= 0 {
			randomLength = 6 // 默认6位（腾讯云）
		}
		// 使用字母数字混合的随机数（适配腾讯云）
		rand = randoms.RandomAlphaNum(randomLength)
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

	// 详细调试日志
	logs.Info("[Type A] path: %s", uri)
	logs.Info("[Type A] timestamp: %d (当前时间，有效期: %d秒)", timestamp, ttl)
	logs.Info("[Type A] expire_time: %d (timestamp + ttl)", timestamp+ttl)
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

	// 2.5. 清理路径中的双斜杠（修复路径映射导致的问题）
	// 将连续的多个 / 替换为单个 /
	for strings.Contains(ossPath, "//") {
		ossPath = strings.ReplaceAll(ossPath, "//", "/")
	}

	// 3. 先解码路径（防止路径已经是部分或完全编码的状态）
	// 这确保我们得到的是原始的、未编码的路径
	decodedOssPath, err := url.QueryUnescape(ossPath)
	if err != nil {
		logs.Warn("路径解码失败，使用原始路径: %v", err)
		decodedOssPath = ossPath
	}

	// 4. 构建用于签名的路径（完全未编码的原始路径）
	// 注意：腾讯云 CDN 会将收到的编码 URI 解码后再验证签名
	signPath := decodedOssPath
	if cfg.Bucket != "" {
		// 如果配置了 bucket，将其添加到路径前面
		signPath = "/" + cfg.Bucket + decodedOssPath
	}

	// 5. 构建用于 URL 的路径（完整编码所有特殊字符）
	// URL 编码路径 (处理中文、空格、括号等所有特殊字符)
	encodedPath := url.PathEscape(signPath)
	// 修复: PathEscape 会将 / 也编码，需要还原
	encodedPath = strings.ReplaceAll(encodedPath, "%2F", "/")

	// 6. 构建基础 URL
	baseURL := cfg.Endpoint + encodedPath

	// 7. 如果启用 CDN 鉴权，添加 sign 参数
	if cfg.CdnAuth.Enable {
		// 重要：签名使用完全解码的原始路径（signPath），而不是编码后的路径
		// 因为腾讯云 CDN 会先解码收到的 URI，再用解码后的路径验证签名
		logs.Info("[Type A] 原始路径: %s", ossPath)
		logs.Info("[Type A] 解码后路径: %s", decodedOssPath)
		logs.Info("[Type A] 签名路径 (未编码): %s", signPath)
		logs.Info("[Type A] URL路径 (完整编码): %s", encodedPath)

		authKey := GenerateAuthKey(
			signPath, // 签名使用未编码的路径
			cfg.CdnAuth.PrivateKey,
			cfg.CdnAuth.TTL,
			cfg.CdnAuth.UID,
			cfg.CdnAuth.UseRandom,
			cfg.CdnAuth.RandomLength, // 随机数长度
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
