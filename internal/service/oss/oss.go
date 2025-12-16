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
	logs.Info("========================================")
	logs.Info("===== Type-A 鉴权签名生成（与测试代码一致）=====")
	logs.Info("========================================")

	// ==================== 步骤1：准备参数 ====================
	logs.Info("[步骤1] 准备时间戳参数")
	timestamp := time.Now().Unix()
	timestampStr := strconv.FormatInt(timestamp, 10)
	logs.Info("  当前时间戳 (Unix): %s", timestampStr)
	logs.Info("  当前时间 (格式化): %s", time.Unix(timestamp, 0).Format("2006-01-02 15:04:05"))
	logs.Info("  TTL (过期秒数): %d 秒", ttl)
	logs.Info("  过期时间: %s", time.Unix(timestamp+ttl, 0).Format("2006-01-02 15:04:05"))

	// ==================== 步骤2：生成随机字符串 ====================
	logs.Info("----------------------------------------")
	logs.Info("[步骤2] 生成随机字符串")
	randStr := "0"
	if useRandom {
		if randomLength <= 0 {
			randomLength = 6
		}
		randStr = randoms.RandomAlphaNum(randomLength)
		logs.Info("  随机字符串长度: %d", randomLength)
		logs.Info("  生成的随机字符串: %s", randStr)
	} else {
		logs.Info("  随机字符串: %s (未启用随机)", randStr)
	}

	// ==================== 步骤3：计算 MD5 签名 ====================
	logs.Info("----------------------------------------")
	logs.Info("[步骤3] ⚠️ 【关键】计算 MD5 签名")
	logs.Info("  参数说明:")
	logs.Info("    [1] uri (已编码路径): %s", uri)
	logs.Info("    [2] timestamp: %s", timestampStr)
	logs.Info("    [3] rand: %s", randStr)
	logs.Info("    [4] uid: %s", uid)
	logs.Info("    [5] pkey: %s", privateKey)
	logs.Info("")
	logs.Info("  拼接格式: uri-timestamp-rand-uid-pkey")

	rawSignStr := fmt.Sprintf("%s-%s-%s-%s-%s", uri, timestampStr, randStr, uid, privateKey)
	logs.Info("  原始字符串: %s", rawSignStr)
	logs.Info("")
	logs.Info("  计算公式: md5(%s)", rawSignStr)

	md5Hash := md5.New()
	md5Hash.Write([]byte(rawSignStr))
	md5hash := hex.EncodeToString(md5Hash.Sum(nil))
	logs.Info("  MD5 结果: %s", md5hash)

	// ==================== 步骤4：生成最终签名参数 ====================
	logs.Info("----------------------------------------")
	logs.Info("[步骤4] 生成最终签名参数")
	logs.Info("  签名格式: timestamp-rand-uid-md5hash")
	signParam := fmt.Sprintf("%s-%s-%s-%s", timestampStr, randStr, uid, md5hash)
	logs.Info("  最终签名: %s", signParam)
	logs.Info("")
	logs.Info("  ✓ 完整 sign 参数: sign=%s", signParam)
	logs.Info("========================================")

	return signParam
}

