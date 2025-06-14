package lsp

import (
	"fmt"
	"os"
	"runtime/debug"
	"strconv"
	"sync"
	"time"

	"github.com/Mrs4s/MiraiGo/client"
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/Sora233/MiraiGo-Template/bot"
	"github.com/Sora233/MiraiGo-Template/config"
	"github.com/Sora233/sliceutil"
	"github.com/cnxysoft/DDBOT-WSa/image_pool"
	"github.com/cnxysoft/DDBOT-WSa/image_pool/local_pool"
	"github.com/cnxysoft/DDBOT-WSa/image_pool/lolicon_pool"
	localdb "github.com/cnxysoft/DDBOT-WSa/lsp/buntdb"
	"github.com/cnxysoft/DDBOT-WSa/lsp/cfg"
	"github.com/cnxysoft/DDBOT-WSa/lsp/concern"
	"github.com/cnxysoft/DDBOT-WSa/lsp/concern_type"
	"github.com/cnxysoft/DDBOT-WSa/lsp/mmsg"
	"github.com/cnxysoft/DDBOT-WSa/lsp/permission"
	"github.com/cnxysoft/DDBOT-WSa/lsp/template"
	"github.com/cnxysoft/DDBOT-WSa/lsp/version"
	"github.com/cnxysoft/DDBOT-WSa/proxy_pool"
	"github.com/cnxysoft/DDBOT-WSa/proxy_pool/local_proxy_pool"
	"github.com/cnxysoft/DDBOT-WSa/proxy_pool/py"
	localutils "github.com/cnxysoft/DDBOT-WSa/utils"
	"github.com/cnxysoft/DDBOT-WSa/utils/msgstringer"
	"github.com/fsnotify/fsnotify"
	jsoniter "github.com/json-iterator/go"
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
	"github.com/tidwall/buntdb"
	"go.uber.org/atomic"
	"golang.org/x/sync/semaphore"
)

const ModuleName = "me.sora233.Lsp"

var logger = logrus.WithField("module", ModuleName)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

var Debug = false

type Lsp struct {
	pool          image_pool.Pool
	concernNotify <-chan concern.Notify
	stop          chan interface{}
	wg            sync.WaitGroup
	status        *Status
	notifyWg      sync.WaitGroup
	msgLimit      *semaphore.Weighted
	cron          *cron.Cron

	PermissionStateManager *permission.StateManager
	LspStateManager        *StateManager
	started                atomic.Bool
}

func (l *Lsp) CommandShowName(command string) string {
	return cfg.GetCommandPrefix(command) + command
}

func (l *Lsp) MiraiGoModule() bot.ModuleInfo {
	return bot.ModuleInfo{
		ID:       ModuleName,
		Instance: Instance,
	}
}

