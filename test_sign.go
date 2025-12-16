package main

import (
	"crypto/md5"
	"fmt"
	"time"
)

// 独立的 Type-A 签名生成器，用于验证算法正确性
// 完全按照 Python Reqable 脚本的实现逻辑
func main() {
	fmt.Println("===== Type-A 签名算法测试工具 =====\n")

	// ================= 配置区域 =================
	// 这些值应该与你的配置文件和 Python 脚本保持一致
	secretKey := "S8fo7IIsSSTTRX3fIE"
	ttl := int64(600)
	uid := "0"
	randStr := "0"

	// 测试路径 - 修改为你实际的文件路径
	path := "/bucket/path/to/video.mp4"
	// ==========================================

	fmt.Println("配置参数:")
	fmt.Printf("  密钥: %s\n", secretKey)
	fmt.Printf("  TTL: %d 秒\n", ttl)
	fmt.Printf("  UID: %s\n", uid)
	fmt.Printf("  随机数: %s\n", randStr)
	fmt.Printf("  路径: %s\n\n", path)

	// 1. 计算时间戳
	nowTs := time.Now().Unix()
	expireTime := nowTs + ttl

	fmt.Println("时间戳计算:")
	fmt.Printf("  当前时间: %d\n", nowTs)
	fmt.Printf("  过期时间: %d (当前时间 + %d秒)\n\n", expireTime, ttl)

	// 2. 构造签名字符串
	// Python: raw_str = f"{path}-{expire_time}-{rand_str}-{uid}-{secret_key}"
	rawStr := fmt.Sprintf("%s-%d-%s-%s-%s", path, expireTime, randStr, uid, secretKey)

	fmt.Println("签名字符串 (raw_str):")
	fmt.Printf("  格式: path-time-rand-uid-key\n")
	fmt.Printf("  内容: %s\n\n", rawStr)

	// 3. 计算 MD5
	// Python: md5_signature = hashlib.md5(raw_str.encode('utf-8')).hexdigest()
	hash := md5.Sum([]byte(rawStr))
	md5Signature := fmt.Sprintf("%x", hash)

	fmt.Println("MD5 计算:")
	fmt.Printf("  MD5: %s\n\n", md5Signature)

	// 4. 构造最终签名
	// Python: auth_value = f"{expire_time}-{rand_str}-{uid}-{md5_signature}"
	authValue := fmt.Sprintf("%d-%s-%s-%s", expireTime, randStr, uid, md5Signature)

	fmt.Println("最终签名 (sign 参数值):")
	fmt.Printf("  格式: time-rand-uid-md5hash\n")
	fmt.Printf("  sign=%s\n\n", authValue)

	// 5. 生成完整 URL
	fullURL := fmt.Sprintf("https://your-cdn.com%s?sign=%s", path, authValue)
	fmt.Println("完整 URL:")
	fmt.Printf("  %s\n\n", fullURL)

	fmt.Println("===== 对比检查清单 =====")
	fmt.Println("请将上述输出与 Python Reqable 脚本的日志对比:")
	fmt.Println("  1. [Type A] path 是否一致?")
	fmt.Println("  2. 时间戳是否在合理范围内? (允许几秒钟误差)")
	fmt.Println("  3. raw_str 是否完全一致?")
	fmt.Println("  4. MD5 是否一致?")
	fmt.Println("  5. 最终 sign 参数是否一致?")
	fmt.Println("\n如果以上都一致但仍然失败，请检查:")
	fmt.Println("  - 服务器时间是否同步 (时差超过 TTL 会导致验证失败)")
	fmt.Println("  - CDN 配置中的密钥是否正确")
	fmt.Println("  - 请求路径是否与签名路径完全匹配 (包括 bucket)")
}
