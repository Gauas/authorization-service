package repository

import "gorm.io/gorm"

func applyArgs(db *gorm.DB, args ...interface{}) *gorm.DB {
	if len(args) == 0 {
		return db
	}

	first := args[0]
	if q, ok := first.(string); ok {
		if q != "" {
			return db.Where(q, args[1:]...)
		}
		return db
	}

	return db.Where(first)
}
