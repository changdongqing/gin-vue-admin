# Gin-Vue-Admin 启动指南（start.md）

本文档介绍如何在本机把本项目的**服务端（server）**和**前端（web）**运行起来。

## 一、环境要求

| 依赖 | 版本要求 | 说明 |
| ---- | -------- | ---- |
| Go   | ≥ 1.24   | 服务端运行环境 |
| Node.js | ≥ 18  | 前端构建/开发环境（建议自带 npm） |
| MySQL | 5.7 / 8.x | 默认数据库（也支持 pgsql 等在 `config.yaml` 中切换） |

> 目录说明：服务端代码在 `server/`，前端代码在 `web/`。

## 二、启动服务端（server）

1. **进入服务端目录**

   ```bash
   cd server
   ```

2. **下载依赖**

   ```bash
   go mod tidy
   go mod download
   ```

3. **（可选）提前准备好数据库**

   可以先在 MySQL 中手动建库（如 `CREATE DATABASE gin-vue-admin;`），也可以跳过此步，稍后通过前端页面初始化向导自动建库建表（推荐，见第四节）。

   > 注意：`server/config.yaml` 中的 mysql 连接信息默认为空。**未初始化之前请勿手动修改数据库信息**，推荐使用页面初始化向导；如需手动初始化请参考官方文档 https://gin-vue-admin.com/docs/first_master

4. **启动服务**

   ```bash
   go run .
   ```

   启动成功后服务端默认监听 **8888** 端口（由 `server/config.yaml` 中 `system.addr: 8888` 决定，如被占用可在该文件修改）。

   若需要编译后再运行：

   ```bash
   go build -o server .
   ./server
   ```

## 三、启动前端（web）

1. **进入前端目录**

   ```bash
   cd web
   ```

2. **安装依赖**

   ```bash
   npm install
   ```

   > 国内网络较慢时可以使用淘宝镜像：`npm install --registry=https://registry.npmmirror.com`

3. **启动开发服务**

   ```bash
   npm run dev
   ```

   启动成功后前端默认运行在 **http://localhost:8080**（端口由 `web/.env.development` 中 `VITE_CLI_PORT` 决定）。

   开发模式下 Vite 已将 `/api` 请求自动代理到 `http://127.0.0.1:8888`（见 `web/.env.development` 的 `VITE_BASE_PATH` 与 `VITE_SERVER_PORT`），因此无需额外配置跨域。

## 四、访问系统并完成初始化

1. 浏览器访问前端地址：**http://localhost:8080**
2. 首次运行会自动跳转到初始化页面（`/init`），按页面提示填写 MySQL 连接信息（地址、端口、数据库名、用户名、密码），点击初始化即可自动建库建表并写入初始数据。
3. 初始化完成后重启一次服务端（`Ctrl + C` 后重新 `go run .`），刷新页面进入登录页。
4. 使用默认管理员账号登录：
   - 账号：**admin**
   - 密码：**123456**

## 五、常用补充

- **修改后端端口**：编辑 `server/config.yaml` 的 `system.addr`；同时同步修改 `web/.env.development` 的 `VITE_SERVER_PORT`（若不是 8888）。
- **修改前端端口**：编辑 `web/.env.development` 的 `VITE_CLI_PORT`。
- **启用 Redis / Mongo / 多数据库**：在 `server/config.yaml` 中将 `system.use-redis`、`use-mongo`、`db-list` 等按注释配置后重启。
- **生产构建**：前端执行 `npm run build` 产物在 `web/dist`；后端 `go build` 后配合 `config.docker.yaml` 或 `deploy/` 下的编排文件部署。

## 六、Docker 运行（复用本机已有的 PostgreSQL 与 Redis）

本机已存在独立的 PostgreSQL（容器 `postgres`，映射宿主机 5432，用户 `postgres`，密码 `chdq5201`）和 Redis（容器 `redis`，映射宿主机 6379，密码 `chdq5201`）。Docker 部署方案已改为**复用它们，不再新建 mysql/redis 容器**，`deploy/docker-compose/docker-compose.yaml` 中只有 `gva-web` 和 `gva-server` 两个容器：

- `gva-server` 通过 `extra_hosts: host.docker.internal:host-gateway` 访问宿主机端口，从而连到本机的 PostgreSQL(5432) 与 Redis(6379)。
- 运行配置使用宿主机的 `deploy/docker-compose/config.docker.yaml`（挂载进容器），`db-type` 已改为 `pgsql`、`use-redis: true`。

### 启动步骤

```bash
cd deploy/docker-compose
docker compose up -d --build
```

### 首次初始化（只需一次）

1. 浏览器打开 **http://localhost:8080**，会跳转到初始化向导，按以下信息填写：

   | 项 | 值 |
   | -- | -- |
   | 数据库类型 | pgsql |
   | 服务器地址 | host.docker.internal（容器内经此访问宿主机的 5432） |
   | 端口 | 5432 |
   | 用户名 | postgres |
   | 密码 | chdq5201 |
   | 数据库名 | gin_vue_admin（不存在会自动创建） |

2. 提交后系统自动建库建表并写入初始数据，同时把连接信息**写回** `deploy/docker-compose/config.docker.yaml`。
3. 初始化完成后即可登录：账号 **admin** / 密码 **123456**。

之后重启、重建容器（如 `docker compose up -d --force-recreate`）都会直接复用已有的库和配置，无需再次初始化。

### 注意事项

- 本方案的端口占用与本地开发模式相同（web 8080、server 8888），两种方式不要同时启动。
- `docker compose down` 只会删掉 gva-web/gva-server 容器，不会影响本机的 PostgreSQL 和 Redis 数据。
- 为让镜像构建通过所做的固定：`web/Dockerfile` 基础镜像已升为 `node:22-slim`（corepack 默认拉取的 pnpm 11 不支持 Node 20）；`web/pnpm-workspace.yaml` 放行了依赖构建脚本（pnpm 11 默认拦截会直接报错）；`web/package.json` 新增了直接依赖 `xe-utils`（`gvaGrid` 组件需要）；`web/.dockerignore` 排除了 `dist/`、`.git/`。
