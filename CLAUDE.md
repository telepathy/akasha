# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**akasha** 是基于 Go Gin + MySQL 的依赖管理系统，用于管理 Java 项目的 Gradle 依赖版本。每条依赖记录都带时间戳，支持版本历史查询和"闪回"能力。

## Tech Stack

- Go 1.25 + Gin web framework
- GORM + MySQL database
- Masterminds/semver for version comparison
- 原生 HTML/CSS/JS frontend (无前端框架)

## Common Commands

```bash
# 启动 MySQL (如未运行)
docker-compose up -d

# 使用 vendor 构建/运行
make run     # 或 go run -mod=vendor cmd/server/main.go
make build  # 或 go build -mod=vendor -o bin/akasha cmd/server/main.go
make test   # 或 go test -mod=vendor ./...

# 更新依赖
make vendor  # go mod vendor
make tidy   # go mod tidy
```

## Architecture

```
akasha/
├── cmd/server/main.go           # 入口 - MySQL连接_AUTO MIGRATE数据导入
├── internal/
│   ├── config/config.go         # YAML配置加载
│   ├── domain/domain.go        # Dependency和Branch实体
│   ├── repository/             # 数据访问层
│   │   ├── branch_repo.go     # 分支CRUD_根据status过滤查询
│   │   └── dependency_repo.go  # 依赖CRUD_软删除_版本历史
│   ├── service/                 # 业务逻辑层
│   │   ├── branch_service.go   # 分支操作_CanModify权限检查
│   │   └── dependency_service.go  # 依赖操作_CreateOrUpdate逻辑
│   ├── handler/                 # HTTP处理层
│   │   ├── branch_handler.go
│   │   ├── dependency_handler.go
│   │   └── helper.go            # parseTime工具
│   └── router/router.go        # Gin路由注册_gradle输出
├── templates/                  # HTML模板
├── static/                     # CSS/JS静态资源
├── pkg/gradle/                # Gradle格式输出
└── config.yaml                 # 配置文件
```

## Key Rules

- **创建分支**: 从已有分支复制全部依赖到新分支名
- **更新依赖**: 不修改旧记录，而是 `INSERT` 新记录，新 `CreatedAt` 即闪回时间戳
- **删除依赖**: 只能在 active 分支上删除
- **锁定分支**: archived 状态不可修改，可解锁恢复
- **合并分支**: 遍历源分支的依赖，若版本号高于目标分支同名依赖，则覆盖

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/dependency?branch=xxx&pwd=xxx` | 输出 Gradle 格式依赖列表 |
| GET | `/api/v1/branches` | 分支列表 |
| POST | `/api/v1/branches` | 创建分支 `{name, baseBranch}` |
| DELETE | `/api/v1/branches/:name` | 软删除分支 |
| POST | `/api/v1/branches/:name/archive` | 锁定分支 |
| POST | `/api/v1/branches/:name/unlock` | 解锁分支 |
| GET | `/api/v1/dependencies?branch=xxx` | 查询依赖列表 |
| POST | `/api/v1/dependencies` | 新增/更新依赖 |
| GET | `/api/v1/dependencies/:name/history?branch=xxx` | 版本历史 |
| GET | `/api/v1/dependencies/:name/history-between?branch=xxx&startAt=xxx&endAt=xxx` | 时间段变更 |
| GET | `/api/v1/dependencies/:name/at?branch=xxx&at=时间` | 闪回查询 |
| GET | `/api/v1/branches/:name/deps-at?at=时间` | 批量闪回 |
| GET | `/api/v1/branches/:name/history` | 分支历史 |

## UI Pages

| Path | Description |
|------|-------------|
| GET `/` | 首页 - 分支列表 + 快速选择 |
| GET `/branches` | 分支管理 - 创建/锁定/解锁/删除 |

## Data Model

### Branch
- `Name`: 分支名 (如 202603)
- `BaseBranch`: 基分支
- `Status`: active / archived / deleted

### Dependency
- `Name`: 短名称 (如 spring-core)
- `GroupID` / `Artifact` / `Version`: Maven 坐标
- `Branch`: 所属分支
- `CreatedAt`: 记录创建时间 (用于闪回)
- `DeletedAt`: 软删除标记

## Branch Status

| Status | 可添加依赖 | 可锁定 | 可解锁 | 可删除 |
|--------|----------|--------|--------|--------|
| active | ✅ | ✅ | - | ✅ |
| archived | ❌ | - | ✅ | ❌ |
| deleted | ❌ | ❌ | ❌ | ❌ |

## Database

- MySQL 8.0 on localhost:3306
- Database: akasha_test
- User: root / password: root123
- 启动时自动创建 main 分支并导入 dependency.gradle 数据

## Docker

- Dockerfile: 多阶段构建
- docker-compose.prod.yml: 生产环境配置
- 通过环境变量覆盖数据库配置