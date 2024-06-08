/*
@author: sk
@date: 2024/6/6
*/
package main

import (
	"fmt"
	"math/rand"
	"os"
	"strings"
)

type TCPHdr struct { // 20byte
	SrcPort   uint16
	DstPort   uint16
	SeqNum    uint32 // 发生使用
	AckNum    uint32 // 响应使用
	HdrLen    uint8  // 4 头长度因为可变长协议 * 4
	Unused    uint8  // 4
	Flags     uint8
	WinSize   uint16
	Checksum  uint16
	UrgentPtr uint16
	// 选项数据不管
}

type TCPHandler func(block *TCPBlock, param *Param, reader *ByteReader)

type TCPBlock struct {
	State   TCPState
	Handler TCPHandler
	SrcPort uint16
	AckNum  uint32
	SeqNum  uint32
	WinSize uint16
	Data    []byte // 要传输的数据，暂存
	Index   int    // 传输的进度
}

var (
	// 只用于监听连接的 block    local_ip
	tcpListenBlocks = make(map[uint16]*TCPBlock)
	// 实际工作的 block       local_ip   remote_ip
	tcpWorkBlocks = make(map[uint16]map[uint16]*TCPBlock)
)

func RegisterTCPHandler(port uint16, handler TCPHandler) {
	tcpListenBlocks[port] = &TCPBlock{Handler: handler, State: StateListen} // 默认一开始就处于监听状态
	tcpWorkBlocks[port] = make(map[uint16]*TCPBlock)                        // 创建其工作集群
}

func HandleTCP(param *Param, reader *ByteReader) {
	hdr := ReadTCPHdr(reader)
	// 移出全部的头数据
	reader.ReadByte(int(hdr.HdrLen*4 - 20))
	param.SrcPort = hdr.SrcPort
	param.DstPort = hdr.DstPort
	param.TCPHdr = hdr
	CheckTCP(hdr, reader)
	if listenBlock, ok := tcpListenBlocks[hdr.DstPort]; ok { // 有对应的监听端口
		if workBlock, ok0 := tcpWorkBlocks[hdr.DstPort][hdr.SrcPort]; ok0 {
			workBlock.Handler(workBlock, param, reader) // 已经有对应的工作节点，继续工作
		} else {
			listenBlock.Handler(listenBlock, param, reader) // 还没有通过监听节点创建工作节点
		}
	} else {
		panic(fmt.Sprintf("invalid TCP port %d", hdr.DstPort))
	}
}

func CheckTCP(hdr *TCPHdr, reader *ByteReader) {
	// TODO checksum
}

func ReadTCPHdr(reader *ByteReader) *TCPHdr {
	srcPort := reader.ReadU16()
	dstPort := reader.ReadU16()
	seqNum := reader.ReadU32()
	ackNum := reader.ReadU32()
	u8 := reader.ReadU8()
	flags := reader.ReadU8()
	winSize := reader.ReadU16()
	checksum := reader.ReadU16()
	urgentPtr := reader.ReadU16()
	return &TCPHdr{
		SrcPort:   srcPort,
		DstPort:   dstPort,
		SeqNum:    seqNum,
		AckNum:    ackNum,
		HdrLen:    (u8 >> 4) & 0xF,
		Unused:    u8 & 0xF,
		Flags:     flags,
		WinSize:   winSize,
		Checksum:  checksum,
		UrgentPtr: urgentPtr,
	}
}

func HTTPTCPHandler(block *TCPBlock, param *Param, reader *ByteReader) {
	if block.State == StateListen {
		// 试图处理第一次握手，第一次握手需要初始化一些数据，需要提前阻拦
		HTTPTCPAccept(block, param, reader)
		return
	}
	hdr := param.TCPHdr // 第一次初始化的信息现在有用了，进行基本校验
	if hdr.SrcPort != block.SrcPort || hdr.SeqNum != block.AckNum || hdr.AckNum != block.SeqNum {
		fmt.Printf("err hdr = %+v , block = %+v\n", hdr, block)
		return
	}
	switch block.State {
	case StateSyn: // 第二次握手已经发了，试图进行第 3 次握手
		// 确认建立连接，切转状态建立链接，这次可以传输数据，但是一般不传输
		block.State = StateEstablish
	case StateEstablish:
		HTTPTCPHandleData(block, param, reader)
	case StateFinWait1: // 第一次挥手结束，对方发来了第二次挥手
		HTTPTCPFinWait1(block, param, reader)
	case StateFinWait2:
		HTTPTCPFinWait2(block, param, reader)
	default:
		panic(fmt.Sprintf("invalid state %v", block.State))
	}
}

