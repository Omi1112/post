package db

import (
	"fmt"
	"os"

	"github.com/SeijiOmi/posts-service/entity"
	"github.com/jinzhu/gorm"

	// gormのmysql接続用インポート
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

var (
	db  *gorm.DB
	err error
)

// Init DB接続設定
func Init() {
	DBMS := "mysql"
	USER := os.Getenv("DB_USER")
	PASS := os.Getenv("DB_PASSWORD")
	PROTOCOL := "tcp(" + os.Getenv("DB_ADDRESS") + ")"
	DBNAME := os.Getenv("DB_NAME")
	CONNECT := USER + ":" + PASS + "@" + PROTOCOL + "/" + DBNAME + "?parseTime=true"
	fmt.Println(CONNECT)
	_db, err := gorm.Open(DBMS, CONNECT)
	db = _db
	if err != nil {
		panic(err)
	}

	autoMigration()
}

// GetDB DB接続情報取得
func GetDB() *gorm.DB {
	return db
}

// Close DB切断
func Close() {
	if err := db.Close(); err != nil {
		panic(err)
	}
}

func autoMigration() {
	db.AutoMigrate(&entity.Post{})
}
