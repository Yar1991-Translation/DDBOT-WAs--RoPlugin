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

// API 端点常量
var (
	// 使用代理的 API 端点
	usersAPI    = config.Proxy + "/users/v1"
	presenceAPI = config.Proxy + "/presence/v1"
	gamesAPI    = config.Proxy + "/games/v1"
	apisAPI     = config.Proxy + "/apis/v1"
)

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

// HACK: model.go 