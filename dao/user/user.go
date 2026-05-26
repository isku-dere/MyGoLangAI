package user

import (
	"GopherAI/common/mysql"
	"GopherAI/model"
	"GopherAI/utils"
	"context"
	"errors"

	"gorm.io/gorm"
)

const (
	CodeMsg     = "GopherAI验证码如下(验证码仅限于2分钟有效): "
	UserNameMsg = "GopherAI的账号如下，请保留好，后续可以用账号进行登录 "
)

var ctx = context.Background()

// 这边只能通过账号进行登录
func IsExistUser(username string) (bool, *model.User) {

	user, err := mysql.GetUserByUsername(username)

	if errors.Is(err, gorm.ErrRecordNotFound) || user == nil {
		return false, nil
	}

	return true, user
}

func FindByUsernameOrEmail(identifier string) (bool, *model.User) {
	if ok, user := IsExistUser(identifier); ok {
		return true, user
	}
	return IsExistEmail(identifier)
}

func IsExistEmail(email string) (bool, *model.User) {
	user, err := mysql.GetUserByEmail(email)

	if errors.Is(err, gorm.ErrRecordNotFound) || user == nil {
		return false, nil
	}

	return true, user
}

func Register(username, email, password string) (*model.User, error) {
	user, err := mysql.InsertUser(&model.User{
		Email:    email,
		Name:     username,
		Username: username,
		Password: utils.MD5(password),
	})
	return user, err
}
