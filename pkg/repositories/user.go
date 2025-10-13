package repositories

import (
	"context"

	"gorm.io/gorm"
)

// User dùng chung cho customer/employee
type User struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Email       string     `json:"email"`
	PhoneNumber *string    `json:"phone_number"`
	Password    string     `json:"password"`
	AvatarURL   string     `json:"avatar_url"`
	Streak      int        `json:"streak"`
	Score       int        `json:"score"`
}

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) GetByUserID(ctx context.Context, tableName string, id string) (*User, error) {
    var user User
    tx := r.db.WithContext(ctx).Table(tableName).Where("id = ?", id).First(&user)
    if tx.Error != nil {
        if tx.Error == gorm.ErrRecordNotFound {
            return nil, nil
        }
        return nil, tx.Error
    }
    return &user, nil
}

func (r *UserRepository) GetByUsername(ctx context.Context, tableName, username string) (*User, error) {
	var user User
	tx := r.db.WithContext(ctx).Table(tableName).Where("username = ?", username).First(&user)
	if tx.Error != nil {
		if tx.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, tx.Error
	}
	return &user, nil
}


func (r *UserRepository) GetByEmail(ctx context.Context, tableName, email string) (*User, error) {
    var user User
    tx := r.db.WithContext(ctx).Table(tableName).Where("email = ?", email).First(&user)
    if tx.Error != nil {
        if tx.Error == gorm.ErrRecordNotFound {
            return nil, nil
        }
        return nil, tx.Error
    }
    return &user, nil
}

func (r *UserRepository) Create(ctx context.Context, tableName string, user *User) (string, error) {
	tx := r.db.WithContext(ctx).Table(tableName).Create(user)
	if tx.Error != nil {
		return "", tx.Error
	}
	return user.ID, nil
}

func (r *UserRepository) UpdateUserInfor(ctx context.Context, tableName string, user *User) error {
	return r.db.Table(tableName).Save(user).Error
}

func (r *UserRepository) Delete(ctx context.Context, tableName string, id int64) error {
	tx := r.db.WithContext(ctx).Table(tableName).Delete(&User{}, id)
	return tx.Error
}
