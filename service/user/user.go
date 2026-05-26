package user

import (
	"GopherAI/common/code"
	myemail "GopherAI/common/email"
	myredis "GopherAI/common/redis"
	"GopherAI/dao/user"
	"GopherAI/model"
	"GopherAI/utils"
	"GopherAI/utils/myjwt"
	"strings"
)

func Login(username, password string) (string, code.Code) {
	var userInformation *model.User
	var ok bool
	username = strings.TrimSpace(username)
	//1:判断用户是否存在
	if ok, userInformation = user.FindByUsernameOrEmail(username); !ok {

		return "", code.CodeUserNotExist
	}
	//2:判断用户是否密码账号正确
	if userInformation.Password != utils.MD5(password) {
		return "", code.CodeInvalidPassword
	}
	//3:返回一个Token
	token, err := myjwt.GenerateToken(userInformation.ID, userInformation.Username)

	if err != nil {
		return "", code.CodeServerBusy
	}
	return token, code.CodeSuccess
}

func Register(email, password, captcha string) (string, string, code.Code) {

	var err error
	var userInformation *model.User

	//1:先判断邮箱是否已经注册过
	if ok, _ := user.IsExistEmail(email); ok {
		return "", "", code.CodeUserExist
	}

	//2:从redis中验证验证码是否有效
	if ok, _ := myredis.CheckCaptchaForEmail(email, captcha); !ok {
		return "", "", code.CodeInvalidCaptcha
	}

	//3：生成11位的账号
	username := utils.GetRandomNumbers(11)

	//4：注册到数据库中
	if userInformation, err = user.Register(username, email, password); err != nil {
		if isDuplicateUserError(err) {
			return "", "", code.CodeUserExist
		}
		return "", "", code.CodeServerBusy
	}

	//5：将账号一并发送到对应邮箱上去，后续需要账号登录
	if err := myemail.SendCaptcha(email, username, user.UserNameMsg); err != nil {
		return "", "", code.CodeServerBusy
	}

	// 6:生成Token
	token, err := myjwt.GenerateToken(userInformation.ID, userInformation.Username)

	if err != nil {
		return "", "", code.CodeServerBusy
	}

	return token, username, code.CodeSuccess
}

func isDuplicateUserError(err error) bool {
	if err == nil {
		return false
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "duplicate") || strings.Contains(message, "1062")
}

// 往指定邮箱发送验证码
// 分为以下任务：
// 1：先存放redis
// 2：再进行远程发送
func SendCaptcha(email_ string) code.Code {
	send_code := utils.GetRandomNumbers(6)
	//1:先存放到redis
	if err := myredis.SetCaptchaForEmail(email_, send_code); err != nil {
		return code.CodeServerBusy
	}

	//2:再进行远程发送
	if err := myemail.SendCaptcha(email_, send_code, myemail.CodeMsg); err != nil {
		return code.CodeServerBusy
	}

	return code.CodeSuccess
}
