# BilibiliDanmuRobot 维护版

这是 [xbclub/BilibiliDanmuRobot](https://github.com/xbclub/BilibiliDanmuRobot) 的独立维护分支，基于 Go 与 Wails，为哔哩哔哩直播间提供桌面弹幕机器人。上游停止维护后，本仓库继续维护已验证的功能修复，并通过 GitHub Releases 发布 Windows GUI 构建。

## 主要功能

- 直播弹幕接收、欢迎词、关注与分享答谢。
- 礼物答谢、盲盒统计、定时弹幕、关键词回复和抽奖等功能。
- [盲盒查询](docs/blind-box-statistics.md)：支持中文月份、今日／本周／本月等时段，以及 `盲盒统计` 查询全部历史。
- GUI 配置、扫码登录、日志查看和本地数据库。
- 支持 `SEND_GIFT` 与 `SEND_GIFT_V2` 礼物事件；V2 事件由 Core 解码后进入现有感谢与统计流程。
- 已移除失效的在线更新入口和 `upgrader.exe` 下载/启动流程。请从 Releases 手动获取新版本。

## 项目组成

- 本仓库：Wails 桌面应用、前端界面与 CLI 启动入口。
- [BilibiliDanmuRobot-Core](https://github.com/Hello-Moeka/BilibiliDanmuRobot-Core)：WebSocket 协议、事件解析和机器人业务逻辑组件。

桌面应用依赖 Core 模块 `github.com/Hello-Moeka/BilibiliDanmuRobot-Core`。本仓库的 Release 是面向 Windows x64 的 GUI 程序；Core 是 Go 模块，不单独生成机器人 exe。

## 使用

1. 从 [Releases](https://github.com/Hello-Moeka/BilibiliDanmuRobot/releases) 下载 Windows GUI 程序。
2. 在程序目录准备 `etc/bilidanmaku-api.yaml`，按需设置房间号和功能。
3. 启动 GUI，完成 B 站扫码登录，并在配置面板中启用所需功能。

程序会在当前工作目录读写 `etc/`、`token/`、`db/` 和 `logs/`。这些目录可能包含登录凭证、账号数据和日志；不要将真实账号数据上传到代码仓库或公开分享。

## 从源码构建

### 环境

- Go 1.24 或更新版本
- Node.js 与 pnpm（Wails 前端使用 pnpm）
- [Wails v2.10.1](https://wails.io/docs/gettingstarted/installation)
- 可访问 Go module proxy 与 npm registry 的网络环境

### Windows GUI

```powershell
go mod download
Set-Location frontend
pnpm install --frozen-lockfile
Set-Location ..
wails build -platform windows/amd64
```

Wails 默认将 GUI 程序输出到 `build/bin/`。

### CLI

```powershell
go build -o dist/BilibiliDanmuRobot.exe ./cli/bilidanmaku.go
```

## 维护说明

- 配置文件由机器人读取；请不要把 `token/`、`logs/`、`db/` 或个人配置提交到 Git。
- 版本更新由维护者通过源码提交和 GitHub Releases 发布，不再连接上游更新服务器。
- 礼物协议或 B 站接口变更时，优先核实实际 WebSocket 事件、鉴权结果与解析数据，再调整对应处理逻辑。

## 来源与致谢

本项目源自 [xbclub/BilibiliDanmuRobot](https://github.com/xbclub/BilibiliDanmuRobot)，核心弹幕协议组件参考并沿用 [Akegarasu/blivedm-go](https://github.com/Akegarasu/blivedm-go)。保留上游提交历史与相关来源说明。