func (l *Lsp) Init() {
	log := logger.WithField("log_level", config.GlobalConfig.GetString("logLevel"))
	lev, err := logrus.ParseLevel(config.GlobalConfig.GetString("logLevel"))
	if err != nil {
		logrus.SetLevel(logrus.DebugLevel)
		log.Warn("无法识别logLevel，将使用Debug级别")
	} else {
		logrus.SetLevel(lev)
		log.Infof("设置logLevel为%v", lev.String())
	}

	l.msgLimit = semaphore.NewWeighted(int64(cfg.GetNotifyParallel()))

	if Tags != "UNKNOWN" {
		logger.Infof("DDBOT版本：Release版本【%v】", Tags)
	} else {
		if CommitId == "UNKNOWN" {
			logger.Infof("DDBOT版本：编译版本未知")
		} else {
			logger.Infof("DDBOT版本：编译版本【%v-%v】", BuildTime, CommitId)
		}
	}

	db := localdb.MustGetClient()
	var count int
	err = db.View(func(tx *buntdb.Tx) error {
		return tx.Ascend("", func(key, value string) bool {
			count++
			return true
		})
	})
	if err == nil && count == 0 {
		if _, err := version.SetVersion(LspVersionName, LspSupportVersion); err != nil {
			log.Fatalf("警告：初始化LspVersion失败！")
		}
	} else {
		curVersion := version.GetCurrentVersion(LspVersionName)
		if curVersion < 0 {
			log.Errorf("警告：无法检查数据库兼容性，程序可能无法正常工作")
		} else if curVersion > LspSupportVersion {
			log.Fatalf("警告：检查数据库兼容性失败！最高支持版本：%v，当前版本：%v", LspSupportVersion, curVersion)
		} else if curVersion < LspSupportVersion {
			// 应该更新下
			backupFileName := fmt.Sprintf("%v-%v", localdb.LSPDB, time.Now().Unix())
			log.Warnf(
				`警告：数据库兼容性检查完毕，当前需要从<%v>更新至<%v>，将备份当前数据库文件到"%v"`,
				curVersion, LspSupportVersion, backupFileName)
			f, err := os.Create(backupFileName)
			if err != nil {
				log.Fatalf(`无法创建备份文件<%v>：%v`, backupFileName, err)
			}
			err = db.Save(f)
			if err != nil {
				log.Fatalf(`无法备份数据库到<%v>：%v`, backupFileName, err)
			}
			log.Infof(`备份完成，已备份数据库到<%v>"`, backupFileName)
			log.Info("五秒后将开始更新数据库，如需取消请按Ctrl+C")
			time.Sleep(time.Second * 5)
			err = version.DoMigration(LspVersionName, lspMigrationMap)
			if err != nil {
				log.Fatalf("更新数据库失败：%v", err)
			}
		} else {
			log.Debugf("数据库兼容性检查完毕，当前已为最新模式：%v", curVersion)
		}
	}

	imagePoolType := config.GlobalConfig.GetString("imagePool.type")
	log = logger.WithField("image_pool_type", imagePoolType)

	switch imagePoolType {
	case "loliconPool":
		pool, err := lolicon_pool.NewLoliconPool(&lolicon_pool.Config{
			ApiKey:   config.GlobalConfig.GetString("loliconPool.apikey"),
			CacheMin: config.GlobalConfig.GetInt("loliconPool.cacheMin"),
			CacheMax: config.GlobalConfig.GetInt("loliconPool.cacheMax"),
		})
		if err != nil {
			log.Errorf("can not init pool %v", err)
		} else {
			l.pool = pool
			log.Infof("初始化%v图片池", imagePoolType)
			l.status.ImagePoolEnable = true
		}
	case "localPool":
		pool, err := local_pool.NewLocalPool(config.GlobalConfig.GetString("localPool.imageDir"))
		if err != nil {
			log.Errorf("初始化%v图片池失败 %v", imagePoolType, err)
		} else {
			l.pool = pool
			log.Infof("初始化%v图片池", imagePoolType)
			l.status.ImagePoolEnable = true
		}
	case "off":
		log.Debug("关闭图片池")
	default:
		log.Errorf("未知的图片池")
	}

	proxyType := config.GlobalConfig.GetString("proxy.type")
	log = logger.WithField("proxy_type", proxyType)
	switch proxyType {
	case "pyProxyPool":
		host := config.GlobalConfig.GetString("pyProxyPool.host")
		log := log.WithField("host", host)
		pyPool, err := py.NewPYProxyPool(host)
		if err != nil {
			log.Errorf("init py pool err %v", err)
		} else {
			proxy_pool.Init(pyPool)
			l.status.ProxyPoolEnable = true
		}
	case "localProxyPool":
		overseaProxies := config.GlobalConfig.GetStringSlice("localProxyPool.oversea")
		mainlandProxies := config.GlobalConfig.GetStringSlice("localProxyPool.mainland")
		var proxies []*local_proxy_pool.Proxy
		for _, proxy := range overseaProxies {
			proxies = append(proxies, &local_proxy_pool.Proxy{
				Type:  proxy_pool.PreferOversea,
				Proxy: proxy,
			})
		}
		for _, proxy := range mainlandProxies {
			proxies = append(proxies, &local_proxy_pool.Proxy{
				Type:  proxy_pool.PreferMainland,
				Proxy: proxy,
			})
		}
		pool := local_proxy_pool.NewLocalPool(proxies)
		proxy_pool.Init(pool)
		log.WithField("local_proxy_num", len(proxies)).Debug("debug")
		l.status.ProxyPoolEnable = true
	case "off":
		log.Debug("proxy pool turn off")
	default:
		log.Errorf("unknown proxy type")
	}
	if cfg.GetTemplateEnabled() {
		log.Infof("已启用模板")
		template.InitTemplateLoader()
	}
	cfg.ReloadCustomCommandPrefix()
	config.GlobalConfig.OnConfigChange(func(in fsnotify.Event) {
		go cfg.ReloadCustomCommandPrefix()
		l.CronjobReload()
	})
}

func (l *Lsp) PostInit() {
}

func (l *Lsp) DebugCheck(groupCode int64, uin int64, isGroupMessage bool) bool {
	var ok bool
	if Debug {
		if isGroupMessage {
			if sliceutil.Contains(config.GlobalConfig.GetStringSlice("debug.group"), strconv.FormatInt(groupCode, 10)) {
				ok = true
			}
		}
		if sliceutil.Contains(config.GlobalConfig.GetStringSlice("debug.uin"), strconv.FormatInt(uin, 10)) {
			ok = true
		}
	} else {
		ok = true
	}
	return ok
}

