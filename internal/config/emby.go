package config

import (
	"errors"
	"fmt"
	"strings"

	"github.com/AmbitiousJun/go-emby2openlist/v2/internal/util/logs"
	"github.com/AmbitiousJun/go-emby2openlist/v2/internal/util/maps"
	"github.com/AmbitiousJun/go-emby2openlist/v2/internal/util/randoms"
	"github.com/AmbitiousJun/go-emby2openlist/v2/internal/util/strs"
)

// PeStrategy 代理异常策略类型
type PeStrategy string

const (
	PeStrategyOrigin PeStrategy = "origin" // 回源
	PeStrategyReject PeStrategy = "reject" // 拒绝请求
)

// DlStrategy 下载策略类型
type DlStrategy string

const (
	DlStrategyOrigin DlStrategy = "origin" // 代理到源服务器
	DlStrategyDirect DlStrategy = "direct" // 获取并重定向到直链
	DlStrategy403    DlStrategy = "403"    // 拒绝响应
)

// validPeStrategy 用于校验用户配置的策略是否合法
var validPeStrategy = map[PeStrategy]struct{}{
	PeStrategyOrigin: {}, PeStrategyReject: {},
}

// validDlStrategy 用于校验用户配置的下载策略是否合法
var validDlStrategy = map[DlStrategy]struct{}{
	DlStrategyOrigin: {}, DlStrategyDirect: {}, DlStrategy403: {},
}

// Emby 相关配置
type Emby struct {
	// Emby 源服务器地址
	Host string `yaml:"host"`
	// rclone 或者 cd 的挂载目录
	MountPath string `yaml:"mount-path"`
	// EpisodesUnplayPrior 在获取剧集列表时是否将未播资源优先展示
	EpisodesUnplayPrior bool `yaml:"episodes-unplay-prior"`
	// ResortRandomItems 是否对随机的 items 进行重排序
	ResortRandomItems bool `yaml:"resort-random-items"`
	// ProxyErrorStrategy 代理错误时的处理策略
	ProxyErrorStrategy PeStrategy `yaml:"proxy-error-strategy"`
	// ImagesQuality 图片质量
	ImagesQuality int `yaml:"images-quality"`
	// Strm strm 配置
	Strm *Strm `yaml:"strm"`
	// DownloadStrategy 下载接口响应策略
	DownloadStrategy DlStrategy `yaml:"download-strategy"`
	// LocalMediaRoot 本地媒体根路径
	LocalMediaRoot string `yaml:"local-media-root"`
	// ItemsCounts 媒体库统计自定义配置
	ItemsCounts *ItemsCountsConfig `yaml:"items-counts"`
}

func (e *Emby) Init() error {
	if strs.AnyEmpty(e.Host) {
		return errors.New("emby.host 配置不能为空")
	}
	if strs.AnyEmpty(string(e.ProxyErrorStrategy)) {
		// 失败默认回源
		e.ProxyErrorStrategy = PeStrategyOrigin
	}
	if strs.AnyEmpty(string(e.DownloadStrategy)) {
		// 默认响应直链
		e.DownloadStrategy = DlStrategyDirect
	}

	e.ProxyErrorStrategy = PeStrategy(strings.TrimSpace(string(e.ProxyErrorStrategy)))
	if _, ok := validPeStrategy[e.ProxyErrorStrategy]; !ok {
		return fmt.Errorf("emby.proxy-error-strategy 配置错误, 有效值: %v", maps.Keys(validPeStrategy))
	}

	if e.ImagesQuality == 0 {
		// 不允许配置零值
		e.ImagesQuality = 70
	}
	if e.ImagesQuality < 0 || e.ImagesQuality > 100 {
		return fmt.Errorf("emby.images-quality 配置错误: %d, 允许配置范围: [1, 100]", e.ImagesQuality)
	}

	if e.Strm == nil {
		e.Strm = new(Strm)
	}
	if err := e.Strm.Init(); err != nil {
		return fmt.Errorf("emby.strm 配置错误: %v", err)
	}

	e.DownloadStrategy = DlStrategy(strings.TrimSpace(string(e.DownloadStrategy)))
	if _, ok := validDlStrategy[e.DownloadStrategy]; !ok {
		return fmt.Errorf("emby.download-strategy 配置错误, 有效值: %v", maps.Keys(validDlStrategy))
	}

	// 如果没有配置, 生成一个随机前缀, 避免将网盘资源误识别为本地
	if e.LocalMediaRoot = strings.TrimSpace(e.LocalMediaRoot); e.LocalMediaRoot == "" {
		e.LocalMediaRoot = "/" + randoms.RandomHex(32)
	}

	// 初始化 ItemsCounts 配置
	if e.ItemsCounts == nil {
		e.ItemsCounts = &ItemsCountsConfig{}
	}
	if err := e.ItemsCounts.Init(); err != nil {
		return fmt.Errorf("emby.items-counts 配置错误: %v", err)
	}

	return nil
}

