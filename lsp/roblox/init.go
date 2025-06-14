package roblox

import (
	"github.com/cnxysoft/DDBOT-WSa/lsp/concern"
	"github.com/Sora233/MiraiGo-Template/utils"
)

var log = utils.GetModuleLogger("lsp.roblox")

const (
	// ServiceName 服务名称
	ServiceName = "roblox"
)

// 初始化函数
func init() {
	// 注册服务
	concern.RegisterConcern(NewRobloxConcern(concern.GetNotifyChan()))

	log.Infof("roblox 插件已加载")
} 