func (l *Lsp) Serve(bot *bot.Bot) {
	bot.GroupMemberJoinEvent.Subscribe(func(qqClient *client.QQClient, event *client.MemberJoinGroupEvent) {
		if err := localdb.Set(localdb.Key("OnGroupMemberJoined", event.Group.Code, event.Member.Uin, event.Member.JoinTime), "",
			localdb.SetExpireOpt(time.Minute*2), localdb.SetNoOverWriteOpt()); err != nil {
			return
		}
		m, _ := template.LoadAndExec("trigger.group.member_in.tmpl", map[string]interface{}{
			"group_code":  event.Group.Code,
			"group_name":  event.Group.Name,
			"member_code": event.Member.Uin,
			"member_name": event.Member.DisplayName(),
		})
		if m != nil && l.DebugCheck(event.Group.Code, event.Member.Uin, true) {
			l.SendMsg(m, mmsg.NewGroupTarget(event.Group.Code))
		}
	})
	bot.GroupMemberLeaveEvent.Subscribe(func(qqClient *client.QQClient, event *client.MemberLeaveGroupEvent) {
		if err := localdb.Set(localdb.Key("OnGroupMemberLeaved", event.Group.Code, event.Member.Uin, event.Member.JoinTime), "",
			localdb.SetExpireOpt(time.Minute*2), localdb.SetNoOverWriteOpt()); err != nil {
			return
		}
		m, _ := template.LoadAndExec("trigger.group.member_out.tmpl", map[string]interface{}{
			"group_code":  event.Group.Code,
			"group_name":  event.Group.Name,
			"member_code": event.Member.Uin,
			"member_name": event.Member.DisplayName(),
		})
		if m != nil && l.DebugCheck(event.Group.Code, event.Member.Uin, true) {
			l.SendMsg(m, mmsg.NewGroupTarget(event.Group.Code))
		}
	})
	bot.GroupInvitedEvent.Subscribe(func(qqClient *client.QQClient, request *client.GroupInvitedRequest) {
		log := logger.WithFields(logrus.Fields{
			"GroupCode":   request.GroupCode,
			"GroupName":   request.GroupName,
			"InvitorUin":  request.InvitorUin,
			"InvitorNick": request.InvitorNick,
		})

		if l.PermissionStateManager.CheckBlockList(request.InvitorUin) {
			log.Debug("收到加群邀请，该用户在block列表中，将拒绝加群邀请")
			l.PermissionStateManager.AddBlockList(request.GroupCode, 0)
			request.Reject(false, "")
			return
		}

		requests, err := l.LspStateManager.ListGroupInvitedRequest()
		if err != nil {
			log.Errorf("ListGroupInvitedRequest error - %v", err)
			return
		}
		for _, r := range requests {
			if r.GroupCode == request.GroupCode {
				l.LspStateManager.DeleteGroupInvitedRequest(request.RequestId)
				log.Info("收到加群邀请，该群聊已在申请列表中，将忽略该申请")
				return
			}
		}

		fi := bot.FindFriend(request.InvitorUin)
		if fi == nil {
			log.Error("收到加群邀请，无法找到好友信息，将拒绝加群邀请")
			l.PermissionStateManager.AddBlockList(request.GroupCode, 0)
			request.Reject(false, "未找到阁下的好友信息，请添加好友进行操作")
			return
		}

		if l.PermissionStateManager.CheckAdmin(request.InvitorUin) {
			log.Info("收到管理员的加群邀请，将同意加群邀请")
			l.PermissionStateManager.DeleteBlockList(request.GroupCode)
			request.Accept()
			return
		}

		switch l.LspStateManager.GetCurrentMode() {
		case PrivateMode:
			log.Info("收到加群邀请，当前BOT处于私有模式，将拒绝加群邀请")
			l.PermissionStateManager.AddBlockList(request.GroupCode, 0)
			request.Reject(false, "当前BOT处于私有模式")
		case ProtectMode:
			if err := l.LspStateManager.SaveGroupInvitedRequest(request); err != nil {
				log.Errorf("收到加群邀请，但记录申请失败，将拒绝该申请，请将该问题反馈给开发者 - error %v", err)
				request.Reject(false, "内部错误")
			} else {
				log.Info("收到加群邀请，当前BOT处于审核模式，将保留加群邀请")
			}
		case PublicMode:
			request.Accept()
			l.PermissionStateManager.DeleteBlockList(request.GroupCode)
			log.Info("收到加群邀请，当前BOT处于公开模式，将接受加群邀请")
			m, _ := template.LoadAndExec("trigger.private.group_invited.tmpl", map[string]interface{}{
				"member_code": request.InvitorUin,
				"member_name": request.InvitorNick,
				"group_code":  request.GroupCode,
				"group_name":  request.GroupName,
				"command":     CommandMaps,
			})
			if m != nil {
				l.SendMsg(m, mmsg.NewPrivateTarget(request.InvitorUin))
			}
			if err := l.PermissionStateManager.GrantGroupRole(request.GroupCode, request.InvitorUin, permission.GroupAdmin); err != nil {
				if err != permission.ErrPermissionExist {
					log.Errorf("设置群管理员权限失败 - %v", err)
				}
			}
		default:
			// impossible
			log.Errorf("收到加群邀请，当前BOT处于未知模式，将拒绝加群邀请，请将该问题反馈给开发者")
			request.Reject(false, "内部错误")
		}
	})

	bot.NewFriendRequestEvent.Subscribe(func(qqClient *client.QQClient, request *client.NewFriendRequest) {
		log := logger.WithFields(logrus.Fields{
			"RequesterUin":  request.RequesterUin,
			"RequesterNick": request.RequesterNick,
			"Message":       request.Message,
		})
		if l.PermissionStateManager.CheckBlockList(request.RequesterUin) {
			log.Info("收到好友申请，该用户在block列表中，将拒绝好友申请")
			request.Reject()
			return
		}
		req, err := l.LspStateManager.ListNewFriendRequest()
		if err != nil {
			log.Errorf("ListNewFriendRequest error %v", err)
			return
		}
		for _, r := range req {
			if r.RequesterUin == request.RequesterUin {
				l.LspStateManager.DeleteNewFriendRequest(request.RequestId)
				log.Info("收到好友申请，该用户已在申请列表中，将忽略该申请")
				return
			}
		}
		switch l.LspStateManager.GetCurrentMode() {
		case PrivateMode:
			log.Info("收到好友申请，当前BOT处于私有模式，将拒绝好友申请")
			request.Reject()
		case ProtectMode:
			if err := l.LspStateManager.SaveNewFriendRequest(request); err != nil {
				log.Errorf("收到好友申请，但记录申请失败，将拒绝该申请，请将该问题反馈给开发者 - error %v", err)
				request.Reject()
			} else {
				log.Info("收到好友申请，当前BOT处于审核模式，将保留好友申请")
			}
		case PublicMode:
			log.Info("收到好友申请，当前BOT处于公开模式，将通过好友申请")
			request.Accept()
		default:
			// impossible
			log.Errorf("收到好友申请，当前BOT处于未知模式，将拒绝好友申请，请将该问题反馈给开发者")
			request.Reject()
		}
	})

	bot.NewFriendEvent.Subscribe(func(qqClient *client.QQClient, event *client.NewFriendEvent) {
		log := logger.WithFields(logrus.Fields{
			"Uin":      event.Friend.Uin,
			"Nickname": event.Friend.Nickname,
		})
		log.Info("添加新好友")

		l.LspStateManager.RWCover(func() error {
			requests, err := l.LspStateManager.ListNewFriendRequest()
			if err != nil {
				log.Errorf("ListNewFriendRequest error %v", err)
				return err
			}
			for _, req := range requests {
				if req.RequesterUin == event.Friend.Uin {
					l.LspStateManager.DeleteNewFriendRequest(req.RequestId)
				}
			}
			return nil
		})

		m, _ := template.LoadAndExec("trigger.private.new_friend_added.tmpl", map[string]interface{}{
			"member_code": event.Friend.Uin,
			"member_name": event.Friend.Nickname,
			"command":     CommandMaps,
		})
		if m != nil {
			l.SendMsg(m, mmsg.NewPrivateTarget(event.Friend.Uin))
		}
	})

	bot.GroupJoinEvent.Subscribe(func(qqClient *client.QQClient, info *client.GroupInfo) {
		l.FreshIndex()
		log := logger.WithFields(logrus.Fields{
			"GroupCode":   info.Code,
			"MemberCount": info.MemberCount,
			"GroupName":   info.Name,
			"OwnerUin":    info.OwnerUin,
		})
		log.Info("进入新群聊")

		rename := config.GlobalConfig.GetString("bot.onJoinGroup.rename")
		if len(rename) > 0 {
			if len(rename) > 60 {
				rename = rename[:60]
			}
			minfo := info.FindMember(bot.Uin)
			if minfo != nil {
				minfo.EditCard(rename)
			}
		}

		l.LspStateManager.RWCover(func() error {
			requests, err := l.LspStateManager.ListGroupInvitedRequest()
			if err != nil {
				log.Errorf("ListGroupInvitedRequest error %v", err)
				return err
			}
			for _, req := range requests {
				if req.GroupCode == info.Code {
					if err = l.LspStateManager.DeleteGroupInvitedRequest(req.RequestId); err != nil {
						log.WithField("RequestId", req.RequestId).Errorf("DeleteGroupInvitedRequest error %v", err)
					}
					if err = l.PermissionStateManager.GrantGroupRole(info.Code, req.InvitorUin, permission.GroupAdmin); err != nil {
						if err != permission.ErrPermissionExist {
							log.WithField("target", req.InvitorUin).Errorf("设置群管理员权限失败 - %v", err)
						}
					}
				}
			}
			return nil
		})
	})

	bot.GroupLeaveEvent.Subscribe(func(qqClient *client.QQClient, event *client.GroupLeaveEvent) {
		log := logger.WithField("GroupCode", event.Group.Code).
			WithField("GroupName", event.Group.Name).
			WithField("MemberCount", event.Group.MemberCount)
		for _, c := range concern.ListConcern() {
			_, ids, _, err := c.GetStateManager().ListConcernState(
				func(groupCode int64, id interface{}, p concern_type.Type) bool {
					return groupCode == event.Group.Code
				})
			if err != nil {
				log = log.WithField(fmt.Sprintf("%v订阅", c.Site()), "查询失败")
			} else {
				log = log.WithField(fmt.Sprintf("%v订阅", c.Site()), len(ids))
			}
		}
		if event.Operator == nil {
			log.Info("退出群聊")
		} else {
			log.Infof("被 %v 踢出群聊", event.Operator.DisplayName())
		}
		l.RemoveAllByGroup(event.Group.Code)
	})

	bot.GroupNotifyEvent.Subscribe(func(qqClient *client.QQClient, ievent client.INotifyEvent) {
		switch event := ievent.(type) {
		case *client.GroupPokeNotifyEvent:
			data := map[string]interface{}{
				"member_code":   event.Sender,
				"receiver_code": event.Receiver,
				"group_code":    event.GroupCode,
			}
			if gi := localutils.GetBot().FindGroup(event.GroupCode); gi != nil {
				data["group_name"] = gi.Name
				if fi := gi.FindMember(event.Sender); fi != nil {
					data["member_name"] = fi.DisplayName()
				}
				if fi := gi.FindMember(event.Receiver); fi != nil {
					data["receiver_name"] = fi.DisplayName()
				}
			}
			m, _ := template.LoadAndExec("trigger.group.poke.tmpl", data)
			if m != nil && l.DebugCheck(event.GroupCode, event.Sender, true) {
				l.SendMsg(m, mmsg.NewGroupTarget(event.GroupCode))
			}
		}
	})

	bot.FriendNotifyEvent.Subscribe(func(qqClient *client.QQClient, ievent client.INotifyEvent) {
		switch event := ievent.(type) {
		case *client.FriendPokeNotifyEvent:
			if event.Receiver == localutils.GetBot().GetUin() {
				data := map[string]interface{}{
					"member_code": event.Sender,
				}
				if fi := localutils.GetBot().FindFriend(event.Sender); fi != nil {
					data["member_name"] = fi.Nickname
				}
				m, _ := template.LoadAndExec("trigger.private.poke.tmpl", data)
				if m != nil && l.DebugCheck(0, event.Sender, false) {
					l.SendMsg(m, mmsg.NewPrivateTarget(event.Sender))
				}
			}
		}
	})

	bot.GroupMessageEvent.Subscribe(func(qqClient *client.QQClient, msg *message.GroupMessage) {
		if len(msg.Elements) <= 0 {
			return
		}
		if err := l.LspStateManager.SaveMessageImageUrl(msg.GroupCode, msg.Id, msg.Elements); err != nil {
			logger.Errorf("SaveMessageImageUrl failed %v", err)
		}
		if !l.started.Load() {
			return
		}
		//fmt.Printf("运行到cmd := NewLspGroupCommand(l, msg)啦!")
		//fmt.Printf("%+v\n", msg)
		logger.Debugf("%+v\n", msg)
		cmd := NewLspGroupCommand(l, msg)
		if Debug {
			cmd.Debug()
		}
		if !l.LspStateManager.IsMuted(msg.GroupCode, bot.Uin) {
			go cmd.Execute()
		}
	})

	bot.SelfGroupMessageEvent.Subscribe(func(qqClient *client.QQClient, msg *message.GroupMessage) {
		if len(msg.Elements) <= 0 {
			return
		}
		if err := l.LspStateManager.SaveMessageImageUrl(msg.GroupCode, msg.Id, msg.Elements); err != nil {
			logger.Errorf("SaveMessageImageUrl failed %v", err)
		}
	})

	bot.GroupMuteEvent.Subscribe(func(qqClient *client.QQClient, event *client.GroupMuteEvent) {
		if err := l.LspStateManager.Muted(event.GroupCode, event.TargetUin, event.Time); err != nil {
			logger.Errorf("Muted failed %v", err)
		}
		if event.TargetUin == localutils.GetBot().GetUin() {
			data := map[string]interface{}{
				"group_code":    event.GroupCode,
				"member_code":   event.TargetUin,
				"operator_code": event.OperatorUin,
				"mute_duration": event.Time,
			}
			if gi := localutils.GetBot().FindGroup(event.GroupCode); gi != nil {
				data["group_name"] = gi.Name
				if fi := gi.FindMember(event.TargetUin); fi != nil {
					data["member_name"] = fi.DisplayName()
				}
				if fi := gi.FindMember(event.OperatorUin); fi != nil {
					data["operator_name"] = fi.DisplayName()
				}
			}
			m, _ := template.LoadAndExec("trigger.group.bot_mute.tmpl", data)
			if m != nil {
				if admin := l.PermissionStateManager.ListAdmin(); len(admin) > 0 {
					l.SendMsg(m, mmsg.NewPrivateTarget(admin[0]))
				} else {
					logger.Warn("未设置管理员，取消提示")
				}
			}
		}
	})

	bot.PrivateMessageEvent.Subscribe(func(qqClient *client.QQClient, msg *message.PrivateMessage) {
		if !l.started.Load() {
			return
		}
		if len(msg.Elements) == 0 {
			return
		}
		cmd := NewLspPrivateCommand(l, msg)
		if Debug {
			cmd.Debug()
		}
		go cmd.Execute()
	})
	bot.DisconnectedEvent.Subscribe(func(qqClient *client.QQClient, event *client.ClientDisconnectedEvent) {
		logger.Errorf("收到OnDisconnected事件 %v", event.Message)
		if config.GlobalConfig.GetString("bot.onDisconnected") == "exit" {
			logger.Fatalf("onDisconnected设置为exit，bot将自动退出")
		}
		if err := bot.ReLogin(event); err != nil {
			logger.Fatalf("重连时发生错误%v，bot将自动退出", err)
		}
	})

	bot.MemberCardUpdatedEvent.Subscribe(func(qqClient *client.QQClient, event *client.MemberCardUpdatedEvent) {
		// 群名片更新通知
		data := map[string]interface{}{
			"group_code":      event.Group.Code,
			"group_name":      event.Group.Name,
			"member_code":     event.Member.Uin,
			"old_member_name": event.OldCard,
			"member_name":     event.Member.DisplayName(),
		}
		if event.OldCard == "" {
			data["old_member_name"] = event.Member.Nickname
		}
		// if gi := localutils.GetBot().FindGroup(event.Group.Code); gi != nil {
		// 	data["group_name"] = gi.Name
		// 	if fi := gi.FindMember(event.Member.Uin); fi != nil {
		// 		data["member_name"] = fi.DisplayName()
		// 	}
		// }
		m, _ := template.LoadAndExec("trigger.group.card_updated.tmpl", data)
		if m != nil && l.DebugCheck(event.Group.Code, event.Member.Uin, true) {
			l.SendMsg(m, mmsg.NewGroupTarget(event.Group.Code))
		}
	})

	bot.GroupUploadNotifyEvent.Subscribe(func(qqClient *client.QQClient, event *client.GroupUploadNotifyEvent) {
		data := map[string]interface{}{
			"member_code": event.Sender,
			"group_code":  event.GroupCode,
			"file_name":   event.File.FileName,
			"file_size":   event.File.FileSize,
			"file_id":     event.File.FileId,
			"file_url":    event.File.FileUrl,
			"file_busId":  event.File.BusId,
		}
		if gi := localutils.GetBot().FindGroup(event.GroupCode); gi != nil {
			data["group_name"] = gi.Name
			if fi := gi.FindMember(event.Sender); fi != nil {
				data["member_name"] = fi.DisplayName()
			}
		}
		m, _ := template.LoadAndExec("trigger.group.upload.tmpl", data)
		if m != nil && l.DebugCheck(event.GroupCode, event.Sender, true) {
			l.SendMsg(m, mmsg.NewGroupTarget(event.GroupCode))
		}
	})

	bot.GroupMemberPermissionChangedEvent.Subscribe(func(qqClient *client.QQClient, event *client.MemberPermissionChangedEvent) {
		// 群名片更新通知
		data := map[string]interface{}{
			"group_code":  event.Group.Code,
			"group_name":  event.Group.Name,
			"member_code": event.Member.Uin,
			"member_name": event.Member.DisplayName(),
		}
		permission := func(permission client.MemberPermission) string {
			switch permission {
			case client.Member:
				return "群员"
			case client.Administrator:
				return "管理员"
			case client.Owner:
				return "群主"
			}
			return "未知权限"
		}
		data["old_permission"] = permission(event.OldPermission)
		data["permission"] = permission(event.NewPermission)
		// if gi := localutils.GetBot().FindGroup(event.Group.Code); gi != nil {
		// 	data["group_name"] = gi.Name
		// 	if fi := gi.FindMember(event.Member.Uin); fi != nil {
		// 		data["member_name"] = fi.DisplayName()
		// 	}
		// }
		m, _ := template.LoadAndExec("trigger.group.admin_changed.tmpl", data)
		if m != nil && l.DebugCheck(event.Group.Code, event.Member.Uin, true) {
			l.SendMsg(m, mmsg.NewGroupTarget(event.Group.Code))
		}
	})

	bot.BotOfflineEvent.Subscribe(func(qqClient *client.QQClient, event *client.BotOfflineEvent) {
		templateName := "notify.bot.offline.tmpl"
		data := map[string]interface{}{
			"template_name": templateName,
		}
		_, _ = template.LoadAndExec(templateName, data)
	})

}

