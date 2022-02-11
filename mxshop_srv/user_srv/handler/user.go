package handler

import (
	"awesomeProject/shengxian/mxshop_srv/user_srv/global"
	"awesomeProject/shengxian/mxshop_srv/user_srv/model"
	userpb "awesomeProject/shengxian/mxshop_srv/user_srv/proto/user"
	"context"
	"crypto/sha512"
	"errors"
	"fmt"
	"github.com/anaskhan96/go-password-encoder"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"gorm.io/gorm"
	"strings"
	"time"
)


type Service struct {

}

func Model2Response(user model.User)userpb.UserInfoResponse{
	userInfoRsp := userpb.UserInfoResponse{
		Id: user.ID,
		Password: user.Password,
		Nickname: user.NickName,
		Gender: user.Gender,
		Mobile: user.Mobile,
		Role: int32(user.Role),
	}
	if user.Birthday != nil{
		userInfoRsp.Birthday = uint64(user.Birthday.Unix())
	}
	return userInfoRsp
}

func Paginate(page, pageSize int) func(db *gorm.DB) *gorm.DB {
	return func (db *gorm.DB) *gorm.DB {
		if page == 0 {
			page = 1
		}

		switch {
		case pageSize > 100:
			pageSize = 100
		case pageSize <= 0:
			pageSize = 10
		}

		offset := (page - 1) * pageSize
		return db.Offset(offset).Limit(pageSize)
	}
}


//	GetUserList 获取用户list
func (s *Service) GetUserList(req context.Context, PageInfo *userpb.PageInfo) (*userpb.UserListResponse, error){
	var users []model.User
	result := global.DB.Find(&users)
	if result.Error != nil{
		return nil,result.Error
	}
	rsp := &userpb.UserListResponse{}
	rsp.Total = int32(result.RowsAffected)

	global.DB.Scopes(Paginate(int(PageInfo.Pn),int(PageInfo.PSize))).Find(&users)
	for _,user := range users{
		userInfoRsp := Model2Response(user)
		rsp.Data = append(rsp.Data,&userInfoRsp)
	}
	return rsp,nil
}

//	GetUserByMobile 通过手机号查找用户
func (s *Service)GetUserByMobile(c context.Context, req *userpb.MobileRequest) (*userpb.UserInfoResponse, error){
	var user model.User
	result := global.DB.Where(model.User{
		Mobile:req.Mobile,
	}).First(&user)
	if result.RowsAffected == 0 {
		return nil,status.Errorf(codes.NotFound,"用户不存在")
	}
	if result.Error != nil{
		return nil,result.Error
	}
	userInfoRsp := Model2Response(user)
	return &userInfoRsp,nil
}

//	GetUserByID 通过用户ID查找用户
func (s *Service) GetUserByID(c context.Context, req *userpb.IDRequest) (*userpb.UserInfoResponse, error){
	var user model.User
	//通过id查询用户可以不写where
	result := global.DB.First(&user,req.Id)
	if result.RowsAffected == 0 {
		return nil,status.Error(codes.NotFound,"")
	}
	if result.Error != nil{
		return nil,result.Error
	}
	rspUserInfo := Model2Response(user)
	return &rspUserInfo,nil
}
//	CreateUser 新建用户
func (s *Service)CreateUser(c context.Context, req *userpb.CreateUserRequest) (*userpb.UserInfoResponse, error){
	var user model.User
	result := global.DB.Where(&model.User{
		Mobile: req.Mobile,
	}).First(&user)
	if result.RowsAffected == 1 {
		return nil,status.Error(codes.AlreadyExists,"用户已存在")
	}
	user.Mobile = req.Mobile
	user.NickName = req.Nickname

	//密码加密
	options := &password.Options{16,100,32,sha512.New}
	salt, encodePwd := password.Encode(req.Password,options)
	newPwd := fmt.Sprintf("$pbkdf2-sha512$%s$%s",salt,encodePwd)
	user.Password = newPwd

	result = global.DB.Create(&user)
	if result.Error != nil{
		return nil,status.Error(codes.Internal,result.Error.Error())
	}
	userInfoRsp := Model2Response(user)
	return &userInfoRsp,nil
}

// UpDateUserInfo 个人中心更新用户
func (s *Service) UpDateUserInfo(c context.Context, req *userpb.UpDateUserInfoRequest) (*emptypb.Empty, error){
	var user model.User
	result := global.DB.First(&user,req.Id)
	if result.RowsAffected == 0 {
		return nil,status.Error(codes.NotFound,"用户不存在")
	}
	//把uint64类型的值转换成time
	birthday := time.Unix(int64(req.Birthday),0)
	user.NickName = req.Nickname
	user.Birthday = &birthday
	user.Gender = req.Gender

	result = global.DB.Save(&user)
	if result.Error != nil{
		return nil,status.Error(codes.Internal,result.Error.Error())
	}
	return &emptypb.Empty{},nil
}


//	CheckPassword 校验密码
func (s *Service)CheckPassword(c context.Context, req *userpb.CheckPasswordRequest) (*userpb.CheckPasswordResponse, error)  {
	pwdInfo := strings.Split(req.EncryptedPassword,"$")
	options := &password.Options{16,100,32,sha512.New}
	check := password.Verify(req.Password,pwdInfo[2],pwdInfo[3],options)
	if check == false{
		return &userpb.CheckPasswordResponse{Success: check},errors.New("校验出错")
	}
	return &userpb.CheckPasswordResponse{Success: check},nil
}