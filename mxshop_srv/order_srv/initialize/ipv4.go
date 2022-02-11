package initialize

import (
	"awesomeProject/shengxian/mxshop_srv/order_srv/global"
	"net"
)

func InitLocalIPv4Address() {
	//获取所有网卡
	addrs, err := net.InterfaceAddrs()
	if err != nil{
		panic(err.Error())
	}

	//遍历
	for _, addr := range addrs {
		//取网络地址的网卡的信息
		ipNet, isIpNet := addr.(*net.IPNet)
		//是网卡并且不是本地环回网卡
		if isIpNet && !ipNet.IP.IsLoopback() {
			ipv4 := ipNet.IP.To4()
			//能正常转成ipv4
			if ipv4 != nil {
				global.Ipv4Addr = ipv4.String()
			}
		}
	}

}
