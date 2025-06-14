package roblox

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/pkg/errors"
)

// getUserInfo 获取用户信息
func getUserInfo(uid int64) (*UserInfo, error) {
	resp, err := http.Get(fmt.Sprintf("%s/v1/users/%d", usersAPI, uid))
	if err != nil {
		return nil, errors.Wrap(err, "请求用户信息失败")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("API 请求失败，状态码: %d", resp.StatusCode)
	}

	var info UserInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, errors.Wrap(err, "解析用户信息失败")
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

	resp, err := http.Post(fmt.Sprintf("%s/v1/usernames/users", usersAPI), "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, errors.Wrap(err, "请求用户信息失败")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("API 请求失败, 状态码: %d", resp.StatusCode)
	}

	var searchResult UserSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResult); err != nil {
		return nil, errors.Wrap(err, "解析用户信息失败")
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
	// 首先尝试将 ID 作为 Universe ID 使用
	resp, err := http.Get(fmt.Sprintf("%s/v1/games?universeIds=%d", gamesAPI, gameOrPlaceId))
	if err != nil {
		return nil, errors.Wrap(err, "请求游戏信息失败")
	}
	defer resp.Body.Close()

	var data struct {
		Data []GameInfo `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err == nil && len(data.Data) > 0 {
		return data.Data, nil
	}

	// 如果失败，假设它是一个 Place ID，获取 Universe ID
	resp, err = http.Get(fmt.Sprintf("%s/universes/v1/places/%d/universe", apisAPI, gameOrPlaceId))
	if err != nil {
		return nil, errors.Wrap(err, "请求 Universe ID 失败")
	}
	defer resp.Body.Close()

	var universe struct {
		UniverseId int64 `json:"universeId"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&universe); err != nil {
		return nil, errors.Wrap(err, "解析 Universe ID 失败")
	}

	if universe.UniverseId == 0 {
		return nil, errors.Errorf("无法找到 Universe ID %d", gameOrPlaceId)
	}

	// 使用获取到的 Universe ID 重试
	return getGameInfo(universe.UniverseId)
}

// getUsersPresence 获取用户在线状态
func getUsersPresence(uids []int64) ([]UserPresence, error) {
	if len(uids) == 0 {
		return nil, nil
	}

	body, err := json.Marshal(map[string][]int64{"userIds": uids})
	if err != nil {
		return nil, errors.Wrap(err, "序列化请求体失败")
	}

	resp, err := http.Post(fmt.Sprintf("%s/v1/presence/users", presenceAPI), "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, errors.Wrap(err, "请求用户在线状态失败")
	}
	defer resp.Body.Close()

	var presences struct {
		UserPresences []UserPresence `json:"userPresences"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&presences); err != nil {
		return nil, errors.Wrap(err, "解析用户在线状态失败")
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