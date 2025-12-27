package config

import (
	"fmt"
	"strings"

	"github.com/AmbitiousJun/go-emby2openlist/v2/internal/util/logs"
)

// Oss 对象存储配置
type Oss struct {
	// Enable 是否启用 OSS 重定向功能
	Enable bool `yaml:"enable"`
	// Endpoint 对象存储访问域名
	Endpoint string `yaml:"endpoint"`
	// Bucket 存储桶名称
	Bucket string `yaml:"bucket"`
	// PathMapping Emby路径到OSS路径的映射关系
	PathMapping []string `yaml:"path-mapping"`
	// CdnAuth CDN Type-A 鉴权配置
	CdnAuth *CdnAuth `yaml:"cdn-auth"`
	// ApiKey 源站验证 API Key 配置
	ApiKey *ApiKeyConfig `yaml:"api-key"`

	// 内部使用的映射表
	pathMappingTable map[string]string
}

// CdnAuth CDN Type-A 鉴权配置
type CdnAuth struct {
	// Enable 是否启用 CDN Type-A 鉴权
	Enable bool `yaml:"enable"`
	// PrivateKey CDN 鉴权密钥
	PrivateKey string `yaml:"private-key"`
	// TTL 鉴权URL有效期 (秒)
	TTL int64 `yaml:"ttl"`
	// UID 用户ID
	UID string `yaml:"uid"`
	// UseUID 是否在签名中使用 UID（某些 CDN 配置可能不需要 UID 参与鉴权）
	UseUID bool `yaml:"use-uid"`
	// Separator MD5 计算时的连接符（腾讯云使用 "-"，其他 CDN 可能使用 "@" 等）
	// 默认值："-"（适配腾讯云）
	Separator string `yaml:"separator"`
	// UseRandom 是否使用随机数增强安全性
	UseRandom bool `yaml:"use-random"`
	// RandomLength 随机数长度（腾讯云建议 6 位，阿里云建议 32 位）
	// 默认值：6（适配腾讯云）
	RandomLength int `yaml:"random-length"`
}

// ApiKeyConfig 源站验证 API Key 配置
type ApiKeyConfig struct {
	// Enable 是否启用 API Key 验证
	Enable bool `yaml:"enable"`
	// HeaderName 响应头名称
	HeaderName string `yaml:"header-name"`
	// Key API Key 值
	Key string `yaml:"key"`
}

// Init 配置初始化
func (o *Oss) Init() error {
	if !o.Enable {
		return nil
	}

	// 初始化零值
	if o.CdnAuth == nil {
		o.CdnAuth = &CdnAuth{}
	}
	if o.ApiKey == nil {
		o.ApiKey = &ApiKeyConfig{}
	}

	// 验证必填配置
	if o.Endpoint == "" {
		return fmt.Errorf("oss.endpoint 不能为空")
	}

	// 移除 endpoint 末尾的斜杠
	o.Endpoint = strings.TrimRight(o.Endpoint, "/")

	// 初始化路径映射表
	o.pathMappingTable = make(map[string]string)
	for _, mapping := range o.PathMapping {
		parts := strings.Split(mapping, ":")
		if len(parts) != 2 {
			logs.Warn("无效的路径映射配置: %s, 跳过", mapping)
			continue
		}
		embyPath := strings.TrimSpace(parts[0])
		ossPath := strings.TrimSpace(parts[1])
		if embyPath == "" || ossPath == "" {
			logs.Warn("无效的路径映射配置: %s, 跳过", mapping)
			continue
		}
		o.pathMappingTable[embyPath] = ossPath
		logs.Info("OSS 路径映射: %s -> %s", embyPath, ossPath)
	}

	if len(o.pathMappingTable) == 0 {
		logs.Warn("未配置有效的 OSS 路径映射")
	}

	// 验证 CDN 鉴权配置
	if o.CdnAuth.Enable {
		if o.CdnAuth.PrivateKey == "" {
			return fmt.Errorf("oss.cdn-auth.private-key 不能为空")
		}
		if o.CdnAuth.TTL <= 0 {
			o.CdnAuth.TTL = 3600 // 默认1小时
		}
		if o.CdnAuth.UID == "" {
			o.CdnAuth.UID = "0"
		}
		// UseUID 默认值为 true（向后兼容）
		// 注意：如果 YAML 中明确配置了 false，此处不会覆盖
		if o.CdnAuth.Separator == "" {
			o.CdnAuth.Separator = "-" // 默认使用 "-"（适配腾讯云）
		}
		if o.CdnAuth.RandomLength <= 0 {
			o.CdnAuth.RandomLength = 6 // 默认6位（适配腾讯云）
		}
		logs.Success("CDN Type-A 鉴权已启用, TTL: %d 秒, UseUID: %v, Separator: %s, 随机数: %v (长度: %d)",
			o.CdnAuth.TTL, o.CdnAuth.UseUID, o.CdnAuth.Separator, o.CdnAuth.UseRandom, o.CdnAuth.RandomLength)
	}

	// 验证 API Key 配置
	if o.ApiKey.Enable {
		if o.ApiKey.HeaderName == "" {
			o.ApiKey.HeaderName = "X-Api-Key"
		}
		if o.ApiKey.Key == "" {
			logs.Warn("oss.api-key.key 未配置")
		}
		logs.Success("源站 API Key 验证已启用, Header: %s", o.ApiKey.HeaderName)
	}

	logs.Success("OSS 配置初始化完成: endpoint=%s", o.Endpoint)
	return nil
}

// MapPath 将 Emby 路径映射为 OSS 路径
func (o *Oss) MapPath(embyPath string) (string, error) {
	if o.pathMappingTable == nil {
		return "", fmt.Errorf("路径映射表未初始化")
	}

	// 遍历映射表,找到匹配的前缀
	for embyPrefix, ossPrefix := range o.pathMappingTable {
		if strings.HasPrefix(embyPath, embyPrefix) {
			// 替换前缀
			ossPath := strings.Replace(embyPath, embyPrefix, ossPrefix, 1)
			return ossPath, nil
		}
	}

	return "", fmt.Errorf("无法映射 Emby 路径到 OSS 路径: %s", embyPath)
}
