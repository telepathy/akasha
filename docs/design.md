# 依赖管理系统 (Akasha) 设计文档

## 概述

Akasha 是一个基于 Go Gin 框架的依赖管理系统，用于管理 Java 项目的 Gradle 依赖版本。该系统参考 `basic.md` 中的分支式二方依赖管理模型，支持多分支版本管理、版本历史追踪和闪回查询。

## 技术栈

- **后端**: Go 1.25 + Gin
- **数据库**: MySQL 8.0
- **ORM**: GORM
- **前端**: 原生 HTML/CSS/JS

## 数据模型

### Branch (分支)

| 字段 | 类型 | 说明 |
|------|------|------|
| id | uint | 主键 |
| name | string | 分支名，如 `202603`、`main` |
| baseBranch | string | 创建时基于的分支 |
| status | string | 状态：`active` / `archived` / `deleted` |
| createdAt | time | 创建时间 |

### Dependency (依赖)

| 字段 | 类型 | 说明 |
|------|------|------|
| id | uint | 主键 |
| name | string | 简称，如 `spring-core` |
| groupId | string | Maven groupId |
| artifact | string | Maven artifactId |
| version | string | 版本号 |
| branch | string | 所属分支 |
| sourceIp | string | 来源 IP |
| remark | string | 备注 |
| createdAt | time | 创建时间（闪回时间戳） |
| deletedAt | time | 软删除时间 |

## 核心规则

### 1. 分支创建
新分支从已有分支复制全部依赖，新分支创建后自动复制源分支的所有依赖。

### 2. 版本更新
不修改已有记录，而是插入新记录。查询时按 `created_at DESC` 取最新版本。

### 3. 版本历史
同一依赖在同一分支下有多条记录，通过 `createdAt` 实现版本追溯。

### 4. 删除规则
只有在 `active` 状态的分支上可以删除依赖条目。

### 5. 锁定/解锁
锁定后的分支（archived）不可修改，但可继续获取依赖。可以解锁恢复为 active。

### 6. 分支合并
将源分支的依赖合并到目标分支，仅当版本号高于目标分支时覆盖。

## API 接口

### 分支管理

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/v1/branches` | 分支列表 |
| GET | `/api/v1/branches/:name` | 分支详情 |
| POST | `/api/v1/branches` | 创建分支 |
| DELETE | `/api/v1/branches/:name` | 删除分支 |
| POST | `/api/v1/branches/:name/merge` | 合并分支 |
| POST | `/api/v1/branches/:name/archive` | 锁定分支 |
| POST | `/api/v1/branches/:name/unlock` | 解锁分支 |

### 依赖管理

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/v1/dependencies?branch=xxx` | 依赖列表 |
| GET | `/api/v1/dependencies/:name?branch=xxx` | 依赖详情 |
| GET | `/api/v1/dependencies/:name/history?branch=xxx` | 版本历史 |
| GET | `/api/v1/dependencies/:name/history-between?branch=xxx&startAt=xxx&endAt=xxx` | 时间段变更查询 |
| GET | `/api/v1/dependencies/:name/at?branch=xxx&at=时间` | 闪回查询 |
| POST | `/api/v1/dependencies` | 新增/更新依赖 |
| DELETE | `/api/v1/dependencies/:name?branch=xxx` | 删除依赖 |

### 批量查询

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/v1/branches/:name/deps-at?at=时间` | 批量闪回查询 |
| GET | `/api/v1/branches/:name/history` | 分支所有依赖历史 |

### Gradle 输出

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/dependency?branch=xxx` | 输出 Gradle 格式依赖列表 |

### 系统管理

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/v1/health/db` | 检查数据库表结构是否存在 |
| POST | `/api/v1/init` | 初始化数据库（建表、创建 main 分支、导入数据） |
| POST | `/api/v1/login` | 登录（返回 JWT Cookie） |
| POST | `/api/v1/logout` | 登出（清除 JWT Cookie） |

## 页面路由

| 路径 | 说明 | 认证 |
|------|------|------|
| `/` | 首页：分支选择 + 依赖列表 | 公开 |
| `/login` | 登录页面 | 公开 |
| `/dependencies` | 依赖列表（同首页） | 公开 |
| `/dependency` | Gradle 格式依赖输出 | 公开 |
| `/compare` | 分支比较 | 公开 |
| `/branches` | 分支管理页面 | 需要认证 |
| `/merge` | 分支合并 | 需要认证 |

## 项目结构

```
akasha/
├── cmd/server/main.go          # 入口
├── internal/
│   ├── config/              # 配置加载
│   ├── domain/            # 数据模型
│   ├── repository/        # 数据访问层
│   ├── service/         # 业务逻辑
│   ├── handler/         # HTTP 处理器
│   └── router/          # 路由注册
├── templates/               # HTML 模板
├── static/
│   ├── css/style.css
│   └── js/app.js
├── docs/
│   └── design.md          # 本文档
├── config.yaml             # 配置文件
├── docker-compose.yml     # MySQL 开发环境
├── dependency.gradle    # 示例数据
├── Makefile
└── README.md
```

## 配置说明

### config.yaml

```yaml
database:
  host: "127.0.0.1"
  port: 3306
  username: "root"
  password: "root123"
  name: "akasha_test"

app:
  host: "0.0.0.0"
  port: 8080

admin:
  password: "akasha_admin"

auth:
  jwtSecret: "change-me-to-a-random-string-at-least-32-characters-long"

apiKey: "your-api-key-here"

externalHost: "http://localhost:8080"
```

### docker-compose.yml

```yaml
version: '3.8'
services:
  mysql:
    image: mysql:8.0
    ports:
      - "3306:3306"
    environment:
      MYSQL_ROOT_PASSWORD: root123
      MYSQL_DATABASE: akasha
```

## 快速开始

### 1. 启动 MySQL

```bash
docker-compose up -d
```

### 2. 启动服务

```bash
go run -mod=vendor cmd/server/main.go
# 或
make run
```

### 3. 访问

- Web UI: `http://localhost:8080`
- 分支管理: `http://localhost:8080/branches`
- Gradle 输出: `http://localhost:8080/dependency?branch=main`

### 4. 运行测试

```bash
go test -v ./internal/test/...
```

## 数据导入

服务启动时，如果数据库为空，会自动从 `dependency.gradle` 文件导入约 390 条依赖到 `main` 分支。

## 开发命令

```bash
make run      # 使用 vendor 运行服务
make build   # 编译
make test    # 运行测试
make vendor  # 更新 vendor 目录
make clean  # 清理
```