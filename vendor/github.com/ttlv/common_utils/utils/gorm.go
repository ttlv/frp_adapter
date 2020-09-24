package utils

import (
	"strings"

	"github.com/jinzhu/gorm"
)

type Sum struct {
	Value float64 `json:"value"`
}

type Count struct {
	Value int64
}

type String struct {
	Value string
}

func RunSQL(db *gorm.DB, sqls string) {
	for _, sql := range strings.Split(sqls, "\n") {
		if strings.TrimSpace(sql) != "" {
			db.Exec(strings.TrimSpace(sql))
		}
	}
}
