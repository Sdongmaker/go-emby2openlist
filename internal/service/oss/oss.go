package oss

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/AmbitiousJun/go-emby2openlist/v2/internal/config"
	"github.com/AmbitiousJun/go-emby2openlist/v2/internal/util/logs"
	"github.com/AmbitiousJun/go-emby2openlist/v2/internal/util/randoms"
)

// encodePathForCDN 对路径进行编码，保留斜杠
// 关键原则：
// - 空格 → %20（不是 +，路径部分必须用 %20）
// - 中文 → UTF-8 百分号编码
// - 斜杠 / → 保持不变（路径分隔符）
// - 特殊字符 → 百分号编码
func encodePathForCDN(path string) string {
	// 将路径按 / 分割
	segments := strings.Split(path, "/")

	// 对每个片段进行编码（保留 /）
	for i, segment := range segments {
		segments[i] = url.QueryEscape(segment)
	}

	// 重新组合
	encoded := strings.Join(segments, "/")

	// QueryEscape 会将空格编码为 +，需要改为 %20（路径标准）
	encoded = strings.ReplaceAll(encoded, "+", "%20")

	return encoded
}

// GenerateAuthKey 生成腾讯云 CDN Type-A 鉴权的 sign 参数
// 算法: sign = {timestamp}-{rand}-{uid}-{md5hash}
// MD5: md5hash = md5("{uri}-{timestamp}-{rand}-{uid}-{privateKey}")
// 重要：
//   1. 使用当前 UTC 时间戳（Docker 环境已是 UTC）
//   2. uri 必须使用 URL 编码后的形式（与 CDN 服务器接收到的路径一致）
//   3. 连接符固定使用 "-"
//   4. UID 参与签名计算
func GenerateAuthKey(uri string, privateKey string, ttl int64, uid string, useRandom bool, randomLength int) string {
	// 步骤1：获取当前时间戳（UTC）
	timestamp := time.Now().Unix()
	timestampStr := strconv.FormatInt(timestamp, 10)

	logs.Info("[CDN Auth Step 1] 当前时间戳: %s", timestampStr)

	// 步骤2：生成随机字符串
	randStr := "0"
	if useRandom {
		if randomLength <= 0 {
			randomLength = 6
		}
		randStr = randoms.RandomAlphaNum(randomLength)
	}
	logs.Info("[CDN Auth Step 2] 随机字符串: %s (useRandom=%v, length=%d)", randStr, useRandom, randomLength)

	// 步骤3：计算 MD5 签名（腾讯云标准 Type-A 算法）
	logs.Info("[CDN Auth Step 3.1] 开始计算 MD5 签名（腾讯云 Type-A）")
	logs.Info("[CDN Auth Step 3.2] 参数详情 - uri: %s", uri)
	logs.Info("[CDN Auth Step 3.3] 参数详情 - timestamp: %s", timestampStr)
	logs.Info("[CDN Auth Step 3.4] 参数详情 - rand: %s", randStr)
	logs.Info("[CDN Auth Step 3.5] 参数详情 - uid: %s", uid)
	logs.Info("[CDN Auth Step 3.6] 参数详情 - privateKey: %s", privateKey)

	// 腾讯云固定使用 "-" 作为连接符
	rawSignStr := fmt.Sprintf("%s-%s-%s-%s-%s", uri, timestampStr, randStr, uid, privateKey)
	logs.Info("[CDN Auth Step 3.7] 原始签名字符串: %s", rawSignStr)

	md5Hash := md5.New()
	md5Hash.Write([]byte(rawSignStr))
	md5hash := hex.EncodeToString(md5Hash.Sum(nil))
	logs.Info("[CDN Auth Step 3.8] MD5 计算结果（小写）: %s", md5hash)

	// 步骤4：生成最终签名（包含 UID）
	signParam := fmt.Sprintf("%s-%s-%s-%s", timestampStr, randStr, uid, md5hash)
	logs.Info("[CDN Auth Step 4] 最终签名参数: %s", signParam)

	// 签名格式说明
	logs.Info("[CDN Auth] 签名格式: timestamp-rand-uid-md5hash")
	logs.Info("[CDN Auth] 连接符: - (固定), UID: %s", uid)

	return signParam
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

	// 3. 清理路径中的双斜杠
	for strings.Contains(ossPath, "//") {
		ossPath = strings.ReplaceAll(ossPath, "//", "/")
	}

	// 4. 构建签名路径（添加 bucket 前缀）
	signPath := ossPath
	if cfg.Bucket != "" {
		signPath = "/" + cfg.Bucket + ossPath
	}

	// 5. 【关键】先编码路径（用于 MD5 计算和 URL 构建）
	encodedPath := encodePathForCDN(signPath)
	logs.Info("[OSS] Path: %s -> Encoded: %s", signPath, encodedPath)

	// 6. 构建基础 URL（使用编码后的路径）
	baseURL := cfg.Endpoint + encodedPath

	// 7. 如果启用 CDN 鉴权，添加 sign 参数
	if cfg.CdnAuth.Enable {
		authKey := GenerateAuthKey(
			encodedPath, // 使用编码后的路径
			cfg.CdnAuth.PrivateKey,
			cfg.CdnAuth.TTL,
			cfg.CdnAuth.UID,
			cfg.CdnAuth.UseRandom,
			cfg.CdnAuth.RandomLength,
		)
		baseURL += "?sign=" + authKey
	}

	logs.Success("[OSS] URL: %s", baseURL)
	return baseURL, nil
}

// MapPath 是 BuildURL 的辅助方法，仅用于路径映射测试
func MapPath(embyPath string) (string, error) {
	return config.C.Oss.MapPath(embyPath)
}
