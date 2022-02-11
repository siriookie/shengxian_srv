package handler

import (
	"awesomeProject/shengxian/mxshop_srv/goods_srv/global"
	"awesomeProject/shengxian/mxshop_srv/goods_srv/model"
	goodspb "awesomeProject/shengxian/mxshop_srv/goods_srv/proto/goods"
	"context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

////轮播图
func (s *GoodsService)BannerList(c context.Context, req *emptypb.Empty) (*goodspb.BannerListResponse, error){
	bannerListRsp := goodspb.BannerListResponse{}

	var banners []model.Banner
	result := global.DB.Find(&banners)
	bannerListRsp.Total = int32(result.RowsAffected)

	var bannerRep []*goodspb.BannerResponse
	for _,banner := range banners{
		bannerRep = append(bannerRep,&goodspb.BannerResponse{
			Id: banner.ID,
			Image: banner.Image,
			Url: banner.Url,
			Index: banner.Index,
		})
	}
	bannerListRsp.Data = bannerRep
	return &bannerListRsp,nil
}
func (s *GoodsService)CreateBanner(c context.Context, req *goodspb.BannerRequest) (*goodspb.BannerResponse, error){
	banner := model.Banner{}

	banner.Image = req.Image
	banner.Index = req.Index
	banner.Url = req.Url

	global.DB.Save(&banner)

	return &goodspb.BannerResponse{Id:banner.ID}, nil
}
func (s *GoodsService) DeleteBanner(c context.Context, req *goodspb.BannerRequest) (*emptypb.Empty, error){
	if result := global.DB.Delete(&model.Banner{}, req.Id); result.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "轮播图不存在")
	}
	return &emptypb.Empty{}, nil
}
func (s *GoodsService) UpdateBanner(c context.Context, req *goodspb.BannerRequest) (*emptypb.Empty, error){
	var banner model.Banner

	if result := global.DB.First(&banner, req.Id); result.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "轮播图不存在")
	}

	if req.Url != "" {
		banner.Url = req.Url
	}
	if req.Image != "" {
		banner.Image = req.Image
	}
	if req.Index != 0 {
		banner.Index = req.Index
	}

	global.DB.Save(&banner)

	return &emptypb.Empty{}, nil
}
