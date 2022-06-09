package main

import (
	_ "embed"
	"fmt"
	app2 "miner-proxy/app"
	"miner-proxy/pkg"
	"miner-proxy/pkg/config"
	"miner-proxy/pkg/middleware"
	"miner-proxy/proxy/client"
	"miner-proxy/proxy/server"
	"miner-proxy/proxy/wxPusher"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/denisbrodbeck/machineid"
	"github.com/gin-gonic/gin"
	"github.com/jmcvetta/randutil"
	"github.com/kardianos/service"
	"github.com/liushuochen/gotable"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/urfave/cli"
	"go.uber.org/zap/zapcore"
)

var (
	// build 時加入
	gitCommit string
	version   string
	//go:embed web/index.html
	indexHtml []byte
)
var (
	reqeustUrls = []string{
		"https://www.baidu.com/",
		"https://m.baidu.com/",
		"https://www.jianshu.com/",
		"https://www.jianshu.com/p/4fbdab9fb44c",
		"https://www.jianshu.com/p/5d25218fb22d",
		"https://www.tencent.com/",
		"https://tieba.baidu.com/",
	}
)
var RootCmd = &cobra.Command{
	Use:   "miner-pool",
	Short: "Open source, miner pool",
}

func init() {
	version = "0.2"
	RootCmd.PersistentFlags().StringP("config", "c", "", "Configuration file to use.")
}

type proxyService struct {
	args *cli.Context
	cfg  *config.Config
}

func (p *proxyService) checkWxPusher(wxPusherToken string, newWxPusherUser bool) error {
	if len(wxPusherToken) <= 10 {
		pkg.Fatal("您輸入的微信通知token無效, 請在 https://wxpusher.zjiecode.com/admin/main/app/appToken 中獲取")
	}
	w := wxPusher.NewPusher(wxPusherToken)
	if newWxPusherUser {
		qrUrl, err := w.ShowQrCode()
		if err != nil {
			pkg.Fatal("獲取二維碼url失敗: %s", err.Error())
		}
		fmt.Printf("請複製網址, 在瀏覽器打開, 並使用微信進行掃碼登陸: %s\n", qrUrl)
		pkg.Input("您是否掃描完成?(y/n):", func(s string) bool {
			if strings.ToLower(s) == "y" {
				return true
			}
			return false
		})
	}

	users, err := w.GetAllUser()
	if err != nil {
		pkg.Fatal("獲取所有的user失敗: %s", err.Error())
	}
	table, _ := gotable.Create("uid", "微信暱稱")
	for _, v := range users {
		_ = table.AddRow(map[string]string{
			"uid":  v.UId,
			"微信暱稱": v.NickName,
		})
	}
	fmt.Println("您已經註冊的微信通知用戶, 如果您還需要增加用戶, 請再次運行 ./miner-proxy -add_wx_user -wx tokne, 增加用戶, 已經運行的程序將會在5分鐘內更新訂閱的用戶:")
	fmt.Println(table.String())
	if !p.args.Bool("c") && (p.args.String("l") != "" && p.args.String("k") != "") {
		// 不是客戶端並且不是只想要增加新的用戶, 就直接將wxpusher obj 註冊回調
		if err := server.AddConnectErrorCallback(w); err != nil {
			pkg.Fatal("註冊失敗通知callback失敗: %s", err.Error())
		}
	}
	return nil
}

func (p *proxyService) startHttpServer() {
	gin.SetMode(gin.ReleaseMode)
	app := gin.New()
	app.Use(gin.Recovery(), middleware.Cors())

	skipAuthPaths := []string{
		"/download/",
	}

	if p.args.String("p") != "" {
		middlewareFunc := gin.BasicAuth(gin.Accounts{
			"admin": p.args.String("p"),
		})
		app.Use(func(ctx *gin.Context) {
			for _, v := range skipAuthPaths {
				if strings.HasPrefix(ctx.Request.URL.Path, v) {
					return
				}
			}
			middlewareFunc(ctx)
			return
		})
	}

	port := strings.Split(p.args.String("l"), ":")[1]

	app.Use(func(ctx *gin.Context) {
		ctx.Set("tag", version)
		ctx.Set("secretKey", p.args.String("k"))
		ctx.Set("server_port", port)
		ctx.Set("download_github_url", p.args.String("g"))
	})

	app2.NewRouter(app)

	app.NoRoute(func(c *gin.Context) {
		c.Data(http.StatusOK, "text/html", indexHtml)
	})

	pkg.Info("web server address: %s", p.args.String("a"))
	if err := app.Run(p.args.String("a")); err != nil {
		pkg.Panic(err.Error())
	}
}