func (l *Lsp) PostStart(bot *bot.Bot) {
	l.FreshIndex()
	go func() {
		for range time.Tick(time.Second * 30) {
			l.FreshIndex()
		}
	}()
	l.CronjobReload()
	l.CronStart()
	concern.StartAll()
	l.started.Store(true)

	var newVersionChan = make(chan string, 1)
	go func() {
		newVersionChan <- CheckUpdate()
		for range time.Tick(time.Hour * 24) {
			newVersionChan <- CheckUpdate()
		}
	}()
	go l.NewVersionNotify(newVersionChan)

	logger.Infof("DDBOT启动完成")
	logger.Infof("D宝，一款真正人性化的单推BOT")
	if len(l.PermissionStateManager.ListAdmin()) == 0 {
		logger.Infof("您似乎正在部署全新的BOT，请通过qq对bot私聊发送<%v>(不含括号)获取管理员权限，然后私聊发送<%v>(不含括号)开始使用您的bot",
			l.CommandShowName(WhosyourdaddyCommand), l.CommandShowName(HelpCommand))
	}

}

func (l *Lsp) Start(bot *bot.Bot) {
	go l.ConcernNotify()
}

func (l *Lsp) Stop(bot *bot.Bot, wg *sync.WaitGroup) {
	defer wg.Done()
	if l.stop != nil {
		close(l.stop)
	}
	l.CronStop()
	concern.StopAll()

	l.wg.Wait()
	logger.Debug("等待所有推送发送完毕")
	l.notifyWg.Wait()
	logger.Debug("推送发送完毕")

	proxy_pool.Stop()
}

