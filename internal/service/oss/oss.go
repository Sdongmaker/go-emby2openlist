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
// 算法: sign = {timestamp}-{rand}-{uid}-{md5hash}
// 其中: md5hash = md5("{uri}-{timestamp}-{rand}-{uid}-{privateKey}")
// 重要：
//   1. 根据腾讯云官方 Demo，使用当前时间，不是过期时间
//   2. uri 必须使用 URL 编码后的形式（与 CDN 服务器接收到的路径一致）
// 官方代码: ts = now (当前时间)
// randomLength: 随机数长度（腾讯云建议 6 位，阿里云建议 32 位）
func GenerateAuthKey(uri string, privateKey string, ttl int64, uid string, useRandom bool, randomLength int) string {
	// ==================== 步骤1：准备参数 ====================
	// 严格按照测试代码的方式
	timestamp := time.Now().Unix()
	timestampStr := strconv.FormatInt(timestamp, 10)  // 使用 strconv.FormatInt（与测试代码一致）

	// 时间戳诊断信息
	logs.Info("========== 时间戳诊断 ==========")
	logs.Info("[当前时间戳] %s", timestampStr)
	logs.Info("[当前时间] %s", time.Unix(timestamp, 0).Format("2006-01-02 15:04:05"))
	logs.Info("[TTL] %d 秒", ttl)
	logs.Info("=================================")

	// ==================== 步骤2：生成随机字符串 ====================
	randStr := "0"
	if useRandom {
		if randomLength <= 0 {
			randomLength = 6 // 默认6位
		}
		randStr = randoms.RandomAlphaNum(randomLength)
	}

	// ==================== 步骤3：计算签名 (使用编码后的路径) ====================
	// 严格按照测试代码的 calculateMD5Hash 函数
	// 格式: uri-timestampStr-randStr-uid-pkey
	rawSignStr := fmt.Sprintf("%s-%s-%s-%s-%s", uri, timestampStr, randStr, uid, privateKey)

	// 详细输出
	logs.Info("========== Type-A 签名计算详情 ==========")
	logs.Info("[1] uri (已编码): %s", uri)
	logs.Info("[2] timestamp: %s", timestampStr)
	logs.Info("[3] rand: %s", randStr)
	logs.Info("[4] uid: %s", uid)
	logs.Info("[5] pkey: %s", privateKey)
	logs.Info("----------------------------------------")
	logs.Info("[拼接格式] uri-timestamp-rand-uid-pkey")
	logs.Info("[原始字符串] %s", rawSignStr)

	// 计算 MD5（严格按照测试代码方式）
	md5Hash := md5.New()
	md5Hash.Write([]byte(rawSignStr))
	md5hash := hex.EncodeToString(md5Hash.Sum(nil))  // 使用 hex.EncodeToString（与测试代码一致）
	logs.Info("[MD5 结果] %s", md5hash)

	// ==================== 步骤4：生成签名参数 ====================
	// 格式: timestamp-rand-uid-md5hash
	signParam := fmt.Sprintf("%s-%s-%s-%s", timestampStr, randStr, uid, md5hash)
	logs.Info("----------------------------------------")
	logs.Info("[最终签名] sign=%s", signParam)
	logs.Info("==========================================")

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

	// 2.5. 清理路径中的双斜杠（修复路径映射导致的问题）
	// 将连续的多个 / 替换为单个 /
	for strings.Contains(ossPath, "//") {
		ossPath = strings.ReplaceAll(ossPath, "//", "/")
	}

	// 3. 构建签名路径（添加 bucket 前缀）
	signPath := ossPath
	if cfg.Bucket != "" {
		// 如果配置了 bucket，将其添加到路径前面
		signPath = "/" + cfg.Bucket + ossPath
	}

	// 4. 【关键修改】先编码路径（用于 MD5 计算和 URL 构建）
	// 核心原则：计算 MD5 签名时，uri 必须使用 URL 编码后的形式（与 CDN 服务器接收到的路径一致）
	encodedPath := encodePathForCDN(signPath)

	// 5. 构建基础 URL（使用编码后的路径）
	baseURL := cfg.Endpoint + encodedPath

	// 6. 如果启用 CDN 鉴权，添加 sign 参数
	if cfg.CdnAuth.Enable {
		// 【关键】使用编码后的路径计算签名
		logs.Info("[Type A] 原始路径: %s", ossPath)
		logs.Info("[Type A] 签名路径 (添加 bucket): %s", signPath)
		logs.Info("[Type A] 编码路径 (用于 MD5 和 URL): %s", encodedPath)
		logs.Info("[Type A] ⚠️ 关键：MD5 计算必须使用编码后的路径！")

		authKey := GenerateAuthKey(
			encodedPath, // 使用编码后的路径
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
