package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/Hello-Moeka/BilibiliDanmuRobot-Core/companion"
	"github.com/Hello-Moeka/BilibiliDanmuRobot-Core/config"
	"github.com/Hello-Moeka/BilibiliDanmuRobot-Core/entity"
	"github.com/Hello-Moeka/BilibiliDanmuRobot-Core/handler"
	"github.com/Hello-Moeka/BilibiliDanmuRobot-Core/http"
	"github.com/Hello-Moeka/BilibiliDanmuRobot-Core/logic"
	"github.com/Hello-Moeka/BilibiliDanmuRobot-Core/svc"
	"github.com/Hello-Moeka/BilibiliDanmuRobot-Core/utiles"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var configFile = flag.String("f", "etc/bilidanmaku-api.yaml", "the config file")
var checkConfig = flag.Bool("check-config", false, "validate configuration without connecting or sending")
var previewScene = flag.String("preview-scene", "", "generate chat/welcome/gift/clock/idle preview without sending to Bilibili")
var previewInput = flag.String("preview-input", "", "facts or viewer text for preview")
var previewName = flag.String("preview-name", "", "native mention nickname included in preview length budget")
var Version string
var cls handler.WsHandler

func main() {
	flag.Parse()
	handler.ConfigFile = *configFile
	var err error
	var c config.Config
	conf.MustLoad(*configFile, &c, conf.UseEnv())
	if err := c.ValidateCompanion(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if *checkConfig {
		fmt.Println("configuration valid")
		return
	}
	if *previewScene != "" {
		switch *previewScene {
		case "chat", "welcome", "gift", "clock", "idle":
		default:
			fmt.Fprintln(os.Stderr, "invalid preview scene")
			os.Exit(1)
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(c.Companion.TimeoutSeconds)*time.Second)
		defer cancel()
		e := companion.Event{Scene: *previewScene, Facts: *previewInput, Name: *previewName, Budget: companion.ReplyBudget(c.Companion.MaxUnits, *previewName), Time: companion.LocalTime(time.Now())}
		out, err := companion.Generate(ctx, c.ChatGPT, c.Companion.MaxTokens, e)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		out = companion.Fit(out, *previewName, c.Companion.MaxUnits)
		prefix := ""
		if *previewName != "" {
			prefix = "@" + *previewName + " "
		}
		_ = json.NewEncoder(os.Stdout).Encode(map[string]interface{}{"scene": *previewScene, "text": prefix + out, "units": companion.Units(prefix + out), "sent": false})
		return
	}
	logx.MustSetup(c.Log)
	logx.DisableStat()

	logx.Infof("当前版本: %s", Version)
	logx.Infof("监听直播间: %d", c.RoomId)
	dir := "./token"
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		// Directory does not exist, create it
		err = os.Mkdir(dir, 0755)
		if err != nil {
			panic(fmt.Sprintf("无法创建token文件夹 请手动创建:%s", err))
		}
	}
	//配置数据库文件夹
	dbdir, err := os.Stat(c.DBPath)
	if os.IsNotExist(err) || !dbdir.IsDir() {
		err = os.MkdirAll(c.DBPath, 0777)
		if err != nil {
			logx.Errorf("数据库文件夹创建失败：%s", c.DBPath)
			panic(fmt.Sprintf("无法创建数据库文件夹 请手动创建:%s", err))
		}
	}
	http.InitHttpClient()
	// 判断是否存在历史cookie
	if http.FileExists("token/bili_token.txt") && http.FileExists("token/bili_token.json") {
		err = http.SetHistoryCookie()
		if err != nil {
			if err = userlogin(); err != nil {
				logx.Errorf("用户登录失败：%v", err)
				return
			}
		}
		logx.Info("用户登录成功")
	} else {
		if err = userlogin(); err != nil {
			logx.Errorf("用户登录失败：%v", err)
			return
		}
		logx.Info("用户登录成功")
	}
	ctx := svc.NewServiceContext(c)
	sig := make(chan os.Signal)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	var interval = 10 * time.Second
	t := time.NewTimer(interval)
	defer t.Stop()
	var info *entity.RoomInitInfo
	var preStatus int

	input := make(chan string)
	if ctx.Config.CustomizeBullet {
		logx.Info("启动命令行输入线程...")
		go getTerminalInput(input)
	}
	logx.Info("正在检测直播间是否开播...")
	cls = handler.NewWsHandler()
	for {
		select {

		// 程序退出
		case <-sig:
			goto END

		case ta := <-input:
			if ctx.Config.CustomizeBullet {
				logic.PushToBulletSender(ta)
			}

		// 每10秒检查一次直播间是否开播
		case <-t.C:
			if info, err = http.RoomInit(ctx.Config.RoomId); err != nil {
				logx.Infof("RoomInit错误：%v", err)
				t.Reset(interval)
				continue
			}
			if info.Data.LiveStatus == entity.Live && preStatus == entity.NotStarted { // 由NotStarted到Live是开播
				logx.Infof("开播啦！%v", ctx.Config.RoomId)
				//preStatus = entity.Live

				if cls == nil {
					t.Reset(interval)
					continue
				}
				err := cls.StartWsClient() // 开启弹幕姬
				if err != nil {
					logx.Error(err)
					t.Reset(interval)
					continue
				}
				preStatus = entity.Live
			} else if info.Data.LiveStatus != entity.Live && preStatus == entity.Live { // 由Live到NotStarted是下播
				logx.Info("下播啦！")
				preStatus = entity.NotStarted
				cls.SayGoodbye()
				cls.StopWsClient()
			}
			t.Reset(interval)
		}
	}
END:
	if cls != nil {
		cls.StopWsClient()
		cls.StopChanel()
	}
	return
}
func userlogin() error {
	var err error
	http.InitHttpClient()
	var loginUrl *entity.LoginUrl
	if loginUrl, err = http.GetLoginUrl(); err != nil {
		logx.Error(err)
		return err
	}

	if err = utiles.GenerateQr(loginUrl.Data.Url); err != nil {
		logx.Error(err)
		return err
	}

	if _, err = http.GetLoginInfo(loginUrl.Data.OauthKey); err != nil {
		logx.Error(err)
		return err
	}

	return err
}
func getTerminalInput(input chan string) {
	scanner := bufio.NewScanner(os.Stdin)
	for {
		scanner.Scan()
		input <- scanner.Text()
	}
}
