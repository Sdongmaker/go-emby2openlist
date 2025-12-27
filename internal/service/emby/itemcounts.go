package emby

import (
	"encoding/json"
	"net/http"

	"github.com/AmbitiousJun/go-emby2openlist/v2/internal/config"
	"github.com/AmbitiousJun/go-emby2openlist/v2/internal/util/https"
	"github.com/AmbitiousJun/go-emby2openlist/v2/internal/util/logs"
	"github.com/gin-gonic/gin"
)

// ItemCounts 媒体库统计数据结构
type ItemCounts struct {
	MovieCount       int `json:"MovieCount"`       // 电影数量
	SeriesCount      int `json:"SeriesCount"`      // 剧集数量
	EpisodeCount     int `json:"EpisodeCount"`     // 剧集集数
	GameCount        int `json:"GameCount"`        // 游戏数量
	ArtistCount      int `json:"ArtistCount"`      // 艺术家数量
	ProgramCount     int `json:"ProgramCount"`     // 节目数量
	GameSystemCount  int `json:"GameSystemCount"`  // 游戏系统数量
	TrailerCount     int `json:"TrailerCount"`     // 预告片数量
	SongCount        int `json:"SongCount"`        // 歌曲数量
	AlbumCount       int `json:"AlbumCount"`       // 专辑数量
	MusicVideoCount  int `json:"MusicVideoCount"`  // 音乐视频数量
	BoxSetCount      int `json:"BoxSetCount"`      // 合集数量
	BookCount        int `json:"BookCount"`        // 书籍数量
	ItemCount        int `json:"ItemCount"`        // 总项目数量
}

// HandleItemsCounts 处理 /Items/Counts 接口
func HandleItemsCounts(c *gin.Context) {
	cfg := config.C.Emby.ItemsCounts

	// 如果未启用自定义或模式为 origin，代理回源
	if !cfg.Enable || cfg.Mode == "origin" {
		logs.Info("[ItemsCounts] 代理回源获取真实数据")
		ProxyOrigin(c)
		return
	}

	// 自定义模式：返回配置的自定义值
	if cfg.Mode == "custom" {
		// 检查是否请求特定媒体库的统计（通过 ParentId 参数）
		parentID := c.Query("ParentId")
		if parentID != "" {
			// 尝试获取该媒体库的自定义配置
			libCounts := cfg.GetLibraryCounts(parentID)
			if libCounts != nil {
				// 返回该媒体库的自定义统计
				libName := libCounts.Name
				if libName == "" {
					libName = parentID
				}
				logs.Info("[ItemsCounts] 返回媒体库 [%s] 的自定义统计", libName)
				counts := ItemCounts{
					MovieCount:      libCounts.MovieCount,
					SeriesCount:     libCounts.SeriesCount,
					EpisodeCount:    libCounts.EpisodeCount,
					GameCount:       libCounts.GameCount,
					ArtistCount:     libCounts.ArtistCount,
					ProgramCount:    libCounts.ProgramCount,
					GameSystemCount: libCounts.GameSystemCount,
					TrailerCount:    libCounts.TrailerCount,
					SongCount:       libCounts.SongCount,
					AlbumCount:      libCounts.AlbumCount,
					MusicVideoCount: libCounts.MusicVideoCount,
					BoxSetCount:     libCounts.BoxSetCount,
					BookCount:       libCounts.BookCount,
					ItemCount:       libCounts.ItemCount,
				}
				c.JSON(http.StatusOK, counts)
				return
			}
			// 没有配置该媒体库，记录警告并代理回源
			logs.Warn("[ItemsCounts] 媒体库 ID [%s] 未在配置中，代理回源", parentID)
			ProxyOrigin(c)
			return
		}

		// 没有 ParentId 参数，返回全局默认值
		logs.Info("[ItemsCounts] 返回全局默认自定义统计数据")
		counts := ItemCounts{
			MovieCount:      cfg.MovieCount,
			SeriesCount:     cfg.SeriesCount,
			EpisodeCount:    cfg.EpisodeCount,
			GameCount:       cfg.GameCount,
			ArtistCount:     cfg.ArtistCount,
			ProgramCount:    cfg.ProgramCount,
			GameSystemCount: cfg.GameSystemCount,
			TrailerCount:    cfg.TrailerCount,
			SongCount:       cfg.SongCount,
			AlbumCount:      cfg.AlbumCount,
			MusicVideoCount: cfg.MusicVideoCount,
			BoxSetCount:     cfg.BoxSetCount,
			BookCount:       cfg.BookCount,
			ItemCount:       cfg.ItemCount,
		}
		c.JSON(http.StatusOK, counts)
		return
	}

	// 修改模式：基于真实数据进行修改
	if cfg.Mode == "modify" {
		logs.Info("[ItemsCounts] 获取真实数据并修改")

		// 获取真实数据
		realCounts, err := fetchRealItemsCounts(c)
		if err != nil {
			logs.Error("[ItemsCounts] 获取真实数据失败: %v, 回源处理", err)
			ProxyOrigin(c)
			return
		}

		// 应用修改（乘以系数）
		realCounts.MovieCount = int(float64(realCounts.MovieCount) * cfg.Multiplier)
		realCounts.SeriesCount = int(float64(realCounts.SeriesCount) * cfg.Multiplier)
		realCounts.EpisodeCount = int(float64(realCounts.EpisodeCount) * cfg.Multiplier)
		realCounts.GameCount = int(float64(realCounts.GameCount) * cfg.Multiplier)
		realCounts.ArtistCount = int(float64(realCounts.ArtistCount) * cfg.Multiplier)
		realCounts.ProgramCount = int(float64(realCounts.ProgramCount) * cfg.Multiplier)
		realCounts.GameSystemCount = int(float64(realCounts.GameSystemCount) * cfg.Multiplier)
		realCounts.TrailerCount = int(float64(realCounts.TrailerCount) * cfg.Multiplier)
		realCounts.SongCount = int(float64(realCounts.SongCount) * cfg.Multiplier)
		realCounts.AlbumCount = int(float64(realCounts.AlbumCount) * cfg.Multiplier)
		realCounts.MusicVideoCount = int(float64(realCounts.MusicVideoCount) * cfg.Multiplier)
		realCounts.BoxSetCount = int(float64(realCounts.BoxSetCount) * cfg.Multiplier)
		realCounts.BookCount = int(float64(realCounts.BookCount) * cfg.Multiplier)
		realCounts.ItemCount = int(float64(realCounts.ItemCount) * cfg.Multiplier)

		logs.Success("[ItemsCounts] 返回修改后的统计数据 (系数: %.2f)", cfg.Multiplier)
		c.JSON(http.StatusOK, realCounts)
		return
	}

	// 未知模式，回源处理
	logs.Warn("[ItemsCounts] 未知模式: %s, 回源处理", cfg.Mode)
	ProxyOrigin(c)
}

// fetchRealItemsCounts 从 Emby 源服务器获取真实的统计数据
func fetchRealItemsCounts(c *gin.Context) (*ItemCounts, error) {
	// 构建源服务器 URL
	originURL := config.C.Emby.Host + c.Request.RequestURI

	// 发起请求
	resp, err := https.Get(originURL).
		Header(c.Request.Header.Clone()).
		Do()
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// 解析响应
	var counts ItemCounts
	if err := json.NewDecoder(resp.Body).Decode(&counts); err != nil {
		return nil, err
	}

	return &counts, nil
}