// BuildURL 根据 Emby 路径构建完整的 OSS URL (带 CDN 鉴权)
func BuildURL(embyPath string) (string, error) {
	cfg := config.C.Oss
	if !cfg.Enable {
		return "", fmt.Errorf("OSS 功能未启用")
	}

	logs.Info("========================================")
	logs.Info("===== CDN 鉴权 URL 构建过程（详细日志）=====")
	logs.Info("========================================")

	// 1. 映射 Emby 路径到 OSS 路径
	logs.Info("[步骤1] Emby 路径映射")
	logs.Info("  输入 Emby 路径: %s", embyPath)
	ossPath, err := cfg.MapPath(embyPath)
	if err != nil {
		logs.Error("  ❌ 路径映射失败: %v", err)
		return "", fmt.Errorf("路径映射失败: %v", err)
	}
	logs.Info("  映射后 OSS 路径: %s", ossPath)

	// 2. 确保路径以 / 开头
	if !strings.HasPrefix(ossPath, "/") {
		logs.Info("[步骤2] 添加前导斜杠")
		logs.Info("  修改前: %s", ossPath)
		ossPath = "/" + ossPath
		logs.Info("  修改后: %s", ossPath)
	}

	// 2.5. 清理路径中的双斜杠（修复路径映射导致的问题）
	originalOssPath := ossPath
	for strings.Contains(ossPath, "//") {
		ossPath = strings.ReplaceAll(ossPath, "//", "/")
	}
	if originalOssPath != ossPath {
		logs.Info("[步骤3] 清理双斜杠")
		logs.Info("  修改前: %s", originalOssPath)
		logs.Info("  修改后: %s", ossPath)
	}

	// 3. 构建签名路径（添加 bucket 前缀）
	signPath := ossPath
	if cfg.Bucket != "" {
		logs.Info("[步骤4] 添加 Bucket 前缀")
		logs.Info("  Bucket: %s", cfg.Bucket)
		logs.Info("  修改前: %s", signPath)
		signPath = "/" + cfg.Bucket + ossPath
		logs.Info("  修改后: %s", signPath)
	}

	// 4. 【关键修改】先编码路径（用于 MD5 计算和 URL 构建）
	logs.Info("========================================")
	logs.Info("[步骤5] ⚠️ 【关键】编码路径（与测试代码一致）")
	logs.Info("  编码前原始路径: %s", signPath)
	logs.Info("  编码规则: 空格→%%20, 中文→UTF-8编码, 斜杠保留")
	encodedPath := encodePathForCDN(signPath)
	logs.Info("  编码后路径: %s", encodedPath)
	logs.Info("  ⚠️ 注意: MD5 计算将使用编码后的路径！")
	logs.Info("========================================")

	// 5. 构建基础 URL（使用编码后的路径）
	logs.Info("[步骤6] 构建基础 URL")
	logs.Info("  Endpoint: %s", cfg.Endpoint)
	logs.Info("  编码路径: %s", encodedPath)
	baseURL := cfg.Endpoint + encodedPath
	logs.Info("  基础 URL: %s", baseURL)

	// 6. 如果启用 CDN 鉴权，添加 sign 参数
	if cfg.CdnAuth.Enable {
		logs.Info("========================================")
		logs.Info("[步骤7] 生成 CDN 鉴权签名")
		logs.Info("  PrivateKey: %s", cfg.CdnAuth.PrivateKey)
		logs.Info("  TTL: %d 秒", cfg.CdnAuth.TTL)
		logs.Info("  UID: %s", cfg.CdnAuth.UID)
		logs.Info("  UseRandom: %v", cfg.CdnAuth.UseRandom)
		logs.Info("  RandomLength: %d", cfg.CdnAuth.RandomLength)
		logs.Info("  ⚠️ 签名计算输入: 编码后的路径")
		logs.Info("  签名输入路径: %s", encodedPath)

		authKey := GenerateAuthKey(
			encodedPath, // 使用编码后的路径
			cfg.CdnAuth.PrivateKey,
			cfg.CdnAuth.TTL,
			cfg.CdnAuth.UID,
			cfg.CdnAuth.UseRandom,
			cfg.CdnAuth.RandomLength,
		)

		logs.Info("  返回签名: %s", authKey)
		baseURL += "?sign=" + authKey
		logs.Info("  最终 URL: %s", baseURL)
	}

	logs.Info("========================================")
	logs.Success("✓ OSS URL 生成成功")
	logs.Info("========================================")
	return baseURL, nil
}

// MapPath 是 BuildURL 的辅助方法，仅用于路径映射测试
func MapPath(embyPath string) (string, error) {
	return config.C.Oss.MapPath(embyPath)
}
