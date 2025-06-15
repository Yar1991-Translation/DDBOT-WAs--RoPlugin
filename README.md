# DDBOT-WSa Roblox 插件

> 最近更新：2025-06-15  |  **完整更新日志见 /blog 指令**

本仓库基于 [cnxysoft/DDBOT-WSa](https://github.com/cnxysoft/DDBOT-WSa) 进行二次开发，
重点在于：

* 🚀 **Roblox 订阅插件**（用户状态 / 游戏人数 / 好友上线）
* 🆕 `/roblox` 一站式指令（info / watch / list）
* 📰 **Markdown 更新日志**：Bot 自动拉取 GitHub 博文并渲染为图片推送
* 🔧 兼容原版全部 B 站 / 斗鱼 / Twitter 等订阅功能

如需**直接体验已部署的机器人**，请：

* 添加 QQ：**3069049949**
* 或发送邮件至 **yar20000628@gmail.com**

---

## 快速使用

```bash
# 克隆项目
git clone https://github.com/Yar1991-Translation/DDBOT-WAs--RoPlugin.git
cd DDBOT-WAs--RoPlugin

# 安装依赖（含 lute / chromedp）
go mod tidy

# 构建
GOOS=linux GOARCH=amd64 go build -o ddbot ./cmd

# 运行
./ddbot -config application.yaml
```

> 运行前请确保服务器已安装可执行的 **Chrome / Chromium**，或在环境变量 `CHROME_EXECUTABLE` 指定路径。

---

## 核心指令

| 指令 | 说明 |
|------|------|
| `/roblox info <UID|用户名>` | 查询 Roblox 用户信息 |
| `/roblox watch <user\|game\|friend> <ID>` | 添加订阅 |
| `/roblox unwatch ...` | 取消订阅 |
| `/roblox list` | 查看本群 Roblox 订阅 |
| `/blog [file.md]` | 推送 README 或指定博文的更新日志到群 |

更多通用指令请参考原版 README。

---

<details>
<summary>原版 README（折叠）</summary>

DDBOT-WSa 是基于 DDBOT-ws 的修改版本，目的是恢复DDBOT的原有功能。

DDBOT-WSa 基于 [DDBOT-ws](https://github.com/cnxysoft/DDBOT-ws) 二次开发，致力于「恢复 + 增强」原版全部能力，并带来以下改进：

- 🛠️ **功能补完**：修复并补齐历史指令/事件，开箱即用  
- 🧩 **可扩展模板**：新增模板函数与事件钩子，方便二次开发  
- 🔌 **多协议兼容**：原生支持 LLOnebot / NapCat / Lagrange  
- 🐦 **Twitter 推送**：实验性接入，持续完善中  
- ♻️ **持续维护**：社区驱动的活跃迭代

**该分支所有核心指令均已恢复可用。**

DDBOT是一个基于 [MiraiGO](https://github.com/Mrs4s/MiraiGo) 的QQ群推送框架， 内置支持b站直播/动态，斗鱼直播，YTB直播/预约直播，虎牙直播，ACFUN直播，微博动态，
也可以通过插件支持任何订阅源。

*DDBOT不是一个聊天机器人。*

[sora233的Bilibili专栏](https://www.bilibili.com/read/cv10602230)

-----


## 使用方法

1. 使用 `trss` 安装并启动 「云崽」  
2. 在 `trss` 中连接 ChronoCat (red)  
3. 准备任一 OneBot 协议端：LLOnebot / NapCat / Lagrange  
4. 在云崽安装 `ws-plugin`，并将 WS 地址设为 `ws://127.0.0.1:15630/ws`  
5. 如从纯血 DDBOT 迁移：  
   打开 `lsp.db`，将字段 `ae` 批量替换为 `ex`，完成数据库升级

## 设计理念

制作bot的本意是为了减轻一些重复的工作负担，bot只会做好bot份内的工作：

- ddbot的交互被刻意设计成最小程度，正常交流时永远不必担心会误触ddbot。
- ddbot只有两种情况会主动发言，更新动态和直播，以及答复命令结果。

## **基本功能：**

- **B站 直播 / 动态**：关键字及类型过滤（仅视频 / 专栏 / 含图等）  
- **斗鱼 直播**  
- **YouTube 直播 / 视频**：含预约提醒  
- **虎牙、AcFun 直播**  
- **微博 动态**  
- **自定义插件**：少量代码即可新增任意订阅源  
- **@全体**：可选、可按群启用  
- **娱乐 / 工具**：倒放、Roll、签到 等  
- **权限管理**：按命令 / 按用户 细粒度控制  
- **帮助系统**：内置说明书，一键查询

<details>
  <summary>里命令</summary>

以下命令默认禁用，使用enable命令后才能使用

- **随机图片**
  - 由 [api.lolicon.app](https://api.lolicon.app/#/) 提供

</details>

### 推送效果

<img src="https://user-images.githubusercontent.com/11474360/111737379-78fbe200-88ba-11eb-9e7e-ecc9f2440dd8.jpg" width="300">

### 用法示例

详细介绍及示例请查看：[详细示例](/EXAMPLE.md)

~~阁下可添加官方Demo机器人体验~~

不再提供官方的公开BOT，你可以加入交流群申请使用群友搭建的BOT，也可以选择自己搭建。

## 使用与部署

对于普通用户，推荐您选择使用开放的官方Demo机器人。

您也可以选择私人部署，[详见部署指南](/INSTALL.md)。

私人部署的好处：

- 保护您的隐私，bot完全属于您，我无法得知您bot的任何信息（我甚至无法知道您部署了一个私人bot）
- 稳定的@全体成员功能
- 可定制BOT账号的头像、名字、签名
- 减轻我的服务器负担
- 很cool

如果您遇到任何问题，或者有任何建议，可以加入**交流群：755612788（已满）、980848391**

## 最近更新

请参考[更新文档](/UPDATE.md)。

## 常见问题FAQ

提问前请先查看[FAQ文档](/FAQ.md)，如果仍然未能解决，请咨询指定交流群。

## 增加推送来源 （为DDBOT编写插件）

DDBOT可以作为一个通用的QQ推送框架来使用。

您可以通过为DDBOT编写插件，DDBOT会为您完成大部分工作，您只需要实现少量代码，就能支持一个新的来源。

如果您对此有兴趣，请查看[框架文档](/FRAMEWORK.md) 。

## 自定义消息模板 & 自定义命令回复

DDBOT已实现消息模板功能，一些内置命令和推送可通过模板自定义格式。

同时支持自定义命令，自动回复模板内容。

详细介绍请看[模板文档](/TEMPLATE.md) 。

## 注意事项

- **bot只在群聊内工作，但命令可以私聊使用，以避免在群内刷屏**（少数次要娱乐命令暂不支持，详细列表请看用法指南）
- **建议bot秘密码设置足够强，同时不建议把bot设置为QQ群管理员，因为存在密码被恶意爆破的可能（包括但不限于盗号、广告等）**
- **您应当知道，bot账号可以人工登陆，请注意个人隐私**
- bot掉线无法重连时将自动退出，请自行实现保活机制
- bot使用 [buntdb](https://github.com/tidwall/buntdb) 作为embed database，会在当前目录生成文件`.lsp.db`
  ，删除该文件将导致bot恢复出厂设置，可以使用 [buntdb-cli](https://github.com/Sora233/buntdb-cli) 作为运维工具，但注意不要在bot运行的时候使用（buntdb不支持多写）

## 声明

- 您可以免费使用DDBOT进行其他商业活动，但不允许通过出租、出售DDBOT等方式进行商业活动。
- 如果您运营了私人部署的BOT，可以接受他人对您私人部署的BOT进行捐赠以帮助BOT运行，但该过程必须本着自愿的原则，不允许用BOT使用权来强制他人进行捐赠。
- 如果您使用了DDBOT的源代码，或者对DDBOT源代码进行修改，您应该用相同的开源许可（AGPL3.0）进行开源，并标明著作权。

## 贡献

*Feel free to make your first pull request.*

想要为开源做一点微小的贡献？

[Golang点我入门！](https://github.com/justjavac/free-programming-books-zh_CN#go)

您也可以选择点一下右上角的⭐星⭐

发现问题或功能建议请到 [issues](https://github.com/cnxysoft/DDBOT-WSa/issues)

其他用法问题请到**交流群：755612788（已满）、980848391**

</details>