// Strm strm 配置
type Strm struct {
	// PathMap 远程路径映射
	PathMap []string `yaml:"path-map"`
	// InternalRedirectEnable 是否启用 strm 内部重定向
	InternalRedirectEnable bool `yaml:"internal-redirect-enable"`

	// pathMap 配置初始化后转换为二维数组切片结构
	pathMap [][2]string
}

// Init 配置初始化
func (s *Strm) Init() error {
	s.pathMap = make([][2]string, 0, len(s.PathMap))
	for _, path := range s.PathMap {
		splits := strings.Split(path, "=>")
		if len(splits) != 2 {
			return fmt.Errorf("映射配置不规范: %s, 请使用 => 进行分割", path)
		}
		from, to := strings.TrimSpace(splits[0]), strings.TrimSpace(splits[1])
		s.pathMap = append(s.pathMap, [2]string{from, to})
	}
	return nil
}

// MapPath 将传入路径按照预配置的映射关系从上到下按顺序进行映射,
// 至多成功映射一次
func (s *Strm) MapPath(path string) string {
	for _, m := range s.pathMap {
		from, to := m[0], m[1]
		if strings.Contains(path, from) {
			logs.Tip("映射路径: [%s] => [%s]", from, to)
			return strings.Replace(path, from, to, 1)
		}
	}
	return path
}

// ItemsCountsConfig 媒体库统计自定义配置
type ItemsCountsConfig struct {
	// Enable 是否启用自定义媒体库统计
	Enable bool `yaml:"enable"`
	// Mode 工作模式
	// - origin: 代理回源，不做修改（默认）
	// - custom: 完全自定义每个字段的值
	// - modify: 基于真实数据乘以系数进行修改
	Mode string `yaml:"mode"`
	// Multiplier 修改模式下的乘数系数（仅在 mode=modify 时生效）
	Multiplier float64 `yaml:"multiplier"`

	// 以下字段仅在 mode=custom 时使用
	MovieCount      int `yaml:"movie-count"`       // 电影数量
	SeriesCount     int `yaml:"series-count"`      // 剧集数量
	EpisodeCount    int `yaml:"episode-count"`     // 剧集集数
	GameCount       int `yaml:"game-count"`        // 游戏数量
	ArtistCount     int `yaml:"artist-count"`      // 艺术家数量
	ProgramCount    int `yaml:"program-count"`     // 节目数量
	GameSystemCount int `yaml:"game-system-count"` // 游戏系统数量
	TrailerCount    int `yaml:"trailer-count"`     // 预告片数量
	SongCount       int `yaml:"song-count"`        // 歌曲数量
	AlbumCount      int `yaml:"album-count"`       // 专辑数量
	MusicVideoCount int `yaml:"music-video-count"` // 音乐视频数量
	BoxSetCount     int `yaml:"box-set-count"`     // 合集数量
	BookCount       int `yaml:"book-count"`        // 书籍数量
	ItemCount       int `yaml:"item-count"`        // 总项目数量
}

// Init 配置初始化
func (ic *ItemsCountsConfig) Init() error {
	if !ic.Enable {
		return nil
	}

	// 默认模式为 origin
	if ic.Mode == "" {
		ic.Mode = "origin"
	}

	// 验证模式
	validModes := map[string]bool{
		"origin": true,
		"custom": true,
		"modify": true,
	}
	if !validModes[ic.Mode] {
		return fmt.Errorf("items-counts.mode 配置错误: %s, 有效值: origin, custom, modify", ic.Mode)
	}

	// 修改模式下验证系数
	if ic.Mode == "modify" {
		if ic.Multiplier <= 0 {
			ic.Multiplier = 1.0 // 默认值为 1.0（不修改）
		}
		logs.Success("ItemsCounts 配置: 修改模式，系数 = %.2f", ic.Multiplier)
	}

	// 自定义模式下记录日志
	if ic.Mode == "custom" {
		logs.Success("ItemsCounts 配置: 自定义模式，电影=%d, 剧集=%d, 总计=%d",
			ic.MovieCount, ic.SeriesCount, ic.ItemCount)
	}

	return nil
}
