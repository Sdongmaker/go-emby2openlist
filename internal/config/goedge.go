package config

import (
	"fmt"
	"strings"

	"github.com/AmbitiousJun/go-emby2openlist/v2/internal/util/logs"
)

// GoEdge GoEdge CDN 配置
type GoEdge struct {
	// Enable 是否启用 GoEdge CDN 重定向功能
	Enable bool `yaml:"enable"`
	// Endpoint GoEdge CDN 访问域名
	Endpoint string `yaml:"endpoint"`
	// PathMapping Emby路径到GoEdge路径的映射关系
	PathMapping []string `yaml:"path-mapping"`
	// Auth GoEdge 鉴权配置
	Auth *GoEdgeAuth `yaml:"auth"`

	// 内部使用的映射表
	pathMappingTable map[string]string
}

// GoEdgeAuth GoEdge 鉴权配置
type GoEdgeAuth struct {
	// Enable 是否启用 GoEdge 鉴权
	Enable bool `yaml:"enable"`
	// PrivateKey 鉴权密钥
	PrivateKey string `yaml:"private-key"`
	// TTL 鉴权URL有效期 (秒)
	TTL int64 `yaml:"ttl"`
	// UseRandom 是否使用随机字符串增强安全性
	UseRandom bool `yaml:"use-random"`
	// RandomLength 随机字符串长度，默认16位
	RandomLength int `yaml:"random-length"`
}

// Init 配置初始化
func (g *GoEdge) Init() error {
	if !g.Enable {
		return nil
	}

	// 初始化零值
	if g.Auth == nil {
		g.Auth = &GoEdgeAuth{}
	}

	// 验证必填配置
	if g.Endpoint == "" {
		return fmt.Errorf("goedge.endpoint 不能为空")
	}

	// 移除 endpoint 末尾的斜杠
	g.Endpoint = strings.TrimRight(g.Endpoint, "/")

	// 初始化路径映射表
	g.pathMappingTable = make(map[string]string)
	for _, mapping := range g.PathMapping {
		parts := strings.Split(mapping, ":")
		if len(parts) != 2 {
			logs.Warn("无效的路径映射配置: %s, 跳过", mapping)
			continue
		}
		embyPath := strings.TrimSpace(parts[0])
		goedgePath := strings.TrimSpace(parts[1])
		if embyPath == "" || goedgePath == "" {
			logs.Warn("无效的路径映射配置: %s, 跳过", mapping)
			continue
		}
		g.pathMappingTable[embyPath] = goedgePath
		logs.Info("GoEdge 路径映射: %s -> %s", embyPath, goedgePath)
	}

	if len(g.pathMappingTable) == 0 {
		logs.Warn("未配置有效的 GoEdge 路径映射")
	}

	// 验证鉴权配置
	if g.Auth.Enable {
		if g.Auth.PrivateKey == "" {
			return fmt.Errorf("goedge.auth.private-key 不能为空")
		}
		if g.Auth.TTL <= 0 {
			g.Auth.TTL = 3600 // 默认1小时
		}
		if g.Auth.RandomLength <= 0 {
			g.Auth.RandomLength = 16 // 默认16位
		}
		logs.Success("GoEdge 鉴权已启用, TTL: %d 秒, 随机字符串: %v (长度: %d)",
			g.Auth.TTL, g.Auth.UseRandom, g.Auth.RandomLength)
	}

	logs.Success("GoEdge 配置初始化完成: endpoint=%s", g.Endpoint)
	return nil
}

// MapPath 将 Emby 路径映射为 GoEdge 路径
func (g *GoEdge) MapPath(embyPath string) (string, error) {
	if g.pathMappingTable == nil {
		return "", fmt.Errorf("路径映射表未初始化")
	}

	// 遍历映射表,找到匹配的前缀
	for embyPrefix, goedgePrefix := range g.pathMappingTable {
		if strings.HasPrefix(embyPath, embyPrefix) {
			// 替换前缀
			goedgePath := strings.Replace(embyPath, embyPrefix, goedgePrefix, 1)
			return goedgePath, nil
		}
	}

	return "", fmt.Errorf("无法映射 Emby 路径到 GoEdge 路径: %s", embyPath)
}
