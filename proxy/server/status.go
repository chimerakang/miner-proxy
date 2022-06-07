package server

import (
	"fmt"
	"miner-proxy/pkg"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/emirpasic/gods/sets/hashset"
	"github.com/liushuochen/gotable"
	"github.com/spf13/cast"
	"github.com/wxpusher/wxpusher-sdk-go/model"
)

var (
	pushers sync.Map
	m       sync.Mutex
)

type ClientSize struct {
	lastTime time.Time
	size     int64
}

func init() {
	go func() {
		for {
			pushers.Range(func(key, value interface{}) bool {
				p, ok := value.(*pusher)
				if !ok {
					return true
				}
				if err := p.UpdateUsers(); err != nil {
					pkg.Error("更新訂閱用戶失敗: %s", err)
					return true
				}
				return true
			})
			time.Sleep(time.Second*5*60 + 30) // 每5分30秒执行一次
		}
	}()
}

func Show(offlineTime time.Duration) {
	var offlineClient = hashset.New()
	table, _ := gotable.Create("Client ID", "Miner ID", "IP", "Package Size", "Connection Time", "Online", "Client-Ping", "Pool Connection", "Hash rate")
	for _, v := range ClientInfo() {
		for _, v1 := range v.Miners {
			if !v1.IsOnline && !v1.stopTime.IsZero() && time.Since(v1.stopTime).Seconds() >= offlineTime.Seconds() {
				offlineClient.Add(fmt.Sprintf("ip: %s; pool: %s; stop time: %s", v1.Ip, v1.Pool, v1.StopTime))
				clients.Delete(v1.Id)
			}

			_ = table.AddRow(map[string]string{
				"Client ID":       v.ClientId,
				"Miner ID":        v1.Id,
				"IP":              v1.Ip,
				"Package Size":    v1.Size,
				"Connection Time": v1.ConnTime,
				"Pool Connection": v1.Pool,
				"Online":          cast.ToString(v1.IsOnline),
				"Client-Ping":     v.Delay,
			})
		}
	}
	fmt.Println(table.String())
	if offlineClient.Size() != 0 {
		// 發送掉線通知
		SendOfflineIps(pkg.Interface2Strings(offlineClient.Values()))
	}
}

type Pusher interface {
	SendMessage(text string, uid ...string) error
	GetAllUser() ([]model.WxUser, error)
	GetToken() string
}

type pusher struct {
	Pusher         Pusher
	Users          []model.WxUser
	m              sync.Mutex
	lastUpdateUser time.Time
}

func (p *pusher) UpdateUsers() error {
	if !p.lastUpdateUser.IsZero() && time.Since(p.lastUpdateUser).Minutes() < 5 {
		return nil
	}
	users, _ := p.Pusher.GetAllUser()
	if len(users) == 0 {
		return nil
	}
	m.Lock()
	defer m.Unlock()
	p.Users = users
	p.lastUpdateUser = time.Now()
	return nil
}

func (p *pusher) SendMessage2All(msg string) error {
	var uids []string
	for _, v := range p.Users {
		uids = append(uids, v.UId)
	}
	return p.Pusher.SendMessage(msg, uids...)
}

func AddConnectErrorCallback(p Pusher) error {
	obj := &pusher{
		Pusher: p,
	}
	if err := obj.UpdateUsers(); err != nil {
		return err
	}
	pushers.Store(obj.Pusher.GetToken(), obj)
	return nil
}

func SendOfflineIps(offlineIps []string) {
	if len(offlineIps) <= 0 {
		return
	}
	var ips = strings.Join(offlineIps, "\n")
	if len(offlineIps) > 10 {
		ips = fmt.Sprintf("%s 等 %d 個 ip", strings.Join(offlineIps[:10], "\n"), len(offlineIps))
	}
	ips = fmt.Sprintf("您有掉線的機器:\n%s", ips)
	pushers.Range(func(key, value interface{}) bool {
		p := value.(*pusher)
		pkg.Info("發送掉線通知: %+v", p.Users)
		if err := p.SendMessage2All(ips); err != nil {
			pkg.Error("發送通知失敗: %s", err)
		}
		return true
	})
}

type ClientStatus struct {
	Id              string `json:"id"`
	ClientId        string `json:"client_id"`
	Ip              string `json:"ip"`
	Size            string `json:"size"`
	ConnectDuration string `json:"connect_duration"`
	RemoteAddr      string `json:"remote_addr"`
	IsOnline        bool   `json:"is_online"`
	// 根据传输数据大小判断预估算力
	HashRate        string `json:"hash_rate"`
	Delay           string `json:"delay"`
	connectDuration time.Duration
}

