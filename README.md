# Tic-Tac-Toe RL

Go 实现的井字棋游戏，使用 Q-Learning 强化学习训练 AI。

## 功能

- 人 vs 人、人 vs AI、AI vs AI 观战
- 三个难度等级（初级/中级/高级）
- 训练面板：进度条 + 收敛曲线
- 慢速观战训练过程
- 模型持久化，重启自动加载

## 运行

```bash
make build
make run
# 打开 http://localhost:8080
```

## 使用流程

1. 在训练面板选择难度，点击 Start Training
2. 训练完成后，选择模式和难度，点击 Start 开始对弈

## 技术栈

- Go + gorilla/websocket
- Q-Learning (epsilon-greedy, temporal difference)
- 纯 HTML/CSS/JS 前端，embed.FS 打包为单一二进制