func HTTPTCPFinWait2(block *TCPBlock, param *Param, reader *ByteReader) {
	hdr := param.TCPHdr
	if hdr.Flags&FlagFin == 0 { // fin wait2 中对方必须发 fin  暂时不考虑对方还有很多数据没发的情况
		fmt.Printf("err fin wait2 resp hdr = %+v\n", hdr)
		return
	} // 响应对方的关闭请求就好了
	block.AckNum = GetAckNum(hdr, reader)
	writer := NewByteWriter()
	hdr0 := &TCPHdr{
		SrcPort: hdr.DstPort,
		DstPort: hdr.SrcPort,
		SeqNum:  block.SeqNum,
		AckNum:  block.AckNum,
		HdrLen:  5,
		Flags:   FlagAck,
		WinSize: DefaultWinSize,
	}
	block.SeqNum += GetSeqNum(hdr0, writer)
	SendTCP(param, writer, hdr0)
	// 这里就不 wait 了，自己发的ack对方肯定能接收到,工作完成丢弃节点
	delete(tcpWorkBlocks[hdr.DstPort], hdr.SrcPort)
}

func HTTPTCPFinWait1(block *TCPBlock, param *Param, reader *ByteReader) {
	hdr := param.TCPHdr
	if hdr.Flags&FlagAck == 0 { // 对方至少要接收了，前面的 fin
		fmt.Printf("err fin wait1 resp hdr = %+v\n", hdr)
		return
	}
	if hdr.Flags&FlagFin == 0 { // 对方仅是接受到了我们的 fin 但是他还有数据传输 进入 fin_wait2
		block.State = StateFinWait2
		return
	} // 对方不仅回应我们的fin且表示自己也fin了    回应ack进入time_wait状态
	block.AckNum = GetAckNum(hdr, reader)
	writer := NewByteWriter()
	hdr0 := &TCPHdr{
		SrcPort: hdr.DstPort,
		DstPort: hdr.SrcPort,
		SeqNum:  block.SeqNum,
		AckNum:  block.AckNum,
		HdrLen:  5,
		Flags:   FlagAck,
		WinSize: DefaultWinSize,
	}
	block.SeqNum += GetSeqNum(hdr0, writer)
	SendTCP(param, writer, hdr0)
	// 这里就不 wait 了，自己发的ack对方肯定能接收到,工作完成丢弃节点
	delete(tcpWorkBlocks[hdr.DstPort], hdr.SrcPort)
}

func HTTPTCPHandleData(block *TCPBlock, param *Param, reader *ByteReader) {
	hdr := param.TCPHdr // 核心处理数据的流程
	block.AckNum = GetAckNum(hdr, reader)
	if path, ok := ReadPath(reader); ok { // 请求的话就拿一下数据，普通响应就继续传输数据
		fmt.Println(path)
		bs, err := os.ReadFile("res/" + path)
		if err != nil {
			bs = []byte(err.Error())
		}
		block.Data = bs
		block.Index = 0
	}
	writer := NewByteWriter()
	writer.WriteByte(ReadData(block))
	hdr0 := &TCPHdr{
		SrcPort: hdr.DstPort,
		DstPort: hdr.SrcPort,
		SeqNum:  block.SeqNum,
		AckNum:  block.AckNum,
		HdrLen:  5,
		Flags:   FlagAck,
		WinSize: DefaultWinSize,
	}
	if !HasData(block) { // 没数据了进行关闭
		hdr0.Flags |= FlagFin
		block.State = StateFinWait1
	}
	block.SeqNum += GetSeqNum(hdr0, writer)
	SendTCP(param, writer, hdr0)
}

func HasData(block *TCPBlock) bool {
	return block.Index < len(block.Data)
}

