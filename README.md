# DockPit

基于 [dockerCopilot](https://github.com/onlyLTY/dockercopilot) 二改的 Docker 容器管理工具。

## 新增功能（相比原版）

| 功能 | 原版 | DockPit |
|------|------|---------|
| 容器列表 | ✅ | ✅ |
| 启停重启 | ✅ | ✅ |
| 单容器更新 | ✅ | ✅ |
| **容器日志** | ❌ | ✅ 新增 |
| **Compose 项目发现** | ❌ | ✅ 新增 |
| **Compose 一键更新** | ❌ | ✅ 新增 |

## 部署

### Docker Compose（推荐）

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

## API 新增接口

### 容器日志
```
GET /api/container/:id/logs?tail=100
```

### Compose 项目列表
```
GET /api/compose/list
```

### Compose 一键更新
```
POST /api/compose/update
Body: {"name": "project_name", "working_dir": "/path/to/project"}
```

## 自动构建

推送代码到 `latest` 分支或打 `v*` tag 时，GitHub Actions 自动构建 `linux/amd64` + `linux/arm64` 双架构镜像并推送到 `ghcr.io`。

## License

AGPL-3.0（继承原项目）
