# DDBOT-WSa (Roblox 插件版)

[![Go build](https://github.com/Yar1991-Translation/DDBOT-WAs--RoPlugin/actions/workflows/ci.yml/badge.svg)](https://github.com/Yar1991-Translation/DDBOT-WAs--RoPlugin/actions/workflows/ci.yml)
[![GoDoc](http://img.shields.io/badge/go-documentation-blue.svg?style=flat-square)](https://pkg.go.dev/github.com/cnxysoft/DDBOT-WSa)

本项目是 [cnxysoft/DDBOT-WSa](https://github.com/cnxysoft/DDBOT-WSa) 的一个分支版本，在其强大的QQ群推送框架基础上，新增了对 **Roblox** 的订阅支持。

DDBOT 是一个基于 [MiraiGO](https://github.com/Mrs4s/MiraiGo) 的 QQ 推送框架，原版支持 B站、斗鱼、YouTube、微博等多个平台。本分支在保留所有原有功能的同时，通过插件机制，扩展了对 Roblox 的支持。

---

## 新增功能：Roblox 订阅插件

本插件允许您在 QQ 群内订阅 Roblox 游戏和用户，以便及时获取更新通知。

### 主要功能

- **订阅 Roblox 用户**：监控指定用户的在线状态（离线、在线、游戏中、Studio中），并在状态变化时发送通知。
- **订阅 Roblox 游戏**：监控指定游戏的在线玩家数量，并在人数变化时发送通知。

### 配置方法

您可以通过修改根目录下的 `application.yaml` 文件来配置 Roblox 插件。

```yaml
# application.yaml

# ... 其他配置 ...

# Roblox 插件配置
roblox:
  enable: true # 设置为 true 启用插件，false 禁用
  interval: "30s" # 检查更新的频率，例如 "30s", "1m", "5m"
  proxy: "https://roproxy.com" # API 代理地址，可替换为其他兼容的代理
```

### 使用指令

您可以在群聊或私聊中通过以下指令与 Bot 互动：

- **订阅用户/游戏**:
  ```
  /watch roblox [user|game] <ID>
  ```
  示例:
  - `/watch roblox user 123456` (订阅 ID 为 123456 的用户)
  - `/watch roblox game 987654` (订阅 ID 为 987654 的游戏)

- **取消订阅**:
  ```
  /unwatch roblox [user|game] <ID>
  ```
  示例:
  - `/unwatch roblox user 123456`

- **查看订阅列表**:
  ```
  /list roblox
  ```

---

## 快速开始

### 1. 环境准备

- 安装 [Go (1.18 或更高版本)](https://go.dev/dl/)
- 一个可用的 QQ Bot 框架 (如 [LLOnebot](https://llonebot.github.io/), [NapCat](https://napneko.github.io/), [Lagrange](https://lagrangedev.github.io/Lagrange.Doc/))

### 2. 下载与运行

首先，克隆本仓库：
```bash
git clone https://github.com/Yar1991-Translation/DDBOT-WAs--RoPlugin.git
cd DDBOT-WAs--RoPlugin
```

然后，构建并运行程序：
```bash
# 构建
go build -o ddbot.exe ./cmd/main.go

# 运行 (Windows)
./ddbot.exe

# 运行 (Linux/macOS)
./ddbot
```
首次运行前，请确保已根据您的 Bot 框架配置好 `application.yaml` 中的 `websocket` 部分。

## 原版 README

以下为原版 `DDBOT-WSa` 项目的 README 内容，包含了对项目设计理念、基础功能、插件开发等方面的详细介绍。

---

<details>
<summary>点击展开原版 README</summary>

DDBOT-WSa 是基于 DDBOT-ws 的修改版本，目的是恢复DDBOT的原有功能。
新增的模板函数以及事件（触发）等其它更详细的更动见更新日志和[DDBOT部署教程](https://ddbot.songlist.icu)。
目前已经兼容：LLOnebot / NapCat / Lagrange。
新增对推特推送的支持（实验阶段）

**目前已经修复所有的主要指令（奇奇怪怪的指令没测试）。**

DDBOT是一个基于 [MiraiGO](https://github.com/Mrs4s/MiraiGo) 的QQ群推送框架， 内置支持b站直播/动态，斗鱼直播，YTB直播/预约直播，虎牙直播，ACFUN直播，微博动态，
也可以通过插件支持任何订阅源。

*DDBOT不是一个聊天机器人。*

[sora233的Bilibili专栏](https://www.bilibili.com/read/cv10602230)

-----


## 使用方法

 - 使用trss安装云崽
   
 - 使用trss连接chronocat(red)

 - 使用LLOnebot / NapCat / Lagrange连接
   
 - 云崽安装ws-plugin
   
 - 云崽ws插件设置连接ddbot的ws地址
   
 - ws://127.0.0.1:15630/ws

 - **从纯血DDBOT迁移到WSa**
 - 打开lsp.db的文件用右键记事本开启或vs/vscode打开 搜索ae字段全部替换为ex字段 只需要把ae改成ex就可以成功初始化数据库（可能会at全体失效建议重新配置）
   

## 设计理念

制作bot的本意是为了减轻一些重复的工作负担，bot只会做好bot份内的工作：

- ddbot的交互被刻意设计成最小程度，正常交流时永远不必担心会误触ddbot。
- ddbot只有两种情况会主动发言，更新动态和直播，以及答复命令结果。

## **基本功能：**

- **B站直播/动态推送**
  - 让阁下在DD的时候不错过任何一场突击。
  - 支持按关键字过滤，只推送有关键字的动态。
  - 支持按动态类型过滤，例如：不推送转发的动态，只推送视频/专栏投稿，只推动带图片的动态等等。
- **斗鱼直播推送**
  - 没什么用，主要用来看爽哥。
- **油管直播/视频推送**
  - 支持推送预约直播信息及视频更新。
- **虎牙直播推送**
  - 不知道能看谁。
- **ACFUN直播推送**
  - 好像也有一些虚拟主播
- **微博动态推送**
- 支持自定义**插件**，可通过插件支持任意订阅来源
  - 需要写代码
- 可配置的 **@全体成员**
  - 只建议单推群开启。
- **倒放**
  - 主要用来玩。
- **Roll**
  - 没什么用的roll点。
- **签到**
  - 没什么用的签到。
- **权限管理**
  - 可配置整个命令的启用和禁用，也可对单个用户配置命令权限，防止滥用。
- **帮助**
  - 输出一些没什么帮助的信息。

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