func (l *Lsp) NewVersionNotify(newVersionChan <-chan string) {
	defer func() {
		if err := recover(); err != nil {
			logger.WithField("stack", string(debug.Stack())).
				Errorf("new version notify recoverd %v", err)
			go l.NewVersionNotify(newVersionChan)
		}
	}()
	for newVersion := range newVersionChan {
		if newVersion == "" {
			continue
		}
		var newVersionNotify bool
		err := localdb.RWCover(func() error {
			key := localdb.DDBotReleaseKey()
			releaseVersion, err := localdb.Get(key, localdb.IgnoreNotFoundOpt())
			if err != nil {
				return err
			}
			if releaseVersion != newVersion {
				newVersionNotify = true
			}
			return localdb.Set(key, newVersion)
		})
		if err != nil {
			logger.Errorf("NewVersionNotify error %v", err)
			continue
		}
		if !newVersionNotify {
			continue
		}
		m := mmsg.NewMSG()
		m.Textf("DDBOT管理员您好，DDBOT有可用更新版本【%v】，请前往 https://github.com/cnxysoft/DDBOT-WSa/releases 查看详细信息\n\n", newVersion)
		m.Textf("如果您不想接收更新消息，请输入<%v>(不含括号)", l.CommandShowName(NoUpdateCommand))
		for _, admin := range l.PermissionStateManager.ListAdmin() {
			if localdb.Exist(localdb.DDBotNoUpdateKey(admin)) {
				continue
			}
			if localutils.GetBot().FindFriend(admin) == nil {
				continue
			}
			logger.WithField("Target", admin).Infof("new ddbot version notify")
			l.SendMsg(m, mmsg.NewPrivateTarget(admin))
		}
	}
}

