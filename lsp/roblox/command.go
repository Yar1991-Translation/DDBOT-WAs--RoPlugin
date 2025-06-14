package roblox

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/cnxysoft/DDBOT-WSa/lsp"
	"github.com/cnxysoft/DDBOT-WSa/lsp/mmsg"
)

func init() {
	lsp.RegisterCmd(
		"roblox",
		"Roblox 相关功能",
		lsp.NewBuilder().
			Admin().
			AddChild(buildInfoCmd()).
			Build(),
	)
}

func buildInfoCmd() *lsp.Command {
	return lsp.NewBuilder().
		Sub("info").
		Show("查询 Roblox 用户信息").
		Arg("user").
		More("用户ID或用户名").
		Handle(handleInfoCmd).
		Build()
}

func handleInfoCmd(ctx *lsp.CmdContext) {
	if len(ctx.RemovePrefix) < 1 {
		ctx.Session.SendText("请输入要查询的用户 ID 或用户名。")
		return
	}
	identifier := strings.Join(ctx.RemovePrefix, " ")

	var userInfo *UserInfo
	var err error

	// 尝试将输入作为数字ID解析
	if uid, parseErr := strconv.ParseInt(identifier, 10, 64); parseErr == nil {
		userInfo, err = getUserInfo(uid)
	} else {
		// 如果不是数字，则作为用户名处理
		userInfo, err = findUserByName(identifier)
	}

	if err != nil {
		ctx.Session.SendText(fmt.Sprintf("查询失败: %v", err))
		return
	}

	// 获取用户的在线状态
	presenceInfo, pErr := getUsersPresence([]int64{userInfo.ID})
	statusStr := "未知"
	if pErr == nil && len(presenceInfo) > 0 {
		statusStr = getUserStatusString(presenceInfo[0])
	}

	msg := mmsg.NewMSG()
	msg.Textf("查询到 Roblox 用户信息：\n")
	msg.Textf("- 用户名: %s\n", userInfo.Name)
	msg.Textf("- 显示名: %s\n", userInfo.DisplayName)
	msg.Textf("- 用户 ID: %d\n", userInfo.ID)
	msg.Textf("- 当前状态: %s\n", statusStr)
	msg.Textf("- 个人主页: https://www.roblox.com/users/%d/profile", userInfo.ID)

	ctx.Session.Send(msg)
} 