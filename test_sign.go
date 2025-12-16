package main

 import (
   "crypto/md5"
   "encoding/hex"
   "fmt"
   "math/rand"
   "net/url"
   "strconv"
   "strings"
   "time"
 )

 // generateRandStr 生成随机字符串
 func generateRandStr(length int) string {
   if length <= 0 {
     length = 18
   }
   const chars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
   sb := strings.Builder{}
   rand.Seed(time.Now().UnixNano())
   for i := 0; i < length; i++ {
     sb.WriteByte(chars[rand.Intn(len(chars))])
   }
   return sb.String()
 }

 // calculateMD5Hash 计算鉴权Hash
 func calculateMD5Hash(uri, timestampStr, randStr, uid, pkey string) string {
   // 关键：uri 必须是 URL 编码后的路径（但保留斜杠）
   rawSignStr := fmt.Sprintf("%s-%s-%s-%s-%s", uri, timestampStr, randStr, uid, pkey)

   md5Hash := md5.New()
   md5Hash.Write([]byte(rawSignStr))
   return hex.EncodeToString(md5Hash.Sum(nil))
 }

 // encodePathForCDN 对路径进行编码，保留斜杠
 func encodePathForCDN(path string) string {
   // 将路径按 / 分割
   segments := strings.Split(path, "/")

   // 对每个片段进行编码
   for i, segment := range segments {
     segments[i] = url.QueryEscape(segment)
   }

   // 重新组合
   encoded := strings.Join(segments, "/")

   // QueryEscape 会将空格编码为 +，需要改为 %20（路径标准）
   encoded = strings.ReplaceAll(encoded, "+", "%20")

   return encoded
 }

 func main() {
   // ==================== 配置项 ====================
   pkey := "S8fo7IIsSSTTRX3fIE"
   domain := "qiufeng.huaijiufu.com"

   // 【测试用例】：直接使用带空格和中文的路径
   inputPath := "/meiti/图 片.jpg"
   // inputPath := "/meiti/测试图片.jpg"  // 中文测试

   authExpireSeconds := 600 // 定义了过期时间
   randLength := 18

   // ==================== 步骤1：编码路径 ====================
   // 关键修改：先编码路径，再用于 MD5 计算
   encodedURI := encodePathForCDN(inputPath)

   // ==================== 步骤2：准备参数 ====================
   timestamp := time.Now().Unix()
   timestampStr := strconv.FormatInt(timestamp, 10)
   randStr := generateRandStr(randLength)
   uid := "0"

   // ==================== 步骤3：计算签名 (使用编码后的路径) ====================
   md5hash := calculateMD5Hash(encodedURI, timestampStr, randStr, uid, pkey)

   // ==================== 步骤4：生成URL ====================
   signParam := fmt.Sprintf("%s-%s-%s-%s", timestampStr, randStr, uid, md5hash)

   // 构造最终 URL
   fullURL := fmt.Sprintf("http://%s%s?sign=%s", domain, encodedURI, signParam)

   // ==================== 验证输出 ====================
   fmt.Println("===== 腾讯云鉴权调试 (Go版 - 修复版) =====")
   fmt.Println("1. 原始路径 (用户输入):")
   fmt.Printf("   [%s]\n", inputPath)

   fmt.Println("\n2. 编码路径 (MD5计算用):")
   fmt.Printf("   [%s]\n", encodedURI)
   fmt.Printf("   ⚠️ 关键：MD5 计算必须使用编码后的路径！\n")

   fmt.Printf("\n3. 鉴权参数:\n")
   fmt.Printf("   过期时间设置: %d 秒\n", authExpireSeconds)
   fmt.Printf("   时间戳: %s\n", timestampStr)
   fmt.Printf("   随机字符串: %s\n", randStr)
   fmt.Printf("   用户ID: %s\n", uid)
   fmt.Printf("   密钥: %s\n", pkey)
   fmt.Printf("   当前时间: %s\n", time.Unix(timestamp, 0).Format(time.RFC3339))

   fmt.Println("\n4. 签名计算公式:")
   fmt.Printf("   md5(%s-%s-%s-%s-%s)\n", encodedURI, timestampStr, randStr, uid, pkey)
   fmt.Printf("   = %s\n", md5hash)

   fmt.Println("\n5. 最终结果 (复制到浏览器测试):")
   fmt.Println(fullURL)

   fmt.Println("\n===== 说明 =====")
   fmt.Println("✓ 空格会被编码为 %20")
   fmt.Println("✓ 中文会被编码为 UTF-8 百分号编码")
   fmt.Println("✓ MD5 计算时使用的是编码后的路径")
   fmt.Println("✓ 这样才能与 CDN 服务器的验证逻辑一致")
 }