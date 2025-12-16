package main

import (
	"crypto/md5"
	"fmt"
	"math/rand"
	"strings"
	"time"
)

// 生成腾讯云风格的随机数（6位字母数字混合）
func randomAlphaNum(n int) string {
	const chars = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	sb := strings.Builder{}
	for i := 0; i < n; i++ {
		sb.WriteByte(chars[rand.Intn(len(chars))])
	}
	return sb.String()
}

// 独立的 Type-A 签名生成器，用于验证算法正确性
// 完全按照 Python Reqable 脚本的实现逻辑
func main() {
	fmt.Println("===== Type-A 签名算法测试工具（腾讯云版本）=====\n")

	// ================= 配置区域 =================
	// 这些值应该与你的配置文件和 Python 脚本保持一致
	secretKey := "S8fo7IIsSSTTRX3fIE"
	ttl := int64(600)
	uid := "0"

	// 随机数配置
	useRandom := false      // 是否使用随机数（false = 固定 "0"）
	randomLength := 6       // 随机数长度（腾讯云建议 6 位）

	// 测试路径 - 修改为你实际的文件路径
	path := "/meiti/外语电影/黑豹2 (2022)/黑豹2 (2022) - 2160p.mkv"
	// ==========================================

	fmt.Println("配置参数:")
	fmt.Printf("  密钥: %s\n", secretKey)
	fmt.Printf("  TTL: %d 秒\n", ttl)
	fmt.Printf("  UID: %s\n", uid)
	fmt.Printf("  使用随机数: %v\n", useRandom)
	fmt.Printf("  随机数长度: %d 位\n", randomLength)
	fmt.Printf("  路径: %s\n\n", path)

	// 1. 计算时间戳
	nowTs := time.Now().Unix()
	expireTime := nowTs + ttl

	fmt.Println("时间戳计算:")
	fmt.Printf("  当前时间: %d\n", nowTs)
	fmt.Printf("  过期时间: %d (当前时间 + %d秒)\n\n", expireTime, ttl)

	// 2. 生成随机数
	var randStr string
	if useRandom {
		randStr = randomAlphaNum(randomLength)
		fmt.Printf("随机数生成: %s (长度: %d)\n\n", randStr, len(randStr))
	} else {
		randStr = "0"
		fmt.Println("随机数: 0 (固定值)\n")
	}

	// 3. 构造签名字符串
	// Python: raw_str = f"{path}-{expire_time}-{rand_str}-{uid}-{secret_key}"
	rawStr := fmt.Sprintf("%s-%d-%s-%s-%s", path, expireTime, randStr, uid, secretKey)

	fmt.Println("签名字符串 (raw_str):")
	fmt.Printf("  格式: path-time-rand-uid-key\n")
	fmt.Printf("  内容: %s\n\n", rawStr)

	// 4. 计算 MD5
	// Python: md5_signature = hashlib.md5(raw_str.encode('utf-8')).hexdigest()
	hash := md5.Sum([]byte(rawStr))
	md5Signature := fmt.Sprintf("%x", hash)

	fmt.Println("MD5 计算:")
	fmt.Printf("  MD5: %s\n\n", md5Signature)

	// 5. 构造最终签名
	// Python: auth_value = f"{expire_time}-{rand_str}-{uid}-{md5_signature}"
	authValue := fmt.Sprintf("%d-%s-%s-%s", expireTime, randStr, uid, md5Signature)

	fmt.Println("最终签名 (sign 参数值):")
	fmt.Printf("  格式: time-rand-uid-md5hash\n")
	fmt.Printf("  sign=%s\n\n", authValue)

	// 6. 生成完整 URL
	fullURL := fmt.Sprintf("https://qiufeng.huaijiufu.com%s?sign=%s", path, authValue)
	fmt.Println("完整 URL:")
	fmt.Printf("  %s\n\n", fullURL)

	fmt.Println("===== 对比检查清单 =====")
	fmt.Println("与腾讯云示例对比:")
	fmt.Println("  1. 随机数格式: 6位字母数字混合 (如 'q87NIR')")
	fmt.Println("  2. sign 格式: time-rand-uid-md5hash")
	fmt.Println("\n如果启用了随机数，每次生成的签名都会不同")
	fmt.Println("建议配置: use-random: false（与 CDN 控制台保持一致）")
}
