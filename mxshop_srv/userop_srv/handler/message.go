package handler

import (
	"awesomeProject/shengxian/mxshop_srv/userop_srv/global"
	"awesomeProject/shengxian/mxshop_srv/userop_srv/model"
	"awesomeProject/shengxian/mxshop_srv/userop_srv/proto/message"
	"context"
)

func (*UserOpServer) MessageList(ctx context.Context, req *message.MessageRequest) (*message.MessageListResponse, error) {
	var rsp message.MessageListResponse
	var messages []model.LeavingMessages
	var messageList  []*message.MessageResponse

	result := global.DB.Where(&model.LeavingMessages{User:req.UserId}).Find(&messages)
	rsp.Total = int32(result.RowsAffected)

	for _, message1 := range messages {
		messageList = append(messageList, &message.MessageResponse{
			Id:          message1.ID,
			UserId:      message1.User,
			MessageType: message1.MessageType,
			Subject:     message1.Subject,
			Message:     message1.Message,
			File:        message1.File,
		})
	}

	rsp.Data = messageList
	return &rsp, nil
}


func (*UserOpServer) CreateMessage(ctx context.Context, req *message.MessageRequest) (*message.MessageResponse, error) {
	var msg model.LeavingMessages

	msg.User = req.UserId
	msg.MessageType = req.MessageType
	msg.Subject = req.Subject
	msg.Message = req.Message
	msg.File = req.File

	global.DB.Save(&msg)

	return &message.MessageResponse{Id:msg.ID}, nil
}