func (l *Lsp) FreshIndex() {
	for _, c := range concern.ListConcern() {
		c.FreshIndex()
	}
	l.PermissionStateManager.FreshIndex()
	l.LspStateManager.FreshIndex()
}

func (l *Lsp) RemoveAllByGroup(groupCode int64) {
	for _, c := range concern.ListConcern() {
		c.GetStateManager().RemoveAllByGroupCode(groupCode)
	}
	l.PermissionStateManager.RemoveAllByGroupCode(groupCode)
}

func (l *Lsp) GetImageFromPool(options ...image_pool.OptionFunc) ([]image_pool.Image, error) {
	if l.pool == nil {
		return nil, image_pool.ErrNotInit
	}
	return l.pool.Get(options...)
}

func (l *Lsp) send(msg *message.SendingMessage, target mmsg.Target) interface{} {
	switch target.TargetType() {
	case mmsg.TargetGroup:
		return l.sendGroupMessage(target.TargetCode(), msg)
	case mmsg.TargetPrivate:
		return l.sendPrivateMessage(target.TargetCode(), msg)
	}
	panic("unknown target type")
}

// SendMsg 总是返回至少一个
func (l *Lsp) SendMsg(m *mmsg.MSG, target mmsg.Target) (res []interface{}) {
	msgs := m.ToMessage(target)
	if len(msgs) == 0 {
		switch target.TargetType() {
		case mmsg.TargetPrivate:
			res = append(res, &message.PrivateMessage{Id: -1})
		case mmsg.TargetGroup:
			res = append(res, &message.GroupMessage{Id: -1})
		}
		return
	}
	for idx, msg := range msgs {
		r := l.send(msg, target)
		res = append(res, r)
		// 原本的发送返回值已经无效，故直接无视
		// if reflect.ValueOf(r).Elem().FieldByName("Id").Int() == -1 {
		// 	break
		// }
		if idx > 1 {
			time.Sleep(time.Millisecond * 300)
		}
	}
	return res
}

