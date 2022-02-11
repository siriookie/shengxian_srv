package handler

import (
	"awesomeProject/shengxian/mxshop_srv/userop_srv/proto/addr"
	"awesomeProject/shengxian/mxshop_srv/userop_srv/proto/message"
	"awesomeProject/shengxian/mxshop_srv/userop_srv/proto/userfav"
)

type UserOpServer struct {
	message.UnimplementedMessageServer
	userfav.UnimplementedUserFavServer
	addr.UnimplementedAddressServer
}
