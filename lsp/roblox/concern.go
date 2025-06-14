package roblox

import (
	"fmt"
	"time"

	"github.com/cnxysoft/DDBOT-WSa/lsp/concern"
	"github.com/cnxysoft/DDBOT-WSa/lsp/concern_type"
	"github.com/cnxysoft/DDBOT-WSa/lsp/mmsg"
	"github.com/pkg/errors"
)

// RobloxConcern 实现 concern.Concern 接口
type RobloxConcern struct {
	*concern.StateManager
	*extraKey
}

// NewRobloxConcern 创建新的 RobloxConcern 实例
func NewRobloxConcern(notifyChan chan<- concern.Notify) concern.Concern {
	c := &RobloxConcern{
		StateManager: concern.NewStateManagerWithStringID(ServiceName, notifyChan),
		extraKey:     &extraKey{},
	}
	c.UseEmitQueue()
	c.UseFreshFunc(c.EmitQueueFresher(c.fresh))
	c.UseNotifyGeneratorFunc(c.notifyGenerator())
	return c
}

// Site 返回服务的唯一标识符
func (c *RobloxConcern) Site() string {
	return ServiceName
}

// Types 返回支持的订阅类型
func (c *RobloxConcern) Types() []concern_type.Type {
	return []concern_type.Type{UserType, GameType}
}

// ParseId 解析ID
func (c *RobloxConcern) ParseId(s string) (interface{}, error) {
	return s, nil
}

// Add 添加订阅
func (c *RobloxConcern) Add(ctx mmsg.IMsgCtx, groupCode int64, id interface{}, ctype concern_type.Type) (concern.IdentityInfo, error) {
	switch ctype {
	case UserType:
		uid, err := parseID(id)
		if err != nil {
			return nil, errors.Wrap(err, "无效的用户 ID")
		}
		info, err := getUserInfo(uid)
		if err != nil {
			return nil, errors.Wrap(err, "获取用户信息失败")
		}
		_, err = c.StateManager.AddGroupConcern(groupCode, id, ctype)
		if err != nil {
			return nil, err
		}
		return concern.NewIdentity(id, info.Name), nil
	case GameType:
		gid, err := parseID(id)
		if err != nil {
			return nil, errors.Wrap(err, "无效的游戏 ID")
		}
		infos, err := getGameInfo(gid)
		if err != nil || len(infos) == 0 {
			return nil, errors.Wrap(err, "获取游戏信息失败")
		}
		info := infos[0]
		_, err = c.StateManager.AddGroupConcern(groupCode, id, ctype)
		if err != nil {
			return nil, err
		}
		return concern.NewIdentity(id, info.Name), nil
	default:
		return nil, errors.New("不支持的订阅类型，请使用 'user' 或 'game'")
	}
}

// Remove 取消订阅
func (c *RobloxConcern) Remove(ctx mmsg.IMsgCtx, groupCode int64, id interface{}, ctype concern_type.Type) (concern.IdentityInfo, error) {
	identity, _ := c.Get(id)
	_, err := c.StateManager.RemoveGroupConcern(groupCode, id, ctype)
	return identity, err
}

// Get 获取身份信息
func (c *RobloxConcern) Get(id interface{}) (concern.IdentityInfo, error) {
	switch v := id.(type) {
	case string:
		// 尝试获取用户信息
		uid, err := parseID(id)
		if err == nil {
			info, err := getUserInfo(uid)
			if err == nil {
				return concern.NewIdentity(id, info.Name), nil
			}
		}
		
		// 尝试获取游戏信息
		gid, err := parseID(id)
		if err == nil {
			infos, err := getGameInfo(gid)
			if err == nil && len(infos) > 0 {
				return concern.NewIdentity(id, infos[0].Name), nil
			}
		}
		
		// 如果无法获取具体信息，返回ID作为名称
		return concern.NewIdentity(id, v), nil
	default:
		return concern.NewIdentity(id, fmt.Sprint(id)), nil
	}
}

// fresh 实现轮询逻辑，检查用户状态和游戏信息变化
func (c *RobloxConcern) fresh(ctype concern_type.Type, id interface{}) ([]concern.Event, error) {
	switch ctype {
	case UserType:
		return c.freshUserStatus(id)
	case GameType:
		return c.freshGameInfo(id)
	default:
		return nil, errors.New("不支持的订阅类型")
	}
}

