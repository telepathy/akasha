# Akasha 依赖管理系统

基于 Go Gin + MySQL 的 Gradle 依赖版本管理系统。

## 快速开始

### 1. 启动 MySQL

```bash
docker-compose up -d
```

### 2. 启动服务

```bash
go run cmd/server/main.go
# 或
make run
```

### 3. 访问

- Web UI: http://localhost:8080
- 分支管理: http://localhost:8080/branches
- Gradle 输出: http://localhost:8080/dependency?branch=main

## 技术栈

- Go 1.25 + Gin
- MySQL 8.0 + GORM
- 原生 HTML/CSS/JS

## 主要功能

- **分支管理**: 创建、锁定/解锁、删除
- **依赖管理**: GAV 版本管理、版本历史
- **时间查询**: 变更历史查询、闪回查询
- **Gradle 输出**: dependency.gradle 格式

## 命令

```bash
make run     # 使用 vendor 运行服务
make build # 编译
make test   # 运行测试
make vendor # 更新 vendor 目录
```

## Docker 运行

### 开发模式（使用外部MySQL）

```bash
# 启动 MySQL
docker-compose up -d mysql

# 启动应用
docker run -p 8080:8080 \
  -e DATABASE_HOST=host.docker.internal \
  -e DATABASE_PORT=3306 \
  -e DATABASE_USERNAME=root \
  -e DATABASE_PASSWORD=root123 \
  -e DATABASE_NAME=akasha \
  akasha:latest
```

### 生产模式（docker-compose）

```bash
docker-compose -f docker-compose.prod.yml up -d --build
```

访问 http://localhost:8080

## 分支状态

| 状态 | 说明 | 可添加依赖 | 可锁定 | 可删除 |
|------|------|----------|--------|--------|
| active | 正常 | ✅ | ✅ | ✅ |
| archived | 已锁定 | ❌ | - | ❌ |
| deleted | 已删除 | ❌ | ❌ | ❌ |

## 配置

见 `config.yaml`