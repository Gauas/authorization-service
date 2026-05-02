package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"
)

type Repository[T interface{}] struct {
	db *gorm.DB
}

func (r Repository[T]) FindOne(ctx context.Context, args ...interface{}) (*T, error) {
	record := new(T)

	tx := r.db.WithContext(ctx)
	tx = applyArgs(tx, args...)

	err := tx.First(record).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	return record, err
}

func (r Repository[T]) GetAll(ctx context.Context, args ...interface{}) ([]T, error) {
	var records []T

	tx := r.db.WithContext(ctx)
	tx = applyArgs(tx, args...)

	if err := tx.Find(&records).Error; err != nil {
		return nil, err
	}

	return records, nil
}

func (r Repository[T]) Create(ctx context.Context, entity *T) (*T, error) {
	if err := r.db.WithContext(ctx).Create(entity).Error; err != nil {
		return nil, err
	}

	return entity, nil
}

func (r Repository[T]) Update(ctx context.Context, entity *T) (*T, error) {
	if err := r.db.WithContext(ctx).Save(entity).Error; err != nil {
		return nil, err
	}

	return entity, nil
}

func (r Repository[T]) Delete(ctx context.Context, args ...interface{}) error {
	tx := r.db.WithContext(ctx)
	tx = applyArgs(tx, args...)

	return tx.Delete(new(T)).Error
}

func (r Repository[T]) Exists(ctx context.Context, args ...interface{}) (bool, error) {
	var count int64

	tx := r.db.WithContext(ctx)
	tx = applyArgs(tx, args...)

	if err := tx.Model(new(T)).Count(&count).Error; err != nil {
		return false, err
	}

	return count > 0, nil
}
