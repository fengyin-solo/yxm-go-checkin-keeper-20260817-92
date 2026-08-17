# Checkin Keeper — 签到打卡中心

纯 Go 标准库实现的签到打卡后端服务，零第三方依赖，开箱即跑。

## 业务说明

管理签到打卡的完整生命周期：**活动创建 → 激活 → 用户签到 → 连续天数累计 → 奖励积分发放 → 统计排行**。

- **用户**：参与签到的主体，持有积分余额。
- **签到活动**：定义签到的时间窗口，状态机 `draft → active → finished`。
- **签到记录**：同一用户同一活动每天仅可签到一次；记录当日连续天数。
- **奖励规则**：每日基础积分（daily_base）与连续签到奖励（streak_bonus，连续天数达到门槛的倍数时触发），支持全局或按活动生效。
- **奖励流水**：每次积分发放的不可变留痕（base/bonus 两类）。

> 积分字段（Points）单位为积分（int64）。

## 运行

```bash
cd origin
go run ./cmd/server
# 默认监听 :8080，可通过 PORT / ADDR 环境变量修改
```

环境变量：

| 变量 | 默认值 | 说明 |
|------|--------|------|
| PORT | 8080 | 监听端口 |
| ADDR | :PORT | 完整监听地址（优先于 PORT） |
| MAX_PAGE_SIZE | 100 | 分页最大条数 |
| LOG_LEVEL | info | 日志级别：debug/info/warn/error |

## API 一览

### 用户

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | /api/users | 创建用户 |
| GET | /api/users | 列表（支持 status/keyword 筛选 + 分页，按积分降序） |
| GET | /api/users/{id} | 详情 |
| PUT | /api/users/{id} | 更新 |
| DELETE | /api/users/{id} | 删除（有签到记录则拒绝） |
| GET | /api/users/{id}/streak | 签到统计（?activity_id=，累计/当前/最大连续天数） |
| GET | /api/users/{id}/calendar | 签到日历（?activity_id=&month=YYYY-MM） |

### 签到活动

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | /api/activities | 创建活动（草稿态） |
| GET | /api/activities | 列表（支持 status/keyword 筛选 + 分页） |
| GET | /api/activities/{id} | 详情 |
| PUT | /api/activities/{id} | 更新（已结束不可编辑） |
| DELETE | /api/activities/{id} | 删除（有签到记录则拒绝） |
| POST | /api/activities/{id}/transition | 状态流转（draft→active→finished） |

### 签到

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | /api/checkins | 签到（校验活动窗口/重复签到，自动发放奖励） |
| GET | /api/checkins | 列表（支持 activity_id/user_id/from_date/to_date 筛选 + 分页） |
| GET | /api/checkins/{id} | 详情 |

### 奖励规则

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | /api/reward-rules | 创建规则（daily_base/streak_bonus） |
| GET | /api/reward-rules | 列表（支持 type/status/activity_id 筛选 + 分页） |
| GET | /api/reward-rules/{id} | 详情 |
| PUT | /api/reward-rules/{id} | 更新 |
| DELETE | /api/reward-rules/{id} | 删除 |

### 奖励流水

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/reward-grants | 列表（支持 user_id/activity_id/type 筛选 + 分页） |
| GET | /api/reward-grants/{id} | 详情 |

### 统计

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/stats/overview | 全局概览（用户/活动/签到数、积分总量） |
| GET | /api/stats/by-activity | 按活动分组统计（签到数、参与人数） |
| GET | /api/stats/by-day | 按天分组统计 |
| GET | /api/stats/leaderboard | 积分排行榜（?n=10） |

### 健康检查

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /healthz | 健康检查 |

## 统一响应格式

```json
{"code": 0, "message": "ok", "data": ...}
```

错误码映射：400 参数校验失败 / 404 记录不存在 / 409 状态冲突或唯一性冲突 / 500 内部错误。

## 工程结构

```
origin/
├── go.mod
├── README.md
├── cmd/server/main.go          # 入口：配置加载、依赖装配、优雅关闭
├── internal/
│   ├── app/app.go              # 依赖装配 store -> service -> handler
│   ├── config/config.go        # 环境变量配置
│   ├── model/                  # 领域模型 + 状态机 + 校验
│   ├── store/                  # Store 接口 + 内存实现
│   ├── service/                # 业务逻辑 + 奖励引擎 + 统计
│   └── handler/                # HTTP 路由 + 处理器
└── pkg/
    ├── httpx/httpx.go          # 统一响应、分页、JSON 解析
    ├── idgen/idgen.go          # Hex ID + base62 短码
    └── logger/logger.go        # 分级日志
```

## 测试

```bash
go test ./...
```
