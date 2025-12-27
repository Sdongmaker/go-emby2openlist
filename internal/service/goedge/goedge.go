package goedge

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/AmbitiousJun/go-emby2openlist/v2/internal/config"
	"github.com/AmbitiousJun/go-emby2openlist/v2/internal/util/logs"
	"github.com/AmbitiousJun/go-emby2openlist/v2/internal/util/randoms"
)

// GenerateAuthSign 生成 GoEdge CDN 鉴权的 sign 参数
// 签名格式: sign = {timestamp}-{rand}-{md5hash}
// MD5 计算: md5hash = md5("{path}@{timestamp}@{rand}@{privateKey}")
// 注意:
//   1. path 是原始路径，不需要 URL 编码
//   2. 连接符固定使用 "@"
//   3. MD5 结果必须是小写
func GenerateAuthSign(path string, privateKey string, ttl int64, useRandom bool, randomLength int) string {
	// 获取当前时间戳（UTC）
	timestamp := time.Now().Unix()
	timestampStr := strconv.FormatInt(timestamp, 10)

	// 生成随机字符串
	randStr := "0"
	if useRandom {
		if randomLength <= 0 {
			randomLength = 16
		}
		randStr = randoms.RandomAlphaNum(randomLength)
	}

	// 计算 MD5 签名（固定使用 @ 作为连接符）
	rawSignStr := fmt.Sprintf("%s@%s@%s@%s", path, timestampStr, randStr, privateKey)
	md5Hash := md5.New()
	md5Hash.Write([]byte(rawSignStr))
	// GoEdge 要求 MD5 结果必须是小写
	md5hash := strings.ToLower(hex.EncodeToString(md5Hash.Sum(nil)))

	// 生成最终签名
	signParam := fmt.Sprintf("%s-%s-%s", timestampStr, randStr, md5hash)

	// 输出关键信息日志
	logs.Info("[GoEdge Auth] sign=%s, ts=%s, md5=%s", signParam, timestampStr, md5hash[:8]+"...")

	return signParam
}

// BuildURL 根据 Emby 路径构建完整的 GoEdge URL (带鉴权)
func BuildURL(embyPath string) (string, error) {
	cfg := config.C.GoEdge
	if !cfg.Enable {
		return "", fmt.Errorf("GoEdge 功能未启用")
	}

	// 1. 映射 Emby 路径到 GoEdge 路径
	goedgePath, err := cfg.MapPath(embyPath)
	if err != nil {
		return "", fmt.Errorf("路径映射失败: %v", err)
	}

	// 2. 确保路径以 / 开头
	if !strings.HasPrefix(goedgePath, "/") {
		goedgePath = "/" + goedgePath
	}

	// 3. 清理路径中的双斜杠
	for strings.Contains(goedgePath, "//") {
		goedgePath = strings.ReplaceAll(goedgePath, "//", "/")
	}

	logs.Info("[GoEdge] Path: %s", goedgePath)

	// 4. 构建基础 URL（GoEdge 不需要对路径进行 URL 编码）
	baseURL := cfg.Endpoint + goedgePath

	// 5. 如果启用鉴权，添加 sign 参数
	if cfg.Auth.Enable {
		authSign := GenerateAuthSign(
			goedgePath, // 使用原始路径，不编码
			cfg.Auth.PrivateKey,
			cfg.Auth.TTL,
			cfg.Auth.UseRandom,
			cfg.Auth.RandomLength,
		)
		baseURL += "?sign=" + authSign
	}

	logs.Success("[GoEdge] URL: %s", baseURL)
	return baseURL, nil
}

// MapPath 是 BuildURL 的辅助方法，仅用于路径映射测试
func MapPath(embyPath string) (string, error) {
	return config.C.GoEdge.MapPath(embyPath)
}
