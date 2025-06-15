package roblox

import (
	"github.com/cnxysoft/DDBOT-WSa/lsp/concern_type"
	"github.com/sirupsen/logrus"
	"time"
)

// UserInfo 表示 Roblox 用户的基本信息
type UserInfo struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
}

// UserPresence 表示 Roblox 用户的在线状态信息
type UserPresence struct {
	UserID           int64  `json:"userId"`
	UserPresenceType int    `json:"userPresenceType"`
	LastLocation     string `json:"lastLocation"`
	GameID           int64  `json:"rootPlaceId"`
	LastOnline       string `json:"lastOnline"`
}

// GameInfo 表示 Roblox 游戏的基本信息
type GameInfo struct {
	ID         int64  `json:"rootPlaceId"`
	Name       string `json:"name"`
	Playing    int64  `json:"playing"`
	UniverseId int64  `json:"universeId"`
}

// 用户在线状态类型常量
const (
	UserStatusOffline   = 0 // 离线
	UserStatusOnline    = 1 // 在线
	UserStatusInGame    = 2 // 在游戏中
	UserStatusInStudio  = 3 // 在 Studio 中
)

// API 端点，在 init 中根据配置初始化。
var (
	usersAPI    string // e.g. https://users.roblox.com
	presenceAPI string // e.g. https://presence.roblox.com
	gamesAPI    string // e.g. https://games.roblox.com
	apisAPI     string // e.g. https://apis.roblox.com
)

// init 在包加载时执行，负责加载配置并初始化 API 端点。
func init() {
	// 加载 roblox.yaml 配置，获取 Proxy 等参数
	loadConfig()

	// 如果配置文件指定了代理，则使用代理；否则使用 Roblox 官方 API
	if config.Proxy != "" {
		// 移除可能存在的尾部斜杠
		if config.Proxy[len(config.Proxy)-1] == '/' {
			config.Proxy = config.Proxy[:len(config.Proxy)-1]
		}

		usersAPI = config.Proxy
		presenceAPI = config.Proxy
		gamesAPI = config.Proxy
		apisAPI = config.Proxy
	} else {
		usersAPI = "https://users.roblox.com"
		presenceAPI = "https://presence.roblox.com"
		gamesAPI = "https://games.roblox.com"
		apisAPI = "https://apis.roblox.com"
	}
}

// UserStatusEvent 用户状态事件
type UserStatusEvent struct {
	UserID      int64
	UserName    string
	Status      string
	LastStatus  string
	Time        time.Time
	ProfileLink string
}

// Site 返回服务的唯一标识符
func (e *UserStatusEvent) Site() string {
	return ServiceName
}

// Type 返回事件类型
func (e *UserStatusEvent) Type() concern_type.Type {
	return UserType
}

// GetUid 返回用户ID
func (e *UserStatusEvent) GetUid() interface{} {
	return e.UserID
}

// Logger 返回日志记录器
func (e *UserStatusEvent) Logger() *logrus.Entry {
	return log.WithFields(logrus.Fields{
		"Site":     e.Site(),
		"Type":     e.Type(),
		"Uid":      e.GetUid(),
		"UserName": e.UserName,
	})
}

// GamePlayingEvent 游戏在线人数事件
type GamePlayingEvent struct {
	GameID      int64
	GameName    string
	Playing     int64
	LastPlaying int64
	Time        time.Time
	GameLink    string
}

// Site 返回服务的唯一标识符
func (e *GamePlayingEvent) Site() string {
	return ServiceName
}

// Type 返回事件类型
func (e *GamePlayingEvent) Type() concern_type.Type {
	return GameType
}

// GetUid 返回游戏ID
func (e *GamePlayingEvent) GetUid() interface{} {
	return e.GameID
}

// Logger 返回日志记录器
func (e *GamePlayingEvent) Logger() *logrus.Entry {
	return log.WithFields(logrus.Fields{
		"Site":     e.Site(),
		"Type":     e.Type(),
		"Uid":      e.GetUid(),
		"GameName": e.GameName,
	})
}

// FriendOnlineEvent 好友上线事件
type FriendOnlineEvent struct {
	FriendID    int64
	FriendName  string
	Status      string
	Time        time.Time
	ProfileLink string
}

func (e *FriendOnlineEvent) Site() string {
	return ServiceName
}

func (e *FriendOnlineEvent) Type() concern_type.Type {
	return FriendType
}

func (e *FriendOnlineEvent) GetUid() interface{} {
	return e.FriendID
}

func (e *FriendOnlineEvent) Logger() *logrus.Entry {
	return log.WithFields(logrus.Fields{
		"Site":  e.Site(),
		"Type":  e.Type(),
		"Uid":   e.GetUid(),
		"FName": e.FriendName,
	})
}

// UserSearchResponse 用于解析用户名搜索API的响应
type UserSearchResponse struct {
	Data []struct {
		ID          int64  `json:"id"`
		Name        string `json:"name"`
		DisplayName string `json:"displayName"`
	} `json:"data"`
}

// HACK: model.go 