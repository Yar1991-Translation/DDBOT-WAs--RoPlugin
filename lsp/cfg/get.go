package cfg

import "github.com/Sora233/MiraiGo-Template/config"

// Get 从全局配置中解码指定键到目标结构体。如果对应键不存在或反序列化失败，则返回错误。
func Get(key string, out interface{}) error {
    if key == "" || out == nil {
        return nil
    }
    return config.GlobalConfig.UnmarshalKey(key, out)
} 