package db

import (
	"fmt"
	"time"

	"lychee-go/internal/config"
	mylog "lychee-go/internal/logger"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormlog "gorm.io/gorm/logger"
)

var DB *gorm.DB

// Init 初始化数据库连接
func Init() error {
	host := config.GetString("database.host", "127.0.0.1")
	port := config.GetInt("database.port", 3306)
	database := config.GetString("database.database", "lychee_go")
	username := config.GetString("database.username", "root")
	password := config.GetString("database.password", "")
	charset := config.GetString("database.charset", "utf8mb4")
	maxIdleConns := config.GetInt("database.max_idle_conns", 10)
	maxOpenConns := config.GetInt("database.max_open_conns", 100)
	connMaxLifetime := config.GetInt("database.conn_max_lifetime", 3600)

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
		username, password, host, port, database, charset)

	gormLogLevel := gormlog.Warn
	switch config.GetString("database.log_mode", "warn") {
	case "silent":
		gormLogLevel = gormlog.Silent
	case "error":
		gormLogLevel = gormlog.Error
	case "info":
		gormLogLevel = gormlog.Info
	}

	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: gormlog.Default.LogMode(gormLogLevel),
	})
	if err != nil {
		return fmt.Errorf("failed to connect database: %w", err)
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}
	sqlDB.SetMaxIdleConns(maxIdleConns)
	sqlDB.SetMaxOpenConns(maxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Duration(connMaxLifetime) * time.Second)

	mylog.Info("Database connected successfully (host=%s, db=%s)", host, database)
	return nil
}

// ======== ThinkPHP 风格 API ========

func Table(name string) *gorm.DB {
	return DB.Table(name)
}

func Model(model interface{}) *gorm.DB {
	return DB.Model(model)
}

func GetDB() *gorm.DB {
	return DB
}

// ======== 查询构建器 ========

type Query struct {
	db *gorm.DB
}

func NewQuery() *Query {
	return &Query{db: DB}
}

func (q *Query) Where(query interface{}, args ...interface{}) *Query {
	q.db = q.db.Where(query, args...)
	return q
}

func (q *Query) WhereIn(field string, values interface{}) *Query {
	q.db = q.db.Where(fmt.Sprintf("%s IN ?", field), values)
	return q
}

func (q *Query) WhereLike(field string, value string) *Query {
	q.db = q.db.Where(fmt.Sprintf("%s LIKE ?", field), "%"+value+"%")
	return q
}

func (q *Query) Order(field string, direction string) *Query {
	q.db = q.db.Order(fmt.Sprintf("%s %s", field, direction))
	return q
}

func (q *Query) Limit(n int) *Query {
	q.db = q.db.Limit(n)
	return q
}

func (q *Query) Offset(n int) *Query {
	q.db = q.db.Offset(n)
	return q
}

func (q *Query) Page(page int, listRows int) *Query {
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * listRows
	q.db = q.db.Offset(offset).Limit(listRows)
	return q
}

func (q *Query) Find(dest interface{}) error {
	return q.db.First(dest).Error
}

func (q *Query) Select(dest interface{}) error {
	return q.db.Find(dest).Error
}

func (q *Query) Count() (int64, error) {
	var count int64
	err := q.db.Count(&count).Error
	return count, err
}

func (q *Query) Insert(data interface{}) (int64, error) {
	result := q.db.Create(data)
	if result.Error != nil {
		return 0, result.Error
	}
	return result.RowsAffected, nil
}

func (q *Query) Update(data interface{}) (int64, error) {
	result := q.db.Updates(data)
	if result.Error != nil {
		return 0, result.Error
	}
	return result.RowsAffected, nil
}

func (q *Query) Delete(model interface{}) (int64, error) {
	result := q.db.Delete(model)
	if result.Error != nil {
		return 0, result.Error
	}
	return result.RowsAffected, nil
}

func Transaction(fn func(tx *gorm.DB) error) error {
	return DB.Transaction(fn)
}

// ======== 模型基类 ========

type BaseModel struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
