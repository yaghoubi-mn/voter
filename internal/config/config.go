package config

import (
	"log"
	"strconv"
	"time"

	"github.com/yaghoubi-mn/voter/internal/enums"
	"github.com/yaghoubi-mn/voter/internal/models"
	"gorm.io/gorm"
)

const (
	JWTRefreshExpireTime = 7 * 24 * time.Hour
	JWTAccessExpireTime  = 1000 * time.Minute // reduce this in production
	PostTrendingGravity  = 1.8

	PageLimit = 20

	TimeFormat = "2006-01-02 15:04"

	RedisExpiration = 10 * time.Minute
)

type Settings struct {
	SubCreationPermission enums.Permissions
	SubClosePermission    enums.Permissions
	SubDeletePermission   enums.Permissions
}

func (s *Settings) LoadFromDB(db *gorm.DB) error {
	var list []models.Setting

	if err := db.Model(&models.Setting{}).Find(&list).Error; err != nil {
		return err
	}

	for _, item := range list {
		switch item.Key {
		case "SubCreationPermission":
			v, err := strconv.Atoi(item.Value)
			if err != nil {
				log.Fatalf("error in converting %s to int: %s", item.Key, err.Error())
			}
			s.SubCreationPermission = enums.Permissions(v)

		case "SubClosePermission":
			v, err := strconv.Atoi(item.Value)
			if err != nil {
				log.Fatalf("error in converting %s to int: %s", item.Key, err.Error())
			}
			s.SubClosePermission = enums.Permissions(v)

		case "SubDeletePermission":
			v, err := strconv.Atoi(item.Value)
			if err != nil {
				log.Fatalf("error in converting %s to int: %s", item.Key, err.Error())
			}
			s.SubDeletePermission = enums.Permissions(v)

		}
	}

	return nil
}
