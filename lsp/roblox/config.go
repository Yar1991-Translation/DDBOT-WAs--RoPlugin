package roblox

import (
	"time"

	"github.com/cnxysoft/DDBOT-WSa/lsp/cfg"
)

// Config 是 Roblox 插件的配置结构体
type Config struct {
	Enable   bool          `yaml:"enable"`   // 是否启用插件
	Interval time.Duration `yaml:"interval"` // 轮询间隔
	Proxy    string        `yaml:"proxy"`    // API 代理地址
}

var config = &Config{} // 全局配置实例

// loadConfig 从配置文件中加载配置
func loadConfig() {
	if err := cfg.Get("roblox", config); err != nil {
		log.Fatalf("无法加载 roblox 配置: %v", err)
	}

	// 如果代理地址末尾有 /，则移除
	if len(config.Proxy) > 0 && config.Proxy[len(config.Proxy)-1] == '/' {
		config.Proxy = config.Proxy[:len(config.Proxy)-1]
	}
} 