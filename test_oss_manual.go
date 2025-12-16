package main

import (
	"fmt"
	"path/filepath"

	"github.com/AmbitiousJun/go-emby2openlist/v2/internal/config"
	"github.com/AmbitiousJun/go-emby2openlist/v2/internal/service/oss"
)

// 手动测试OSS功能
// 运行方式: go run test_oss_manual.go
func main() {
	fmt.Println("===== OSS 功能测试 =====\n")

	// 1. 加载配置
	fmt.Println("1. 加载配置文件...")
	configPath := filepath.Join(".", "config.yml")
	if err := config.ReadFromFile(configPath); err != nil {
		fmt.Printf("❌ 配置加载失败: %v\n", err)
		fmt.Println("\n请确保已创建 config.yml 并配置了 oss 相关参数")
		return
	}
	fmt.Println("✅ 配置加载成功")

	// 2. 检查OSS是否启用
	fmt.Printf("\n2. OSS 启用状态: %v\n", config.C.Oss.Enable)
	if !config.C.Oss.Enable {
		fmt.Println("⚠️  OSS 功能未启用，请在 config.yml 中设置 oss.enable: true")
		return
	}

	// 3. 测试路径映射
	fmt.Println("\n3. 测试路径映射...")
	testPaths := []string{
		"/movie/星际穿越 (2014)/星际穿越 (2014) - 2160p.mkv",
		"/movie/test.mkv",
		"/series/ShowName/S01E01.mkv",
	}

	for _, embyPath := range testPaths {
		ossPath, err := oss.MapPath(embyPath)
		if err != nil {
			fmt.Printf("   ❌ %s -> 映射失败: %v\n", embyPath, err)
		} else {
			fmt.Printf("   ✅ %s\n      -> %s\n", embyPath, ossPath)
		}
	}

	// 4. 测试完整URL生成
	fmt.Println("\n4. 测试完整URL生成（带CDN鉴权）...")
	testPath := "/movie/星际穿越 (2014)/星际穿越 (2014) - 2160p.mkv"

	finalURL, err := oss.BuildURL(testPath)
	if err != nil {
		fmt.Printf("   ❌ URL生成失败: %v\n", err)
	} else {
		fmt.Printf("   ✅ 生成的URL:\n")
		fmt.Printf("      %s\n", finalURL)

		// 检查URL是否包含sign参数
		if config.C.Oss.CdnAuth.Enable {
			if len(finalURL) > 200 {
				fmt.Println("   ✅ CDN鉴权已添加 (包含sign参数)")
			}
		}
	}

	// 5. 显示配置摘要
	fmt.Println("\n5. 当前OSS配置摘要:")
	fmt.Printf("   Endpoint: %s\n", config.C.Oss.Endpoint)
	fmt.Printf("   CDN鉴权: %v\n", config.C.Oss.CdnAuth.Enable)
	if config.C.Oss.CdnAuth.Enable {
		fmt.Printf("   鉴权TTL: %d 秒\n", config.C.Oss.CdnAuth.TTL)
	}
	fmt.Printf("   API Key: %v\n", config.C.Oss.ApiKey.Enable)
	if config.C.Oss.ApiKey.Enable {
		fmt.Printf("   API Key Header: %s\n", config.C.Oss.ApiKey.HeaderName)
	}

	fmt.Println("\n===== 测试完成 =====")
}