func (p *proxyService) Start(_ service.Service) error {
	go p.run()
	return nil
}

func (p *proxyService) randomRequestHttp() {
	defer func() {
		sleepTime, _ := randutil.IntRange(10, 60)
		time.AfterFunc(time.Duration(sleepTime)*time.Second, p.randomRequestHttp)
	}()

	index, _ := randutil.IntRange(0, len(reqeustUrls))
	pkg.Debug("request: %s", reqeustUrls[index])
	resp, _ := (&http.Client{Timeout: time.Second * 10}).Get(reqeustUrls[index])
	if resp == nil {
		return
	}
	_ = resp.Body.Close()
}

func (p *proxyService) run() {
	defer func() {
		if err := recover(); err != nil {
			pkg.Error("crash: %v, restarting...", err)
			p.run()
		}
	}()

	// global.InitClientConfig(p.cfg)
	cfg, err := pkg.LoadConfig()
	if err != nil {
		log.Errorf("Error loading configuration: %v", err.Error())
	}
	// 启动服务器
	// return runServer(cfg, interruptChan)

	// if p.args.Bool("d") {
	// 	pkg.Warn("enable -debug")
	// }
	if *cfg.Debugger.Enable {
		pkg.Warn("enable debug")
	}

	// secret key

	// if len(p.args.String("k")) > 32 {
	// 	pkg.Error("password needs to be less then 32 digits")
	// 	os.Exit(1)
	// }
	// secretKey := p.args.String("k")
	// for len(secretKey)%16 != 0 {
	// 	secretKey += "0"
	// }
	// _ = p.args.Set("k", secretKey)
	if *cfg.Secret.Enable {
		if len(*cfg.Secret.Key) > 32 {
			pkg.Error("password needs to be less then 32 digits")
			os.Exit(1)
		}
		secretKey := *cfg.Secret.Key
		for len(secretKey)%16 != 0 {
			secretKey += "0"
		}
		*cfg.Secret.Key = secretKey
	}

	// run frontend or backend
	// if p.args.Bool("c") {
	// 	go p.randomRequestHttp()

	// 	if err := p.runClient(); err != nil {
	// 		pkg.Fatal("run client failed %s", err)
	// 	}

	// 	select {}
	// }
	// if !p.args.Bool("c") {
	// 	go func() {
	// 		for range time.Tick(time.Second * 60) {
	// 			server.Show(time.Duration(p.args.Int64("offline")) * time.Second)
	// 		}
	// 	}()
	// 	if p.args.String("a") != "" {
	// 		go p.startHttpServer()
	// 	}

	// 	if err := p.runServer(); err != nil {
	// 		pkg.Fatal("run server failed %s", err)
	// 	}
	// }

	p.cfg = cfg
	if *cfg.Client.Enable {
		go p.randomRequestHttp()

		if err := p.runClient(); err != nil {
			pkg.Fatal("run client failed %s", err)
		}
	} else {
		go func() {
			for range time.Tick(time.Second * 60) {
				server.Show(time.Duration(*cfg.Backend.Offline) * time.Second)
			}
		}()
		// if p.args.String("a") != "" {
		// 	go p.startHttpServer()
		// }
		if *cfg.Backend.Web {
			go p.startHttpServer()
		}
		if err := p.runServer(); err != nil {
			pkg.Fatal("run server failed %s", err)
		}
	}

}

