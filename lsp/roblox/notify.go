package roblox

import (
	"fmt"

	"github.com/cnxysoft/DDBOT-WSa/lsp/concern"
	"github.com/cnxysoft/DDBOT-WSa/lsp/mmsg"
	"github.com/sirupsen/logrus"
)

// UserStatusNotify 用户状态通知
type UserStatusNotify struct {
	*UserStatusEvent
	groupCode int64
}

// GetGroupCode 获取群号
func (n *UserStatusNotify) GetGroupCode() int64 {
	return n.groupCode
}

// ToMessage 生成通知消息
func (n *UserStatusNotify) ToMessage() *mmsg.MSG {
	m := mmsg.NewMSG()
	m.Textf("Roblox 用户 %s 状态更新\n", n.UserName)
	m.Textf("当前状态: %s\n", n.Status)
	m.Textf("个人主页: %s", n.ProfileLink)
	return m
}

// GamePlayingNotify 游戏在线人数通知
type GamePlayingNotify struct {
	*GamePlayingEvent
	groupCode int64
}

// GetGroupCode 获取群号
func (n *GamePlayingNotify) GetGroupCode() int64 {
	return n.groupCode
}

// ToMessage 生成通知消息
func (n *GamePlayingNotify) ToMessage() *mmsg.MSG {
	m := mmsg.NewMSG()
	m.Textf("Roblox 游戏 %s 信息更新\n", n.GameName)
	m.Textf("在线人数: %d", n.Playing)
	if n.LastPlaying > 0 {
		var changeSymbol string
		var changeValue int64
		
		if n.Playing > n.LastPlaying {
			changeSymbol = "+"
			changeValue = n.Playing - n.LastPlaying
		} else {
			changeSymbol = "-"
			changeValue = n.LastPlaying - n.Playing
		}
		
		m.Textf(" (%s%d)\n", changeSymbol, changeValue)
	} else {
		m.Text("\n")
	}
	m.Textf("游戏链接: %s", n.GameLink)
	return m
}

// FriendOnlineNotify 好友上线通知
type FriendOnlineNotify struct {
	*FriendOnlineEvent
	groupCode int64
}

func (n *FriendOnlineNotify) GetGroupCode() int64 {
	return n.groupCode
}

func (n *FriendOnlineNotify) ToMessage() *mmsg.MSG {
	m := mmsg.NewMSG()
	m.Textf("Roblox 好友 %s 上线啦！\n", n.FriendName)
	m.Textf("当前状态: %s\n", n.Status)
	m.Textf("个人主页: %s", n.ProfileLink)
	return m
}

// NotifyGenerator 生成通知
func (c *RobloxConcern) notifyGenerator() concern.NotifyGeneratorFunc {
	return func(groupCode int64, event concern.Event) []concern.Notify {
		switch e := event.(type) {
		case *UserStatusEvent:
			return []concern.Notify{
				&UserStatusNotify{
					UserStatusEvent: e,
					groupCode:       groupCode,
				},
			}
		case *GamePlayingEvent:
			return []concern.Notify{
				&GamePlayingNotify{
					GamePlayingEvent: e,
					groupCode:        groupCode,
				},
			}
		case *FriendOnlineEvent:
			return []concern.Notify{
				&FriendOnlineNotify{
					FriendOnlineEvent: e,
					groupCode:         groupCode,
				},
			}
		}
		log.WithFields(logrus.Fields{
			"GroupCode": groupCode,
			"Event":     fmt.Sprintf("%T", event),
		}).Error("unknown event type")
		return nil
	}
} 