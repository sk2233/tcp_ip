/*
@author: sk
@date: 2024/6/5
*/
package main

import "fmt"

type ICMPHdr struct {
	Type     ICMPType // 控制类似
	Code     uint8    // 具体原因，例如目标不可达的原因
	Checksum uint16
}

func HandleICMP(param *Param, reader *ByteReader) {
	hdr := ReadICMPHdr(reader)
	CheckICMP(hdr, reader)
	switch hdr.Type {
	case ICMPReq:
		HandlePing(param, reader)
	case ICMPResp:
		// 没有主动ping的场景暂时不处理
	case ICMPUnReach:
		// 没有主动请求，也不会响应不可达，暂时不处理
	default:
		panic(fmt.Sprintf("unknown ICMP type: %v", hdr.Type))
	}
}

type Ping struct {
	ID        uint16
	Seq       uint16
	Timestamp []byte // 8  回显时间戳用于计算耗时
	Data      []byte
}

func HandlePing(param *Param, reader *ByteReader) {
	ping := ReadPing(reader)
	writer := NewByteWriter()
	SendPing(param, writer, ping)
}

func SendPing(param *Param, writer *ByteWriter, ping *Ping) {
	WritePing(writer, ping)
	hdr := &ICMPHdr{
		Type: ICMPResp, // checksum交给下层计算
	}
	SendICMP(param, writer, hdr)
}

func SendICMP(param *Param, writer *ByteWriter, hdr *ICMPHdr) {
	WriteICMPHdr(writer, hdr)
	hdr.Checksum = GetChecksum(writer.GetData())
	writer.Seek(4)
	WriteICMPHdr(writer, hdr)
	hdr0 := &IPv4Hdr{
		Version:  4,
		HdrLen:   5,
		Len:      writer.Len() + 20,
		Ttl:      64,
		Protocol: Ipv4ICMP,
		SrcAddr:  param.DstAddr,
		DstAddr:  param.SrcAddr,
	}
	SendIpv4(param, writer, hdr0)
}

func WriteICMPHdr(writer *ByteWriter, icmp *ICMPHdr) {
	writer.WriteU16(icmp.Checksum)
	writer.WriteU8(icmp.Code)
	writer.WriteU8(uint8(icmp.Type))
}

func WritePing(writer *ByteWriter, ping *Ping) {
	writer.WriteByte(ping.Data)
	writer.WriteByte(ping.Timestamp)
	writer.WriteU16(ping.Seq)
	writer.WriteU16(ping.ID)
}

func ReadPing(reader *ByteReader) *Ping {
	id := reader.ReadU16()
	seq := reader.ReadU16()
	timestamp := reader.ReadByte(8)
	data := reader.ReadLast()
	return &Ping{
		ID:        id,
		Seq:       seq,
		Timestamp: timestamp,
		Data:      data,
	}
}

type UnReach struct {
	Unused uint8
	Len    uint8
	Var    uint16
	Data   []byte // 存放尽可能多的原始数据包
}

func CheckICMP(hdr *ICMPHdr, reader *ByteReader) {
	// TODO Checksum计算
}

func ReadICMPHdr(reader *ByteReader) *ICMPHdr {
	type0 := reader.ReadU8()
	code := reader.ReadU8()
	checksum := reader.ReadU16()
	return &ICMPHdr{
		Type:     ICMPType(type0),
		Code:     code,
		Checksum: checksum,
	}
}