func (p *proxyService) runClient() error {
	id, _ := machineid.ID()
	// pools := strings.Split(p.args.String("u"), ",")
	pools := strings.Split(*p.cfg.Client.Pools, ",")
	// for index, port := range strings.Split(p.args.String("l"), ",") {
	for index, port := range strings.Split(*p.cfg.Client.Listen, ",") {
		port = strings.ReplaceAll(port, " ", "")
		if port == "" {
			continue
		}
		if len(pools) < index {
			return errors.Errorf("-l parameters: %s, --pool parameters:%s; mush match", *p.cfg.Client.Listen, pools)
		}
		pools[index] = strings.ReplaceAll(pools[index], " ", "")
		// clientId := pkg.Crc32IEEEStr(fmt.Sprintf("%s-%s-%s-%s-%s", id,
		// 	p.args.String("k"), p.args.String("r"), port, pools[index]))
		clientId := pkg.Crc32IEEEStr(fmt.Sprintf("%s-%s-%s-%s-%s", id,
			*p.cfg.Secret.Key, *p.cfg.Client.Backend, port, pools[index]))

		if err := pkg.Try(func() bool {
			// if err := client.InitServerManage(p.args.Int("n"), p.args.String("k"), p.args.String("r"), clientId, pools[index]); err != nil {
			if err := client.InitServerManage(*p.cfg.Client.MaxConnect, *p.cfg.Secret.Key, *p.cfg.Client.Backend, clientId, pools[index]); err != nil {
				pkg.Error("connect to %s failed, check backend's firewall open backend listen port, or check service is enable! error: %s", *p.cfg.Client.Backend, err)
				time.Sleep(time.Second)
				return false
			}
			return true
		}, 1000); err != nil {
			pkg.Fatal("connect to backend failed!")
		}

		fmt.Printf("Listen Port '%s', pool host: '%s'\n", port, pools[index])
		go func(pool, clientId, port string) {
			// if err := client.RunClient(port, p.args.String("k"), p.args.String("r"), pool, clientId); err != nil {
			if err := client.RunClient(port, *p.cfg.Secret.Key, *p.cfg.Client.Backend, pool, clientId); err != nil {
				pkg.Panic("Start %s client app failed: %s", clientId, err)
			}
		}(pools[index], clientId, port)
	}
	return nil
}

func (p *proxyService) runServer() error {
	// return server.NewServer(p.args.String("l"), p.args.String("k"), p.args.String("r"))
	return server.NewServer(*p.cfg.Backend.Listen, *p.cfg.Secret.Key, *p.cfg.Backend.Pool)
}

func (p *proxyService) Stop(_ service.Service) error {
	return nil
}

func getArgs() []string {
	var result []string
	cmds := []string{
		"install", "remove", "stop", "restart", "start", "stat", "--delete",
	}
A:
	for _, v := range os.Args[1:] {
		for _, c := range cmds {
			if strings.Contains(v, c) {
				continue A
			}
		}
		result = append(result, v)
	}
	return result
}

func Install(c *cli.Context) error {
	s, err := NewService(c)
	if err != nil {
		return err
	}
	status, _ := s.Status()
	switch status {
	case service.StatusStopped, service.StatusRunning:
		if !c.Bool("delete") {
			pkg.Warn("已經存在一個服務!如果你需要重新安裝請在本次參數尾部加上 --delete")
			return nil
		}
		if status == service.StatusRunning {
			_ = s.Stop()
		}
		if err := s.Uninstall(); err != nil {
			return errors.Wrap(err, "卸載存在的服務失敗")
		}
		pkg.Info("成功卸載已經存在的服務")
	}
	if err := s.Install(); err != nil {
		return errors.Wrap(err, "安裝服務失敗")
	}
	return Start(c)
}

func Remove(c *cli.Context) error {
	s, err := NewService(c)
	if err != nil {
		return err
	}
	status, _ := s.Status()
	switch status {
	case service.StatusStopped, service.StatusRunning, service.StatusUnknown:
		if status == service.StatusRunning {
			_ = s.Stop()
		}
		if err := s.Uninstall(); err != nil {
			return errors.Wrap(err, "卸載服務失敗")
		}
		pkg.Info("成功卸載服務")
	}
	return nil
}

