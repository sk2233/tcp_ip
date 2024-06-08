/*
@author: sk
@date: 2024/6/4
*/
package main

// https://www.saminiir.com/lets-code-tcp-ip-stack-1-ethernet-arp/  断更教程不太推荐
// 单线程模型 注意这里使用的点对点模式，没有链路层，直接就是 ip 层 抓包是有本地回环层的，但是 water 会自动移除它
// http://22.22.22.22/text.html
// http://22.22.22.22/img.html
// http://22.22.22.22/img.jpeg

func main() {
	inst := NewTun(RemoteIP, LocalIP)
	data := make([]byte, MaxPackageSize)
	Init()
	for {
		n, err := inst.Read(data)
		HandleErr(err)
		reader := NewByteReader(data[:n])
		// 层层处理  使用的是本地网络，没有链路层，使用本地环路  点对点默认会丢弃链路层的数据 直接就是 ip 层了
		//HandleEther(reader)
		//HandleLoop(reader)
		HandleIpv4(&Param{Inst: inst}, reader)
	}
}

func Init() {
	RegisterUDPHandler(2233, EchoUDPHandler)
	RegisterTCPHandler(80, HTTPTCPHandler)
}
