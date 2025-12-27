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

// GenerateAuthKey 生成 Type-A CDN 鉴权的 sign 参数
// 当 useUID=true 时：
//   - 算法: sign = {timestamp}-{rand}-{uid}-{md5hash}
//   - MD5: md5hash = md5("{uri}-{timestamp}-{rand}-{uid}-{privateKey}")
// 当 useUID=false 时：
//   - 算法: sign = {timestamp}-{rand}-{md5hash}
//   - MD5: md5hash = md5("{uri}-{timestamp}-{rand}-{privateKey}")
// 重要：
//   1. 使用当前 UTC 时间戳（Docker 环境已是 UTC）
//   2. uri 必须使用 URL 编码后的形式（与 CDN 服务器接收到的路径一致）
//   3. useUID 控制是否将 UID 参与签名计算（某些 CDN 配置不需要 UID）
func GenerateAuthKey(uri string, privateKey string, ttl int64, uid string, useUID bool, useRandom bool, randomLength int) string {
	// 步骤1：获取当前时间戳（UTC）
	timestamp := time.Now().Unix()
	timestampStr := strconv.FormatInt(timestamp, 10)

	// 步骤2：生成随机字符串
	randStr := "0"
	if useRandom {
		if randomLength <= 0 {
			randomLength = 6
		}
		randStr = randoms.RandomAlphaNum(randomLength)
	}

	// 步骤3：计算 MD5 签名（根据配置决定是否包含 UID）
	var rawSignStr string
	var signParam string
	if useUID {
		// UID 参与签名计算
		rawSignStr = fmt.Sprintf("%s-%s-%s-%s-%s", uri, timestampStr, randStr, uid, privateKey)
		md5Hash := md5.New()
		md5Hash.Write([]byte(rawSignStr))
		md5hash := hex.EncodeToString(md5Hash.Sum(nil))

		// 步骤4：生成最终签名（包含 UID）
		signParam = fmt.Sprintf("%s-%s-%s-%s", timestampStr, randStr, uid, md5hash)

		// 关键日志：仅输出签名计算的核心信息
		logs.Info("[CDN Auth] uri=%s, ts=%s, rand=%s, uid=%s (useUID=%v), md5=%s",
			uri, timestampStr, randStr, uid, useUID, md5hash)
	} else {
		// UID 不参与签名计算
		rawSignStr = fmt.Sprintf("%s-%s-%s-%s", uri, timestampStr, randStr, privateKey)
		md5Hash := md5.New()
		md5Hash.Write([]byte(rawSignStr))
		md5hash := hex.EncodeToString(md5Hash.Sum(nil))

		// 步骤4：生成最终签名（不包含 UID）
		signParam = fmt.Sprintf("%s-%s-%s", timestampStr, randStr, md5hash)

		// 关键日志：仅输出签名计算的核心信息
		logs.Info("[CDN Auth] uri=%s, ts=%s, rand=%s (useUID=%v), md5=%s",
			uri, timestampStr, randStr, useUID, md5hash)
	}

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
			cfg.CdnAuth.UseUID,
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
