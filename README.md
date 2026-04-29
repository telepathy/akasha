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

#### 方式一：直接下载（推荐）

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

#### 方式二：Gradle apply from（动态加载）

```
GET /dependency?branch={branch}
```

专为 Gradle `apply from:` 设计，支持通过 query 参数动态指定分支。

**build.gradle 示例**

```groovy
apply from: resources.text.fromInsecureUri("http://localhost:8080/dependency?branch=" + depBranch)
```

建议将 `depBranch` 加入 `gradle.properties`：

```properties
depBranch=main
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
  -H "X-API-Key: your-api-key-here" \
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
- 需在 Header 中携带 `X-API-Key`（当 `apiKey` 配置启用时）

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
curl -X POST http://localhost:8080/api/v1/init \
  -H "X-API-Key: your-api-key-here"
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

### 配置文件 `config.yaml`

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

apiKey: "your-api-key-here"

externalHost: "http://localhost:8080"
```

### 配置项说明

| 配置项 | 说明 | 默认值 |
|--------|------|--------|
| `database.host` | MySQL 主机地址 | 127.0.0.1 |
| `database.port` | MySQL 端口 | 3306 |
| `database.username` | MySQL 用户名 | root |
| `database.password` | MySQL 密码 | root123 |
| `database.name` | 数据库名 | akasha_test |
| `app.host` | 服务监听地址 | 0.0.0.0 |
| `app.port` | 服务端口 | 8080 |
| `admin.password` | 管理后台密码（为空则不启用页面认证） | - |
| `apiKey` | CI 调用 API Key（为空则不启用 API 认证） | - |
| `externalHost` | 外部访问地址（首页依赖URL显示用） | - |

### 配置传递方式

**方式一：配置文件（默认）**

将 `config.yaml` 放置于可执行文件同级目录，程序启动时自动读取。

**方式二：环境变量（优先级高于配置文件）**

| 环境变量 | 对应配置项 |
|----------|-----------|
| `DATABASE_HOST` | database.host |
| `DATABASE_PORT` | database.port |
| `DATABASE_USERNAME` | database.username |
| `DATABASE_PASSWORD` | database.password |
| `DATABASE_NAME` | database.name |
| `APP_HOST` | app.host |
| `APP_PORT` | app.port |
| `ADMIN_PASSWORD` | admin.password |
| `API_KEY` | apiKey |
| `EXTERNAL_HOST` | externalHost |

**Docker 运行示例**：

```bash
docker run -p 8080:8080 \
  -e DATABASE_HOST=mysql \
  -e DATABASE_PASSWORD=secret \
  -e API_KEY=ci-secret-key-123 \
  akasha:latest
```