func (l *Lsp) GM(res []interface{}) []*message.GroupMessage {
	var result []*message.GroupMessage
	for _, r := range res {
		result = append(result, r.(*message.GroupMessage))
	}
	return result
}

func (l *Lsp) PM(res []interface{}) []*message.PrivateMessage {
	var result []*message.PrivateMessage
	for _, r := range res {
		result = append(result, r.(*message.PrivateMessage))
	}
	return result
}

func (l *Lsp) sendPrivateMessage(uin int64, msg *message.SendingMessage) (res *message.PrivateMessage) {
	// if bot.Instance == nil || !bot.Instance.Online.Load() {
	// 	return &message.PrivateMessage{Id: -1, Elements: msg.Elements}
	// }
	if msg == nil {
		logger.WithFields(localutils.FriendLogFields(uin)).Debug("send with nil private message")
		return &message.PrivateMessage{Id: -1}
	}
	//logger.Debugf("发送私聊消息：%v\n", msgstringer.MsgToString(msg.Elements))
	msg.Elements = localutils.MessageFilter(msg.Elements, func(element message.IMessageElement) bool {
		return element != nil
	})
	if len(msg.Elements) == 0 {
		logger.WithFields(localutils.FriendLogFields(uin)).Debug("send with empty private message")
		return &message.PrivateMessage{Id: -1}
	}
	var newstring = msgstringer.MsgToString(msg.Elements)
	res = bot.Instance.SendPrivateMessage(uin, msg, newstring)
	if res == nil || res.Id == -1 {
		logger.WithField("content", msgstringer.MsgToString(msg.Elements)).
			WithFields(localutils.GroupLogFields(uin)).
			Errorf("发送私聊消息失败")
	}
	if res == nil {
		res = &message.PrivateMessage{Id: -1, Elements: msg.Elements}
	}
	return res
}

