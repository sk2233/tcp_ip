/*
@author: sk
@date: 2024/6/4
*/
package main

const (
	LocalIP  = "22.22.22.22"
	RemoteIP = "33.33.33.33"
)

const (
	// 以太网负载在  46~1500之间
	MaxPackageSize = 6 + 6 + 2 + 1500
	MinPackageSize = 6 + 6 + 2 + 46
)

type EtherType uint16

type LoopType uint32

type Ipv4Protocol uint8

const (
	Ipv4ICMP Ipv4Protocol = 0x1
	Ipv4TCP  Ipv4Protocol = 0x6
	Ipv4UDP  Ipv4Protocol = 0x11
)

type ICMPType uint8

const (
	ICMPResp    ICMPType = 0x0
	ICMPUnReach ICMPType = 0x3
	ICMPReq     ICMPType = 0x8
)

const (
	FlagFin = 1 << iota
	FlagSyn
	FlagRst
	FlagPsh
	FlagAck
	FlagUrgent // 紧急数据
	FlagECN
	FlagCWR
)

type TCPState string

const (
	StateListen    TCPState = "StateListen"
	StateSyn       TCPState = "StateSyn"
	StateEstablish TCPState = "StateEstablish"
	StateFinWait1  TCPState = "StateFinWait1"
	StateFinWait2  TCPState = "StateFinWait2"
)

const (
	MaxSegSize     = 1460 // 1500 - tcp_hdr - ip_hdr  最大每个分片的大小
	DefaultWinSize = 1460 // 让对面随便发了，一波带走，暂时只支持比较简单的请求，可以一个包发完
)
