# DockPit

基于 [dockerCopilot](https://github.com/onlyLTY/dockercopilot) 二改的 Docker 容器管理工具。

## 功能对比

| 功能 | 原版 | DockPit |
|------|------|---------|
| 容器列表 | ✅ | ✅ |
| 启停重启 | ✅ | ✅ |
| 单容器更新 | ✅ | ✅ |
| **容器日志** | ❌ | ✅ |
| **容器资源监控** | ❌ | ✅ CPU/内存/网络 |
| **容器执行** | ❌ | ✅ exec 进容器 |
| **批量操作** | ❌ | ✅ 多选停止/重启/删除 |
| **健康检查状态** | ❌ | ✅ |
| **Compose 项目发现** | ❌ | ✅ 自动识别 |
| **Compose 一键更新** | ❌ | ✅ 失败自动回滚 |
| **Compose 日志** | ❌ | ✅ |
| **磁盘清理** | ❌ | ✅ 一键清理未使用资源 |
| **网络管理** | ❌ | ✅ 列出/删除 |
| **Volume 管理** | ❌ | ✅ 列出/删除 |
| **告警通知** | ❌ | ✅ Discord/Slack/飞书/企微 |

## 部署

```yaml
services:
  dockpit:
    container_name: dockpit
    restart: always
    privileged: true
    network_mode: bridge
    ports:
      - 12812:12812
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
      - ./data:/data
    environment:
      - TZ=Asia/Shanghai
      - DOCKER_HOST=unix:///var/run/docker.sock
      - secretKey=your_password_min_8_chars
    image: ghcr.io/atpx4869/dockpit:latest
```

## API 接口

### 容器操作
| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/containers` | 容器列表（含健康检查、CPU/内存） |
| POST | `/api/container/:id/start` | 启动 |
| POST | `/api/container/:id/stop` | 停止 |
| POST | `/api/container/:id/restart` | 重启 |
| POST | `/api/container/:id/update` | 更新 |
| POST | `/api/container/:id/rename` | 重命名 |
| GET | `/api/container/:id/logs?tail=100` | **日志** |
| GET | `/api/container/:id/stats` | **资源监控** |
| POST | `/api/container/:id/exec` | **执行命令** `{"cmd":["ls","-la"]}` |
| GET | `/api/container/:id/health` | **健康检查** |
| POST | `/api/containers/batch` | **批量操作** `{"ids":[...],"action":"stop"}` |

### Compose
| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/compose/list` | 项目列表（含状态：running/partial/stopped） |
| POST | `/api/compose/update` | 一键更新（失败自动回滚） |
| POST | `/api/compose/logs` | 项目日志 |

### 系统管理
| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/system/disk` | 磁盘占用 |
| POST | `/api/system/disk/cleanup` | 一键清理 |
| GET | `/api/system/networks` | 网络列表 |
| DELETE | `/api/system/network/:id` | 删除网络 |
| GET | `/api/system/volumes` | 卷列表 |
| DELETE | `/api/system/volume/:id` | 删除卷 |

## 告警配置

在 `etc/dockerCopilot.yaml` 中添加：

```yaml
AlertWebhook: "https://hooks.slack.com/services/xxx"  # 或 Discord/飞书/企微 webhook
```

支持自动识别：Discord、Slack、飞书、企业微信。

## License

AGPL-3.0
