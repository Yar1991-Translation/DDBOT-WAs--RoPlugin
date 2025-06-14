package roblox

import (
	"encoding/json"
	"fmt"
	"github.com/cnxysoft/DDBOT-WSa/lsp"
	"github.com/cnxysoft/DDBOT-WSa/lsp/concern"
	"github.com/cnxysoft/DDBOT-WSa/lsp/concern_type"
	"strconv"
)

const (
	ServiceName = "roblox"
	UserType    = concern_type.Type("user")
	GameType    = concern_type.Type("game")
)

var log = lsp.Logger.WithField("module", "lsp.roblox")

func init() {
	loadConfig()
	if !config.Enable {
		log.Info("roblox 插件已被禁用")
		return
	}
	lsp.RegisterConcern(ServiceName, func(notifyChan chan<- concern.Notify) concern.Concern {
		return NewRobloxConcern(notifyChan)
	})
	log.Info("roblox 插件已加载")
}

// getUsersPresence 获取用户的在线状态
func getUsersPresence(userIDs []int64) ([]UserPresence, error) {
	// ... existing code ...
} 