func Restart(c *cli.Context) error {
	s, err := NewService(c)
	if err != nil {
		return err
	}
	status, _ := s.Status()
	switch status {
	case service.StatusStopped, service.StatusRunning, service.StatusUnknown:
		if err := s.Restart(); err != nil {
			return errors.Wrap(err, "重新啟動服務失敗")
		}
		status, _ := s.Status()
		if status != service.StatusRunning {
			return errors.New("該服務沒有正常啟動, 請查看日誌!")
		}
		pkg.Info("重新啟動服務成功")
	}
	return nil
}

func Start(c *cli.Context) error {
	s, err := NewService(c)
	if err != nil {
		return err
	}
	status, _ := s.Status()
	switch status {
	case service.StatusRunning:
		pkg.Info("service is runing")
		return nil
	case service.StatusStopped, service.StatusUnknown:
		if err := s.Start(); err != nil {
			return errors.Wrap(err, "start service failed")
		}
		pkg.Info("start service success")
		return nil
	}
	return errors.New("service not installed!!")
}

func Stop(c *cli.Context) error {
	s, err := NewService(c)
	if err != nil {
		return err
	}
	status, _ := s.Status()
	switch status {
	case service.StatusRunning:
		if err := s.Stop(); err != nil {
			return errors.Wrap(err, "stop service failed")
		}
		return nil
	}
	pkg.Info("stop service success")
	return nil
}

func NewService(c *cli.Context) (service.Service, error) {
	svcConfig := &service.Config{
		Name:        "miner-proxy",
		DisplayName: "miner-proxy",
		Description: "miner encryption proxy service",
		Arguments:   getArgs(),
	}
	return service.New(&proxyService{args: c}, svcConfig)
}

var (
	Usages = []string{
		"install client app by background service(only linux): ./miner-proxy install -c -d -l :9999 -r server ip:port number -k secret key number -u pool host name:pool port",
		// "\t ins以服务的方式安装服务端: ./miner-proxy install  -d -l :9998 -r 默认矿池域名:默认矿池端口 -k 密钥",
		// "\t 更新以服務的方式安裝的客戶端/服務端: ./miner-proxy restart",
		// "\t 在客戶端/服務端添加微信掉線通知的訂閱用戶: ./miner-proxy add_wx_user -w appToken",
		// "\t 服務端增加掉線通知: ./miner-proxy install -d -l :9998 -r 默認礦池域名:默認礦池端口 -k 密鑰 --w appToken",
		// "\t linux查看以服務的方式安裝的日誌: journalctl -f -u miner-proxy",
		// "\t 客戶端監聽多個端口並且每個端口轉發不同的礦池: ./miner-proxy -l :監聽端口1,:監聽端口2,:監聽端口3 -r 服務端ip:服務端端口 -u 礦池鏈接1,礦池鏈接2,礦池鏈接3 -k 密鑰 -d",
	}
)