// freshUserStatus 刷新用户状态
func (c *RobloxConcern) freshUserStatus(id interface{}) ([]concern.Event, error) {
	var events []concern.Event
	
	uid, err := parseID(id)
	if err != nil {
		log.Errorf("解析用户ID失败: %v - %v", id, err)
		return nil, errors.Wrap(err, "无效的用户 ID")
	}
	
	log.Infof("正在检查用户 %v 的状态", uid)
	
	presences, err := getUsersPresence([]int64{uid})
	if err != nil || len(presences) == 0 {
		log.Errorf("获取用户 %v 在线状态失败: %v", uid, err)
		return nil, errors.Wrap(err, "获取用户在线状态失败")
	}
	presence := presences[0]
	
	statusStr := getUserStatusString(presence)
	
	info, err := getUserInfo(uid)
	if err != nil {
		log.Errorf("获取用户 %v 信息失败: %v", uid, err)
		return nil, errors.Wrap(err, "获取用户信息失败")
	}

	lastStatus, _ := c.StateManager.Get(c.UserStatusKey(uid))
	log.Infof("用户 %v (%s) 当前状态: %s, 上次状态: %s", uid, info.Name, statusStr, lastStatus)
	
	// 强制发送通知，无论状态是否变化
	forceNotify := true // 设置为true强制发送通知，便于调试
	
	if forceNotify || lastStatus != statusStr {
		log.Infof("用户 %v (%s) 状态已变化: %s -> %s, 生成通知事件", uid, info.Name, lastStatus, statusStr)
		event := &UserStatusEvent{
			UserID:      uid,
			UserName:    info.Name,
			Status:      statusStr,
			LastStatus:  lastStatus,
			Time:        time.Now(),
			ProfileLink: fmt.Sprintf("https://www.roblox.com/users/%d/profile", uid),
		}
		events = append(events, event)
		c.StateManager.Set(c.UserStatusKey(uid), statusStr)
	}
	
	return events, nil
}

// freshGameInfo 刷新游戏信息
func (c *RobloxConcern) freshGameInfo(id interface{}) ([]concern.Event, error) {
	var events []concern.Event
	
	gid, err := parseID(id)
	if err != nil {
		log.Errorf("解析游戏ID失败: %v - %v", id, err)
		return nil, errors.Wrap(err, "无效的游戏 ID")
	}
	
	log.Infof("正在检查游戏 %v 的信息", gid)
	
	infos, err := getGameInfo(gid)
	if err != nil || len(infos) == 0 {
		log.Errorf("获取游戏 %v 信息失败: %v", gid, err)
		return nil, errors.Wrap(err, "获取游戏信息失败")
	}
	info := infos[0]

	lastPlaying, _ := c.StateManager.GetInt64(c.GamePlayingKey(info.ID))
	log.Infof("游戏 %v (%s) 当前在线人数: %d, 上次在线人数: %d", info.ID, info.Name, info.Playing, lastPlaying)
	
	// 强制发送通知，无论数据是否变化
	forceNotify := true // 设置为true强制发送通知，便于调试
	
	if forceNotify || lastPlaying != info.Playing {
		log.Infof("游戏 %v (%s) 在线人数已变化: %d -> %d, 生成通知事件", info.ID, info.Name, lastPlaying, info.Playing)
		event := &GamePlayingEvent{
			GameID:      info.ID,
			GameName:    info.Name,
			Playing:     info.Playing,
			LastPlaying: lastPlaying,
			Time:        time.Now(),
			GameLink:    fmt.Sprintf("https://www.roblox.com/games/%d", info.ID),
		}
		events = append(events, event)
		c.StateManager.SetInt64(c.GamePlayingKey(info.ID), info.Playing)
	}
	
	return events, nil
}

// Start 启动服务
func (c *RobloxConcern) Start() error {
	return nil
}

// Stop 停止服务
func (c *RobloxConcern) Stop() {
	// 无需特殊清理
}

// GetStateManager 获取状态管理器
func (c *RobloxConcern) GetStateManager() concern.IStateManager {
	return c.StateManager
} 