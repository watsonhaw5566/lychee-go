# db · 数据库

基于 [GORM](https://gorm.io/) 封装的数据库访问层，提供类似 ThinkPHP 的快捷查询 API。

## 用法

```go
import "lychee-go/internal/db"

// 由 boot.Boot() 自动初始化
// 获取原始 *gorm.DB 做复杂操作
gormDB := db.DB()

// ---------- 快捷查询 ----------

// 查询单条
var user User
db.Table("users").Where("id = ?", 1).First(&user)

// 原生 SQL
rows, err := db.DB().Raw("SELECT * FROM users WHERE status = ?", 1).Rows()

// ---------- ThinkPHP 风格操作 ----------
users, err := db.Select("users", "status = ?", 1)
affected, err := db.Update("users", map[string]interface{}{"status": 0}, "id = ?", 1)
insertID, err := db.Insert("users", map[string]interface{}{"name": "alice", "age": 20})
deleted, err := db.Delete("users", "id = ?", 1)

// ---------- 事务 ----------
db.Transaction(func(tx *gorm.DB) error {
    tx.Create(&order)
    tx.Model(&stock).Update("quantity", gorm.Expr("quantity - 1"))
    return nil
})
```

## 配置

```yaml
database:
  driver: mysql
  host: 127.0.0.1
  port: 3306
  database: lychee_go
  username: root
  password: "123456"
  charset: utf8mb4
  max_idle_conns: 10
  max_open_conns: 100
  conn_max_lifetime: 3600
  log_mode: warn
```

## 注意事项

- 数据库初始化失败不会让程序崩溃（只会 `logger.Warn`），方便开发时跳过。
- 生产环境 `log_mode` 建议 `warn` 或 `error`，避免 SQL 日志过多。
- 连接池参数根据实际 QPS 调整。