func main() {
	flags := []cli.Flag{
		cli.BoolFlag{
			Name:  "c",
			Usage: "標記當前運行的是客戶端",
		},
		cli.BoolFlag{
			Name:  "d",
			Usage: "是否開啟debug mode, 如果開啟了debug參數將會印出更多的日誌",
		},
		cli.StringFlag{
			Name:  "l",
			Usage: "當前程序監聽的地址",
			Value: ":9977",
		},
		cli.StringFlag{
			Name:  "r",
			Usage: "遠程礦池地址或者遠程本程序的監聽地址 (default \"localhost:80\")",
			// Value: "127.0.0.1:80",
			Value: "gladiatorwar.com:9998",
		},
		cli.StringFlag{
			Name:  "f",
			Usage: "將日誌寫入到指定的文件中",
		},
		cli.StringFlag{
			Name:  "k",
			Usage: "數據包加密密鑰, 長度小於等於32位",
			Value: "12345",
		},
		// cli.StringFlag{
		// 	Name:  "a",
		// 	Usage: "網頁查看狀態端口",
		// },
		// cli.StringFlag{
		// 	Name:  "u",
		// 	Usage: "客戶端如果設置了這個參數, 那麼服務端將會直接使用客戶端的參數連接, 如果需要多個礦池, 請使用 -l :端口1,端口2,端口3 -P 礦池1,礦池2,礦池3",
		// },
		// cli.StringFlag{
		// 	Name:  "w",
		// 	Usage: "掉線微信通知token, 該參數只有在服務端生效, ,請在 https://wxpusher.zjiecode.com/admin/main/app/appToken 註冊獲取appToken",
		// },
		// cli.IntFlag{
		// 	Name:  "o",
		// 	Usage: "掉線多少秒之後就發送微信通知,默認4分鐘",
		// 	Value: 360,
		// },
		// cli.StringFlag{
		// 	Name:  "p",
		// 	Usage: "訪問網頁端時的密碼, 如果沒有設置, 那麼網頁端將不需要密碼即可查看!固定的用戶名為:admin",
		// },
		// cli.StringFlag{
		// 	Name:  "g",
		// 	Usage: "服務端參數, 使用指定的網址加速github下載, 示例: -g https://gh.api.99988866.xyz/  將會使用 https://gh.api.99988866.xyz/https://github.com/PerrorOne/miner-proxy/releases/download/{tag}/miner-proxy下載",
		// },
		// cli.IntFlag{
		// 	Name:  "n",
		// 	Value: 10,
		// 	Usage: "客戶端參數, 指定客戶端啟動時對於每一個轉發端口通過多少tcp隧道連接服務端, 如果不清楚請保持默認, 不要設置小於2",
		// },
	}

	app := &cli.App{
		Name:      "miner-proxy",
		UsageText: strings.Join(Usages, "\n"),
		Commands: []cli.Command{
			{
				Name:   "install",
				Usage:  "./miner-proxy install: 將代理安裝到系統服務中, 開機自啟動, 必須使用root或者管理員權限運行",
				Action: Install,
				Flags: []cli.Flag{
					cli.BoolFlag{
						Name:  "delete",
						Usage: "如果已經存在一個服務, 那麼直接刪除後,再安裝",
					},
				},
			},
			{
				Name:   "remove",
				Usage:  "./miner-proxy remove: 將代理從系統服務中移除",
				Action: Remove,
			},
			{
				Name:   "restart",
				Usage:  "./miner-proxy restart: 重新啟動已經安裝到系統服務的代理",
				Action: Restart,
			},
			{
				Name:   "start",
				Usage:  "./miner-proxy start: 啟動已經安裝到系統服務的代理",
				Action: Start,
			},
			{
				Name:   "stop",
				Usage:  "./miner-proxy start: 停止已經安裝到系統服務的代理",
				Action: Stop,
			},
			{
				Name:  "add_wx_user",
				Usage: "./miner-proxy add_wx_user: 添加微信用戶到掉線通知中",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:     "w",
						Required: true,
						Usage:    "掉線微信通知token, 該參數只有在服務端生效, ,请在 https://wxpusher.zjiecode.com/admin/main/app/appToken 註冊獲取appToken",
					},
				},
				Action: func(c *cli.Context) error {
					return (&proxyService{args: c}).checkWxPusher(c.String("w"), true)
				},
			},
		},
		Flags: flags,
		Action: func(c *cli.Context) error {
			var logLevel = zapcore.InfoLevel
			if c.Bool("d") {
				logLevel = zapcore.DebugLevel
			}
			pkg.InitLog(logLevel, c.String("f"))
			if c.String("w") != "" {
				if err := (&proxyService{args: c}).checkWxPusher(c.String("w"), false); err != nil {
					pkg.Fatal(err.Error())
				}
			}

			s, err := NewService(c)
			if err != nil {
				return err
			}
			return s.Run()
		},
	}

	pkg.PrintHelp()
	fmt.Printf("Version:%s\nUpdate Logger:%s\n", version, gitCommit)
	if err := app.Run(os.Args); err != nil {
		pkg.Fatal("enable failed: %s", err)
	}
}
