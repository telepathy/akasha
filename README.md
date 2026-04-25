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
- 分支比较: http://localhost:8080/compare
- 分支合并: http://localhost:8080/merge
- Gradle 输出: http://localhost:8080/dependency?branch=main

## 技术栈

- Go 1.25 + Gin
- MySQL 8.0 + GORM
- 原生 HTML/CSS/JS

## 主要功能

- **分支管理**: 创建、锁定/解锁、删除
- **依赖管理**: GAV 版本管理、版本历史
- **时间查询**: 变更历史查询、闪回查询
- **分支比较**: 对比两个分支的依赖差异
- **分支合并**: 支持多种策略的依赖合并
- **Gradle 输出**: dependency.gradle 格式

## 分支合并策略

### 策略类型

| 策略 | 说明 | 适用场景 |
|------|------|----------|
| `keep_higher` | 保留较高版本（默认） | 常规合并，确保版本只升不降 |
| `force_source` | 强制使用源分支版本 | 源分支是权威来源 |
| `force_target` | 保留目标分支版本 | 仅添加缺失依赖 |

### 合并规则

| 源分支 | 目标分支 | 处理方式 |
|--------|----------|----------|
| 有依赖X v2.0 | 有依赖X v1.0 | 按策略处理（默认升级） |
| 有依赖X v1.0 | 有依赖X v2.0 | 按策略处理（默认跳过，视为冲突） |
| 有依赖X | 无依赖X | 可选添加（addMissing=true） |
| 无依赖X | 有依赖X | **视为冲突**（可能被误删） |

### API

```bash
# 预览合并
POST /api/v1/branches/{source}/merge
{
    "targetBranch": "main",
    "strategy": "keep_higher",
    "addMissing": true,
    "dryRun": true
}

# 执行合并
POST /api/v1/branches/{source}/merge
{
    "targetBranch": "main",
    "strategy": "keep_higher",
    "addMissing": true,
    "dryRun": false
}
```

### 响应格式

```json
{
    "preview": true,
    "result": {
        "added": 5,
        "updated": 3,
        "skipped": 12,
        "conflicts": [
            {
                "name": "spring-core",
                "sourceVersion": "6.1.0",
                "targetVersion": "6.2.0",
                "reason": "目标版本更高"
            }
        ],
        "details": [
            "添加: new-lib 1.0.0",
            "更新: spring-core 5.3.0 -> 6.1.0"
        ]
    }
}
```

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
