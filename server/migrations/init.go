package migrations

import (
	"context"
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
	"mayfly-go/pkg/rediscli"
	"time"
)

// RunMigrations 数据库迁移操作
func RunMigrations(db *gorm.DB) error {
	// 添加分布式锁, 防止多个服务同时执行迁移
	if rediscli.GetCli() != nil {
		if ok, err := rediscli.GetCli().
			SetNX(context.Background(), "migrations", "lock", time.Minute).Result(); err != nil {
			return err
		} else if !ok {
			return nil
		}
		defer rediscli.Del("migrations")
	}
	return run(db,
		T2022,
		T20230720,
	)
}

func run(db *gorm.DB, fs ...func() *gormigrate.Migration) error {
	var ms []*gormigrate.Migration
	for _, f := range fs {
		ms = append(ms, f())
	}
	m := gormigrate.New(db, &gormigrate.Options{
		TableName:                 "migrations",
		IDColumnName:              "id",
		IDColumnSize:              200,
		UseTransaction:            true,
		ValidateUnknownMigrations: true,
	}, ms)
	if err := m.Migrate(); err != nil {
		return err
	}
	return nil
}
