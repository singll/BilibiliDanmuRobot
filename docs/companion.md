# 人格陪伴模式（CLI 部署）

在原有配置合并 `companion.example.yaml`，替换 Bellkeeper 地址和专用 token。模型组保持 `pool-bili-deepseek-paid`。五个场景共享 `ChatGPT.Prompt`：聊天、欢迎、礼物感谢、报时和冷场。完整人格与语气示例见 [persona-yuqinghuan.md](persona-yuqinghuan.md)。

- `Companion.Enabled` 为总开关，默认关闭；`Welcome/Gift/Clock/Idle` 可单独关闭。聊天仍由 `TalkRobotCmd` 触发。欢迎、关注、分享、礼物和定时任务还需开启原有开关；欢迎白名单、黑名单、财富等级筛选保留。关注/分享使用欢迎人格开关。
- `MaxUnits: 80` 同时限制正文和原生 `@昵称 `，ASCII 计 1、BMP 字符计 2、非 BMP 字符计 4。组合 emoji 不截断。发送前清除换行、Markdown 标记及常见 URL，超长时截短为一条；昵称耗尽预算时不发送。
- 礼物按 UID 合并，安静 8 秒后感谢，连续送礼最长等待 20 秒。合并真实礼物与数量；盲盒保留来源名称。此模式不额外发送旧版盲盒盈亏感谢串；盲盒数据库统计仍保留。舰长和红包走礼物人格。
- 报时沿用 `CronDanmuList`，提供当前北京时间；生成失败使用原任务的固定文案。聊天、欢迎和礼物失败也有短兜底；冷场失败直接跳过。
- 默认安静 300～360 秒后尝试冷场，至少间隔 900 秒，每小时最多尝试 2 次，同一次安静阶段只尝试一次。机器人刚发言 120 秒内不插话。观众弹幕、礼物、关注/分享会重新计时；发言生成中出现互动也会丢弃。
- WebSocket 认证成功且直播中才接收人格任务。断线、下播和超过 75 秒无数据时停发；重连不补发旧任务。欢迎/冷场 30 秒、聊天/报时 45 秒、礼物 60 秒过期。聊天等互动优先于报时/冷场，队列满直接丢弃。
- LLM 请求串行，默认至少相隔 12 秒、12 秒超时、最多 256 输出 token、关闭深度思考。只调用配置的地址与模型组，不自行切换渠道或模型；官方直连限制由 Bellkeeper 专用 token、组成员和渠道共同保证。
- 修改人格或 Companion/模型参数后重启 CLI 容器。GUI 尚无 Companion 配置表单；人格/模型热重载会明确提示重启，避免部分参数未生效。

先检查配置（不会发送、不会调用模型）：

```sh
bili-robot -f etc/bilidanmaku-api.yaml -check-config
```

真实模型预览（消耗一次模型请求，但不登录 B 站、不发送弹幕）：

```sh
bili-robot -f etc/bilidanmaku-api.yaml -preview-scene chat -preview-name 小鱼 -preview-input '幻梦，你在吗？'
bili-robot -f etc/bilidanmaku-api.yaml -preview-scene gift -preview-input '观众小鱼赠送了3朵小花'
```

`preview-scene` 支持 `chat/welcome/gift/clock/idle`。输出 JSON 包含最终文本、计数单位和 `sent:false`。预览用途不会经过运行时频率限制，请控制调用间隔。

维护 fork 同时涉及主程序和 Core；主程序 `go.mod` 固定 Core fork 的提交。构建 CLI：`CGO_ENABLED=0 go build -trimpath -o bili-robot ./cli`。正常运行仍需已有的 B 站登录文件和数据库目录。
