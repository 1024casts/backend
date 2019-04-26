package service

import (
	"sync"
	"time"

	"fmt"

	"github.com/1024casts/1024casts/model"
	"github.com/1024casts/1024casts/pkg/auth"
	"github.com/1024casts/1024casts/repository"
	"github.com/1024casts/1024casts/util"
	"github.com/lexkong/log"
	"github.com/spf13/viper"
	"gopkg.in/gomail.v2"
)

type UserService struct {
	userRepo *repository.UserRepo
}

func NewUserService() *UserService {
	return &UserService{
		repository.NewUserRepo(),
	}
}

func (srv *UserService) CreateUser(user model.UserModel) (id uint64, err error) {
	id, err = srv.userRepo.CreateUser(user)

	if err != nil {
		return id, err
	}

	return id, nil
}

func (srv *UserService) RegisterUser(user model.UserModel) (id uint64, err error) {
	token, err := util.GenShortId()
	if err != nil {
		log.Warnf("[user] gen user activation token err: %v", err)
		return 0, err
	}

	db := model.DB.Self
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if tx.Error != nil {
		return 0, err
	}

	if err := tx.Create(&user).Error; err != nil {
		tx.Rollback()
		log.Warnf("[user] create user err: %v", err)
		return 0, err
	}

	// 注册时候的一些处理
	// 1、写入到激活码到表里
	userActivationModel := model.UserActivationModel{}
	userActivationModel.UserID = user.Id
	userActivationModel.Token = token
	if err := tx.Create(&userActivationModel).Error; err != nil {
		tx.Rollback()
		log.Warnf("[user] add user activation err: %v", err)
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		log.Warnf("[order] commit order fail, err: %+v", err)
		return 0, err
	}

	// 2、发送激活邮件
	go sendActiveMail(user.Username, user.Email, token)

	return user.Id, nil
}

// 发送激活邮件
func sendActiveMail(username, toMail, activeCode string) {
	m := gomail.NewMessage()
	// 发件人
	m.SetAddressHeader("From", "no-reply@phpcasts.org", "1024课堂")
	// 收件人
	m.SetHeader("To",
		m.FormatAddress(toMail, ""),
	)
	// 主题
	m.SetHeader("Subject", "帐号激活链接")
	// 正文
	activeUrl := fmt.Sprintf("https://1024casts.com/users/%s/activation/%s", username, activeCode)
	m.SetBody("text/html", "Hi, "+username+"<br>请激活您的帐号： <a href = '"+activeUrl+"'>"+activeUrl+"</a>")

	// 发送邮件服务器、端口、发件人账号、发件人密码
	d := gomail.NewDialer(viper.GetString("mail.host"), viper.GetInt("mail.port"), viper.GetString("mail.username"), viper.GetString("mail.password"))
	if err := d.DialAndSend(m); err != nil {
		log.Warnf("[register] send active mail err: %v", err)
	}
}

func (srv *UserService) GetUserById(id uint64) (*model.UserModel, error) {
	user, err := srv.userRepo.GetUserById(id)

	if err != nil {
		return user, err
	}

	return user, nil
}

func (srv *UserService) GetUserNameById(id uint64) string {
	user, err := srv.userRepo.GetUserById(id)

	if err != nil {
		log.Warnf("[service] get user info err: %v", err)
		return ""
	}

	return user.Username
}

func (srv *UserService) GetUserByUsername(username string) (*model.UserModel, error) {
	user, err := srv.userRepo.GetUserByUsername(username)

	if err != nil {
		return user, err
	}

	return user, nil
}

func (srv *UserService) GetUserByEmail(email string) (*model.UserModel, error) {
	user, err := srv.userRepo.GetUserByEmail(email)

	if err != nil {
		return user, err
	}

	return user, nil
}

func (srv *UserService) GetUserList(username string, offset, limit int) ([]*model.UserModel, uint64, error) {
	infos := make([]*model.UserModel, 0)
	users, count, err := srv.userRepo.GetUserList(username, offset, limit)
	if err != nil {
		return nil, count, err
	}

	ids := []uint64{}
	for _, user := range users {
		ids = append(ids, user.Id)
	}

	wg := sync.WaitGroup{}
	userList := model.UserList{
		Lock:  new(sync.Mutex),
		IdMap: make(map[uint64]*model.UserModel, len(users)),
	}

	errChan := make(chan error, 1)
	finished := make(chan bool, 1)

	// Improve query efficiency in parallel
	for _, u := range users {
		wg.Add(1)
		go func(u *model.UserModel) {
			defer wg.Done()

			userList.Lock.Lock()
			defer userList.Lock.Unlock()

			u.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", u.CreatedAt.String())
			userList.IdMap[u.Id] = u
		}(u)
	}

	go func() {
		wg.Wait()
		close(finished)
	}()

	select {
	case <-finished:
	case err := <-errChan:
		return nil, count, err
	}

	for _, id := range ids {
		infos = append(infos, userList.IdMap[id])
	}

	return infos, count, nil
}

func (srv *UserService) UpdateUser(userMap map[string]interface{}, id uint64) error {
	err := srv.userRepo.Update(userMap, id)

	if err != nil {
		return err
	}

	return nil
}

func (srv *UserService) UpdateUserPassword(id uint64, password string) error {
	hashedPassword, err := auth.Encrypt(password)
	if err != nil {
		log.Warnf("[user] Encrypt user password err: %v", err)
		return err
	}

	userMap := map[string]interface{}{
		"password": hashedPassword,
	}
	err = srv.userRepo.Update(userMap, id)
	if err != nil {
		return err
	}

	return nil
}

func (srv *UserService) DeleteUser(id uint64) error {
	err := srv.userRepo.DeleteUser(id)

	if err != nil {
		return err
	}

	return nil
}

func (srv *UserService) GetResetPasswordTokenByEmail(email string) (string, error) {
	pwdResetInfo, err := srv.userRepo.GetResetPasswordInfoByEmail(email)

	if err != nil {
		return pwdResetInfo.Token, err
	}

	return "", nil
}
