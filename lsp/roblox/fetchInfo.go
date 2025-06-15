package roblox

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"time"
	"strconv"

	"github.com/pkg/errors"
)

// 统一使用一个可复用的 http.Client，避免频繁创建连接
var httpClient *http.Client

func init() {
	// 创建带有超时与 KeepAlive 的 Transport
	tr := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		IdleConnTimeout:       90 * time.Second,
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   100,
	}

	httpClient = &http.Client{
		Transport: tr,
		Timeout:   15 * time.Second, // 整体超时
	}
}

// apiGet 执行 GET 请求并解析 JSON。
func apiGet(ctx context.Context, url string, result interface{}) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return errors.Wrap(err, "构建 GET 请求失败")
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return errors.Wrap(err, "执行 GET 请求失败")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.Errorf("API 请求失败，状态码: %d", resp.StatusCode)
	}

	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return errors.Wrap(err, "解析响应 JSON 失败")
		}
	}
	return nil
}

// apiPost 执行 POST 请求并解析 JSON。
func apiPost(ctx context.Context, url string, body []byte, result interface{}) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return errors.Wrap(err, "构建 POST 请求失败")
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return errors.Wrap(err, "执行 POST 请求失败")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.Errorf("API 请求失败，状态码: %d", resp.StatusCode)
	}

	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return errors.Wrap(err, "解析响应 JSON 失败")
		}
	}
	return nil
}

// getUserInfo 获取用户信息
func getUserInfo(uid int64) (*UserInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var info UserInfo
	if err := apiGet(ctx, fmt.Sprintf("%s/v1/users/%d", usersAPI, uid), &info); err != nil {
		return nil, err
	}
	return &info, nil
}

// findUserByName 通过用户名查找用户信息
func findUserByName(username string) (*UserInfo, error) {
	requestBody, err := json.Marshal(map[string][]string{
		"usernames": {username},
	})
	if err != nil {
		return nil, errors.Wrap(err, "序列化请求体失败")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var searchResult UserSearchResponse
	if err := apiPost(ctx, fmt.Sprintf("%s/v1/usernames/users", usersAPI), requestBody, &searchResult); err != nil {
		return nil, err
	}

	if len(searchResult.Data) == 0 {
		return nil, errors.Errorf("找不到用户: %s", username)
	}

	foundUser := searchResult.Data[0]
	return &UserInfo{
		ID:          foundUser.ID,
		Name:        foundUser.Name,
		DisplayName: foundUser.DisplayName,
	}, nil
}

// getGameInfo 获取游戏信息
func getGameInfo(gameOrPlaceId int64) ([]GameInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 尝试将 ID 作为 Universe ID 使用
	var data struct {
		Data []GameInfo `json:"data"`
	}
	if err := apiGet(ctx, fmt.Sprintf("%s/v1/games?universeIds=%d", gamesAPI, gameOrPlaceId), &data); err == nil && len(data.Data) > 0 {
		return data.Data, nil
	}

	// 如果失败，假设它是一个 Place ID，获取 Universe ID
	var universe struct {
		UniverseId int64 `json:"universeId"`
	}
	if err := apiGet(ctx, fmt.Sprintf("%s/universes/v1/places/%d/universe", apisAPI, gameOrPlaceId), &universe); err != nil {
		return nil, err
	}

	if universe.UniverseId == 0 {
		return nil, errors.Errorf("无法找到 Universe ID %d", gameOrPlaceId)
	}

	// 使用获取到的 Universe ID 递归查询
	return getGameInfo(universe.UniverseId)
}

// getUsersPresence 获取用户在线状态
func getUsersPresence(uids []int64) ([]UserPresence, error) {
	if len(uids) == 0 {
		return nil, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	body, err := json.Marshal(map[string][]int64{"userIds": uids})
	if err != nil {
		return nil, errors.Wrap(err, "序列化请求体失败")
	}

	var presences struct {
		UserPresences []UserPresence `json:"userPresences"`
	}
	if err := apiPost(ctx, fmt.Sprintf("%s/v1/presence/users", presenceAPI), body, &presences); err != nil {
		return nil, err
	}
	return presences.UserPresences, nil
}

// getUserStatusString 获取用户状态的文字描述
func getUserStatusString(presence UserPresence) string {
	switch presence.UserPresenceType {
	case UserStatusOffline:
		return "离线"
	case UserStatusOnline:
		return "在线"
	case UserStatusInGame:
		return fmt.Sprintf("正在玩 %s", presence.LastLocation)
	case UserStatusInStudio:
		return "在 Studio 中"
	default:
		return "未知状态"
	}
}

// parseID 解析字符串 ID 为整数
func parseID(id interface{}) (int64, error) {
	switch v := id.(type) {
	case int64:
		return v, nil
	case string:
		return strconv.ParseInt(v, 10, 64)
	default:
		return 0, errors.New("无效的 ID 类型")
	}
} 