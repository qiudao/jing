# Tic-Tac-Toe RL Design

## Overview

Go 实现的井字棋游戏，使用 Q-Learning 强化学习训练 AI。Web UI，支持人机/人人/机器对弈，多难度等级。

## 项目结构

```
tictactoe/
├── main.go                 # 入口，启动 HTTP 服务
├── Makefile                # help, build, run, train
├── game/
│   ├── board.go            # 棋盘状态、合法动作、胜负判定
│   └── board_test.go
├── ai/
│   ├── qtable.go           # Q-Table 数据结构、查询/更新
│   ├── trainer.go          # 自我对弈训练循环
│   ├── agent.go            # AI 玩家（epsilon-greedy 选动作）
│   └── model.go            # 模型存储/加载（JSON）
├── server/
│   ├── handler.go          # HTTP/WebSocket 路由
│   └── hub.go              # WebSocket 连接管理
├── web/
│   ├── index.html          # 单页应用
│   ├── style.css
│   └── app.js
└── models/                 # 训练好的 Q-Table 文件
    ├── beginner.json
    ├── intermediate.json
    └── expert.json
```

## 核心逻辑

### 棋盘

- `[9]int` 表示棋盘（0=空, 1=X, 2=O）
- 纯逻辑层，无 IO

### Q-Learning

- **状态编码：** `[9]int` 转字符串 key，如 `"102010000"`
- **Q-Table：** `map[string]map[int]float64` — state -> action -> q_value
- **参数：** α=0.1, γ=0.9
- **奖励：** 赢+1, 输-1, 平+0.5, 每步0
- **训练：** 两个 Agent 自我对弈，各自视角更新同一张 Q-Table
- **ε 衰减：** 训练时从 1.0 线性衰减到 0.01

### 难度等级

| 难度 | 训练局数 | 对弈时 ε |
|------|---------|---------|
| 初级 | 10,000  | 0.5     |
| 中级 | 50,000  | 0.2     |
| 高级 | 200,000 | 0.0     |

## 通信协议

WebSocket JSON 消息，格式 `{"type": "xxx", "data": {...}}`

| 方向 | type | 用途 |
|------|------|------|
| → 后端 | start_game | 开始对局：mode(hvh/hva/ava), difficulty, speed |
| → 后端 | move | 人类落子：position(0-8) |
| → 后端 | train | 触发训练：difficulty |
| ← 前端 | game_state | 棋盘更新：board, turn, status |
| ← 前端 | train_progress | 训练进度：epoch, total, win_rate |
| ← 前端 | train_done | 训练完成 |
| ← 前端 | stats | 统计数据 |

## 前端布局

单页应用，三栏 + 底部训练面板：

- 左栏：模式选择、难度选择、观战速度滑块、开始/重来按钮
- 中栏：3x3 棋盘（CSS Grid）
- 右栏：当前状态、对战统计
- 底部：训练面板（难度选择、开始训练、进度条、Canvas 收敛曲线）

## 功能列表

- 人 vs 人（同浏览器轮流）
- 人 vs AI（选难度）
- AI vs AI 观战（可调速，100ms~2000ms）
- 后台训练 + 进度推送
- 慢速观战训练过程
- Q-Table 持久化（JSON 文件，多难度存档）
- 对战统计（胜/负/平计数，内存中）
- 收敛曲线（Canvas 折线图）

## 技术选型

- 后端：Go 标准库 + gorilla/websocket
- 前端：纯 HTML/CSS/JS，无框架
- 静态文件：embed.FS 内嵌，编译为单一二进制