type ClientStatusArray []ClientStatus

func (c ClientStatusArray) Len() int {
	return len(c)
}

func (c ClientStatusArray) Less(i, j int) bool {
	return c[i].connectDuration.Nanoseconds() > c[j].connectDuration.Nanoseconds()
}

// Swap swaps the elements with indexes i and j.
func (c ClientStatusArray) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]

}

type ClientRemoteAddr struct {
	Delay         string `json:"delay"`
	ClientId      string `json:"client_id"`
	ConnSize      int    `json:"conn_size"`
	dataSize      int64
	DataSize      string  `json:"data_size"`
	RemoteAddr    string  `json:"remote_address"`
	Pool          string  `json:"pool"`
	SendDataCount int     `json:"send_data_count"`
	Miners        []Miner `json:"miners"`
	OnlineTime    string  `json:"online_time"`
}

type Miner struct {
	dataSize int64
	Id       string `json:"id"`
	Ip       string `json:"ip"`
	ConnTime string `json:"conn_time"`
	Pool     string `json:"pool"`
	Size     string `json:"size"`
	StopTime string `json:"stop_time"`
	stopTime time.Time
	IsOnline bool `json:"is_online"`
}

type ClientRemoteAddrs []*ClientRemoteAddr

func (c ClientRemoteAddrs) Len() int {
	return len(c)
}

func (c ClientRemoteAddrs) Less(i, j int) bool {
	var iTotal int
	var jTotal int
	iTotal += c[i].ConnSize + len(c[i].Miners) + int(c[i].dataSize)
	jTotal += c[j].ConnSize + len(c[j].Miners) + int(c[j].dataSize)

	return iTotal > jTotal
}

// Swap swaps the elements with indexes i and j.
func (c ClientRemoteAddrs) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]

}

func ClientInfo() []*ClientRemoteAddr {
	var clientMap = make(map[string][]Miner)
	var clientPools = make(map[string]*hashset.Set)
	var clientSizeMap = make(map[string]int64)
	var existIpMiner = make(map[string]struct{})
	clients.Range(func(key, value interface{}) bool {
		c := value.(*Client)
		if _, ok := clientPools[c.clientId]; !ok {
			clientPools[c.clientId] = hashset.New()
		}

		if _, ok := existIpMiner[c.ip]; ok && c.closed.Load() {
			pkg.Debug("delete old miner connection, use new miner connection")
			clients.Delete(key)
			return true
		}

		if time.Since(c.startTime).Seconds() >= 30 && c.dataSize.Load() <= 0 {
			pkg.Debug("delete unused connection")
			clients.Delete(key)
			return true
		}

		clientPools[c.clientId].Add(c.address)
		m := &Miner{
			Id:       c.id,
			Ip:       c.ip,
			Pool:     c.pool.Address(),
			ConnTime: time.Since(c.startTime).String(),
			Size:     humanize.Bytes(uint64(c.dataSize.Load())),
			IsOnline: !c.closed.Load(),
		}
		if !m.IsOnline && !c.stopTime.IsZero() {
			m.StopTime = time.Since(c.stopTime).String()
			m.stopTime = c.stopTime
		}
		if m.IsOnline {
			existIpMiner[c.ip] = struct{}{}
		}
		clientSizeMap[c.clientId] += c.dataSize.Load()
		clientMap[c.clientId] = append(clientMap[c.clientId], *m)
		return true
	})

	var result ClientRemoteAddrs
	conns.Range(func(key, value interface{}) bool {
		cd := value.(*ClientDispatch)
		c := &ClientRemoteAddr{
			ClientId:   cast.ToString(key),
			ConnSize:   cd.ConnCount(),
			Pool:       cd.pool,
			OnlineTime: time.Since(cd.startTime).String(),
			RemoteAddr: cd.remoteAddr,
		}
		if _, ok := clientMap[cast.ToString(key)]; ok {
			c.Miners = clientMap[cast.ToString(key)]
			c.dataSize = clientSizeMap[cast.ToString(key)]
			c.DataSize = humanize.Bytes(uint64(clientSizeMap[cast.ToString(key)]))
			c.Pool = strings.Join(pkg.Interface2Strings(clientPools[cast.ToString(key)].Values()), ",")
		}
		v, _ := connDelay.Load(c.ClientId)
		if v == nil {
			v = Delay{}
		}

		d := v.(Delay)
		c.Delay = "Checking..."
		if d.delay.Seconds() <= 120 {
			c.Delay = d.delay.String()
		}

		result = append(result, c)

		return true
	})

	sort.Sort(result)
	return result
}
