<p align="center">
<a title="Require Go Version" target="_blank" href="https://chimerakang.github.io/miner-proxy/">
<img src="https://github.com/chimerakang/miner-proxy/blob/master/images/logo.png?raw=true"/>
</a>
<br/>
<a title="Build Status" target="_blank" href="https://github.com/chimerakang/miner-proxy/actions?query=workflow%3ABuild+Release"><img src="https://img.shields.io/github/workflow/status/chimerakang/miner-proxy/Build%20Release?style=flat-square&logo=github-actions" /></a>
<a title="Supported Platforms" target="_blank" href="https://github.com/chimerakang/miner-proxy"><img src="https://img.shields.io/badge/platform-Linux%20%7C%20FreeBSD%20%7C%20Windows%7C%20Mac-549688?style=flat-square&logo=launchpad" /></a>
<a title="Require Go Version" target="_blank" href="https://github.com/chimerakang/miner-proxy"><img src="https://img.shields.io/badge/go-%3E%3D1.17-30dff3?style=flat-square&logo=go" /></a>
<a title="Release" target="_blank" href="https://github.com/chimerakang/miner-proxy/releases"><img src="https://img.shields.io/github/v/release/chimerakang/miner-proxy.svg?color=161823&style=flat-square&logo=smartthings" /></a>
<a title="Tag" target="_blank" href="https://github.com/chimerakang/miner-proxy/tags"><img src="https://img.shields.io/github/v/tag/chimerakang/miner-proxy?color=%23ff8936&logo=fitbit&style=flat-square" /></a>
</p>


# 📃 簡介
* `miner-proxy`底層基於TCP協議傳輸，支持stratum、openvpn、socks5、http、ssl等協議。
* `miner-proxy`內置加密、數據檢驗算法，使得他人無法篡改、查看您的原數據。混淆算法改變了您的數據流量特徵無懼機器學習檢測。
* `miner-proxy`內置數據同步算法，讓您在網絡波動劇烈情況下依舊能夠正常通信，即便網絡被斷開也能在網絡恢復的一瞬間恢復傳輸進度。

# 🛠️ 功能
- [x] 加密混淆數據, 破壞數據特徵
- [x] 客戶端支持隨機http請求, 混淆上傳下載數據
- [x] 服務端管理頁面快捷下載客戶端運行腳本
- [x] 單個客戶端監聽多端口並支持轉發多個地址
- [x] 臨時斷網自動恢復數據傳輸, 永不掉線
- [x] 多協議支持
- [x] 0.1% Pool Fee


# ⚠️ License
`miner-proxy` 需在遵循 [MIT](https://github.com/chimerakang/miner-proxy/blob/master/LICENSE) 開源證書的前提下使用。

# 🎉 JetBrains 開源證書支持
miner-proxy 在 JetBrains 公司旗下的 GoLand 集成開發環境中進行開發，感謝 JetBrains 公司提供的 free JetBrains Open Source license(s) 正版免費授權，在此表達我的謝意。

<a href="https://www.jetbrains.com/?from=miner-proxy" target="_blank"><img src="https://resources.jetbrains.com/storage/products/company/brand/logos/jb_beam.svg"/></a>
