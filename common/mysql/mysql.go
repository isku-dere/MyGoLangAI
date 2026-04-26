package mysql

import (
	"GopherAI/config"
	"GopherAI/model"
	"fmt"
	stdlog "log"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func InitMysql() error {
	host := config.GetConfig().MysqlHost
	port := config.GetConfig().MysqlPort
	dbname := config.GetConfig().MysqlDatabaseName
	username := config.GetConfig().MysqlUser
	password := config.GetConfig().MysqlPassword
	charset := config.GetConfig().MysqlCharset

	//dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=true&loc=Local", username, password, host, port, dbname, charset)
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=true&loc=Local", username, password, host, port, dbname, charset)

	var log logger.Interface
	if gin.Mode() == "debug" {
		log = logger.Default.LogMode(logger.Info)
	} else {
		log = logger.Default
	}

	db, err := gorm.Open(mysql.New(mysql.Config{
		DSN:                       dsn,
		DefaultStringSize:         256,
		DisableDatetimePrecision:  true,
		DontSupportRenameIndex:    true,
		DontSupportRenameColumn:   true,
		SkipInitializeWithVersion: false,
	}), &gorm.Config{
		Logger: log,
	})
	if err != nil {
		return err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	DB = db

	return migration()
}

func migration() error {
	if err := DB.AutoMigrate(
		new(model.User),
		new(model.Session),
		new(model.Message),
		new(model.RAGDocument),
		new(model.OCRTask),
	); err != nil {
		return err
	}
	return ensureUserEmailUniqueIndex()
}

func ensureUserEmailUniqueIndex() error {
	var duplicateEmail string
	err := DB.Raw(`
		SELECT email
		FROM users
		WHERE email <> '' AND deleted_at IS NULL
		GROUP BY email
		HAVING COUNT(*) > 1
		LIMIT 1
	`).Scan(&duplicateEmail).Error
	if err != nil {
		return err
	}
	if duplicateEmail != "" {
		stdlog.Printf("skip unique users.email index: duplicated email exists")
		return nil
	}

	type indexInfo struct {
		NonUnique int `gorm:"column:Non_unique"`
	}
	var indexes []indexInfo
	if err := DB.Raw("SHOW INDEX FROM users WHERE Key_name = 'idx_users_email'").Scan(&indexes).Error; err != nil {
		return err
	}
	if len(indexes) > 0 && indexes[0].NonUnique == 0 {
		return nil
	}
	if len(indexes) > 0 {
		if err := DB.Exec("DROP INDEX idx_users_email ON users").Error; err != nil {
			return err
		}
	}
	return DB.Exec("CREATE UNIQUE INDEX idx_users_email ON users(email)").Error
}

func InsertUser(user *model.User) (*model.User, error) {
	err := DB.Create(&user).Error
	return user, err
}

func GetUserByUsername(username string) (*model.User, error) {
	user := new(model.User)
	err := DB.Where("username = ?", username).First(user).Error
	return user, err
}

func GetUserByEmail(email string) (*model.User, error) {
	user := new(model.User)
	err := DB.Where("email = ?", email).First(user).Error
	return user, err
}
