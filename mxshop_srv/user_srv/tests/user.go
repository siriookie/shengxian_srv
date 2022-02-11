package main


import (
	userpb "awesomeProject/shengxian/mxshop_srv/user_srv/proto/user"
	"context"
	"fmt"
	"google.golang.org/grpc"
)

var userClient userpb.UserClient
var conn *grpc.ClientConn
func Init(){
	var err error
	conn,err = grpc.Dial(":50051",grpc.WithInsecure())

	if err != nil{
		panic("init client failed"+err.Error())
	}
	userClient = userpb.NewUserClient(conn)
}

func TestGetUserList(){
	rsp, err := userClient.GetUserList(context.Background(),&userpb.PageInfo{
		Pn: 1,
		PSize: 2,
	})
	if err != nil{
		panic(err)
	}
	fmt.Println(rsp)
	for _,val := range rsp.Data{
		fmt.Println(val)
		rsp1,err := userClient.CheckPassword(context.Background(),&userpb.CheckPasswordRequest{
			Password: "123",
			EncryptedPassword: val.Password,
		})
		if err != nil{
			panic("校验密码出错"+err.Error())
		}
		fmt.Println(rsp1.Success)
	}
}

func TestCreateUser(){
	for i:=0;i<10;i++{
		rsp,err := userClient.CreateUser(context.Background(),&userpb.CreateUserRequest{
			Password: "123",
			Nickname: fmt.Sprintf("test_rpc%d",i),
			Mobile: fmt.Sprintf("1533456333%d",i),
		})
		if err != nil{
			panic(err)
		}
		fmt.Println(rsp.Id)
	}
}

func main() {
	Init()
	defer conn.Close()
	TestGetUserList()
	//TestCreateUser()

}