// sendGroupMessage 发送一条消息，返回值总是非nil，Id为-1表示发送失败
// miraigo偶尔发送消息会panic？！
func (l *Lsp) sendGroupMessage(groupCode int64, msg *message.SendingMessage, recovered ...bool) (res *message.GroupMessage) {
	//fmt.Printf("运行到发信息了%v\n", msgstringer.MsgToString(msg.Elements))
	defer func() {
		if e := recover(); e != nil {
			if len(recovered) == 0 {
				logger.WithField("content", msgstringer.MsgToString(msg.Elements)).
					WithField("stack", string(debug.Stack())).
					Errorf("sendGroupMessage panic recovered")
				res = l.sendGroupMessage(groupCode, msg, true)
			} else {
				logger.WithField("content", msgstringer.MsgToString(msg.Elements)).
					WithField("stack", string(debug.Stack())).
					Errorf("sendGroupMessage panic recovered but panic again %v", e)
				res = &message.GroupMessage{Id: -1, Elements: msg.Elements}
			}
		}
	}()
	if bot.Instance == nil || !bot.Instance.Online.Load() {
		return &message.GroupMessage{Id: -1, Elements: msg.Elements}
	}
	if l.LspStateManager.IsMuted(groupCode, bot.Instance.Uin) {
		logger.WithField("content", msgstringer.MsgToString(msg.Elements)).
			WithFields(localutils.GroupLogFields(groupCode)).
			Debug("BOT被禁言无法发送群消息")
		return &message.GroupMessage{Id: -1, Elements: msg.Elements}
	}
	if msg == nil {
		logger.Debug("消息为空，返回")
		logger.WithFields(localutils.GroupLogFields(groupCode)).Debug("send with nil group message")
		return &message.GroupMessage{Id: -1}
	}
	//logger.Debugf("发送群消息：%v\n", msgstringer.MsgToString(msg.Elements))
	msg.Elements = localutils.MessageFilter(msg.Elements, func(element message.IMessageElement) bool {
		return element != nil
	})
	if len(msg.Elements) == 0 {
		//logger.Debug("消息元素为空，返回")
		logger.WithFields(localutils.GroupLogFields(groupCode)).Debug("send with empty group message")
		return &message.GroupMessage{Id: -1}
	}
	var newstring = msgstringer.MsgToString(msg.Elements)
	ret := bot.Instance.SendGroupMessage(groupCode, msg, newstring)
	res = ret.RetMSG
	err := ret.Error
	if err != nil {
		msgStr := msgstringer.MsgToString(msg.Elements)
		if len(msgStr) > 150 {
			msgStr = msgStr[:150] + "..."
		}
		logger.WithField("content", msgStr).
			WithFields(localutils.GroupLogFields(groupCode)).
			Error(err)
		// if msg.Count(func(e message.IMessageElement) bool {
		// 	return e.Type() == message.At && e.(*message.AtElement).Target == 0
		// }) > 0 {
		// 	logger.WithField("content", msgStr).
		// 		WithFields(localutils.GroupLogFields(groupCode)).
		// 		Errorf("发送群消息失败，可能是@全员次数用尽")
		// } else {
		// 	logger.WithField("content", msgStr).
		// 		WithFields(localutils.GroupLogFields(groupCode)).
		// 		Errorf("发送群消息失败，可能是被禁言或者账号被风控")
		// }
	}
	if res == nil {
		logger.WithFields(localutils.GroupLogFields(groupCode)).Debug("failed to send message")
		res = &message.GroupMessage{Id: -1, Elements: msg.Elements}
	}
	return res
}

var Instance = &Lsp{
	concernNotify:          concern.ReadNotifyChan(),
	stop:                   make(chan interface{}),
	status:                 NewStatus(),
	msgLimit:               semaphore.NewWeighted(3),
	PermissionStateManager: permission.NewStateManager(),
	LspStateManager:        NewStateManager(),
	cron:                   cron.New(cron.WithLogger(cron.VerbosePrintfLogger(cronLog))),
}

func init() {
	bot.RegisterModule(Instance)

	template.RegisterExtFunc("currentMode", func() string {
		return string(Instance.LspStateManager.GetCurrentMode())
	})
}
