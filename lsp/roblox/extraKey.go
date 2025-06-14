package roblox

import (
	"fmt"
	"github.com/cnxysoft/DDBOT-WSa/lsp/concern_type"
)

// 用户类型和游戏类型
const (
	UserType concern_type.Type = "user" // 用户订阅类型
	GameType concern_type.Type = "game" // 游戏订阅类型
	FriendType concern_type.Type = "friend" // 好友上线订阅类型
)

// extraKey 定义数据库键
type extraKey struct{}

// UserStatusKey 返回用户状态的键
func (e *extraKey) UserStatusKey(uid int64) string {
	return fmt.Sprintf("roblox:user:%d:status", uid)
}

// GamePlayingKey 返回游戏在线人数的键
func (e *extraKey) GamePlayingKey(gid int64) string {
	return fmt.Sprintf("roblox:game:%d:playing", gid)
}

// FriendStatusKey 返回好友状态的键
func (e *extraKey) FriendStatusKey(uid int64) string {
	return fmt.Sprintf("roblox:friend:%d:status", uid)
} 