func ReadData(block *TCPBlock) []byte {
	size := 1500 - 20 - 20         // 最多发满一个包
	if size > int(block.WinSize) { // 要考虑对方的窗口大小
		size = int(block.WinSize)
	}
	if size > len(block.Data[block.Index:]) { // 要考虑还剩多少
		size = len(block.Data[block.Index:])
	}
	start := block.Index
	end := start + size
	block.Index = end
	return block.Data[start:end]
}

func ReadPath(reader *ByteReader) (string, bool) {
	str := string(reader.ReadLast())
	start := strings.IndexRune(str, '/')
	if start < 0 {
		return "", false
	}
	end := strings.IndexRune(str[start+1:], ' ')
	return str[start+1 : start+1+end], true
}

func HTTPTCPAccept(block *TCPBlock, param *Param, reader *ByteReader) {
	hdr := param.TCPHdr
	if hdr.Flags&FlagSyn > 0 { // 确实是第一次握手，进行处理
		// 初始化各种数据，外面的 block是监听节点，这里需要创建新的 work_block 来执行具体业务
		block = &TCPBlock{
			State:   StateSyn,
			Handler: block.Handler, // 复制其处理函数
			SrcPort: hdr.SrcPort,
			AckNum:  GetAckNum(hdr, reader),
			SeqNum:  rand.Uint32(), // 初始发送序号是随机的
			WinSize: hdr.WinSize,
		}
		tcpWorkBlocks[hdr.DstPort][hdr.SrcPort] = block
		fmt.Printf("accept tcp work count %d\n", len(tcpWorkBlocks[hdr.DstPort]))
		// 第二次握手
		writer := NewByteWriter()
		hdr0 := &TCPHdr{
			SrcPort: hdr.DstPort,
			DstPort: hdr.SrcPort,
			SeqNum:  block.SeqNum,
			AckNum:  block.AckNum,
			HdrLen:  5,
			Flags:   FlagAck | FlagSyn,
			WinSize: DefaultWinSize,
		}
		block.SeqNum += GetSeqNum(hdr0, writer)
		SendTCP(param, writer, hdr0)
	} else { // 非法请求，重置连接状态，重来 可以发送 reset报文进行重置
		panic(fmt.Sprintf("invalid TCP flags: %d", hdr.Flags))
	}
}

func SendTCP(param *Param, writer *ByteWriter, hdr *TCPHdr) {
	WriteTCPHdr(writer, hdr)
	fakeHdr := &FakeHdr{
		SrcAddr:  param.DstAddr,
		DstAddr:  param.SrcAddr,
		Protocol: Ipv4TCP,
		Total:    writer.Len(),
	}
	WriteFakeHdr(writer, fakeHdr)
	hdr.Checksum = GetChecksum(writer.GetData())
	writer.Seek(20 + 12)
	WriteTCPHdr(writer, hdr)
	hdr0 := &IPv4Hdr{
		Version:  4,
		HdrLen:   5,
		Len:      writer.Len() + 20,
		Ttl:      64,
		Protocol: Ipv4TCP,
		SrcAddr:  param.DstAddr,
		DstAddr:  param.SrcAddr,
	}
	SendIpv4(param, writer, hdr0)
}

func WriteTCPHdr(writer *ByteWriter, hdr *TCPHdr) {
	writer.WriteU16(hdr.UrgentPtr)
	writer.WriteU16(hdr.Checksum)
	writer.WriteU16(hdr.WinSize)
	writer.WriteU8(hdr.Flags)
	writer.WriteU8((hdr.Unused & 0xF) | (hdr.HdrLen << 4))
	writer.WriteU32(hdr.AckNum)
	writer.WriteU32(hdr.SeqNum)
	writer.WriteU16(hdr.DstPort)
	writer.WriteU16(hdr.SrcPort)
}

func GetSeqNum(hdr *TCPHdr, writer *ByteWriter) uint32 {
	res := uint32(writer.Len())
	if hdr.Flags&FlagSyn > 0 || hdr.Flags&FlagFin > 0 {
		res++
	}
	return res
}

func GetAckNum(hdr *TCPHdr, reader *ByteReader) uint32 {
	res := hdr.SeqNum // 默认认为发来的全部已读
	if hdr.Flags&FlagSyn > 0 || hdr.Flags&FlagFin > 0 {
		res++
	}
	return res + uint32(reader.Len())
}
