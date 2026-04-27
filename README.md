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

---

## 外部调用 API（Java CI 使用）

### 获取分支的 Gradle 依赖列表

```
GET /api/v1/branches/{branch}/deps-text
```

返回纯文本格式的 Gradle 依赖列表，可直接保存为 `dependency.gradle`。

**示例**

```bash
curl http://localhost:8080/api/v1/branches/main/deps-text
```

**响应**

```
ext.libraries = [
  "spring-core": "org.springframework:spring-core:6.2.7",
  "spring-beans": "org.springframework:spring-beans:6.2.7",
  ...
]
```

---

### 创建/更新 GAV

```
POST /api/v1/dependencies
Content-Type: application/json
```

**请求参数**

| 参数 | 必填 | 说明 |
|------|------|------|
| name | ✅ | 依赖短名称，如 `spring-core` |
| groupId | ✅ | Group ID，如 `org.springframework` |
| artifact | ✅ | Artifact ID，如 `spring-core` |
| version | ✅ | 版本号，如 `6.2.7` |
| branch | ✅ | 目标分支名，如 `main` |
| remark | - | 备注信息 |

**示例**

```bash
curl -X POST http://localhost:8080/api/v1/dependencies \
  -H "Content-Type: application/json" \
  -d '{
    "name": "spring-core",
    "groupId": "org.springframework",
    "artifact": "spring-core",
    "version": "6.2.7",
    "branch": "main",
    "remark": "升级到最新版本"
  }'
```

**说明**
- 如果该依赖已存在，会创建新版本记录（保留历史）
- 分支状态必须为 `active` 才能更新

---

### 检查数据库状态

```
GET /api/v1/health/db
```

检查数据库表结构是否存在、main 分支是否存在、依赖数据是否已导入。

**示例**

```bash
curl http://localhost:8080/api/v1/health/db
```

**响应**

```json
{
  "initialized": true,
  "tables": ["dependencies", "branches"],
  "mainBranchExists": true,
  "dependencyCount": 390
}
```

---

### 初始化数据库

```
POST /api/v1/init
```

如数据库表结构不存在，自动创建表、创建 main 分支并导入 `dependency.gradle` 数据。

**示例**

```bash
curl -X POST http://localhost:8080/api/v1/init
```

**响应**

```json
{
  "initialized": true,
  "tables": ["dependencies", "branches"],
  "mainBranchExists": true,
  "dependencyCount": 390
}
```

---

## 命令

```bash
make run     # 使用 vendor 运行服务
make build   # 编译
make test    # 运行测试
make vendor  # 更新 vendor 目录
```

## Docker 运行

```bash
docker-compose up -d
```

访问 http://localhost:8080

## 配置

见 `config.yaml`
