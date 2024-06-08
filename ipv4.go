/*
@author: sk
@date: 2024/6/5
*/
package main

import "fmt"

type IPv4Hdr struct { // 不包含选项20byte
	Version  uint8  // 4
	HdrLen   uint8  // 4 头的整体大小，因为 ip包头是不定长的 要 *4 才是头长度
	Unused   uint8  // 表示服务质量，没有用
	Len      uint16 // 总长度
	ID       uint16 // 自增id 用于分段
	Flags    uint8  // 3 是否允许分片等标记
	Offset   uint16 // 13 分段偏移
	Ttl      uint8
	Protocol Ipv4Protocol // 上层协议类型
	Checksum uint16
	SrcAddr  []byte // 4
	DstAddr  []byte // 4
	// 选项数据不管
}

type FakeHdr struct { // tcp udp 计算校验和需要这个  12 byte
	SrcAddr  []byte // 4
	DstAddr  []byte // 4
	Unused   uint8
	Protocol Ipv4Protocol
	Total    uint16
}

func WriteFakeHdr(writer *ByteWriter, hdr *FakeHdr) {
	writer.WriteU16(hdr.Total)
	writer.WriteU8(uint8(hdr.Protocol))
	writer.WriteU8(hdr.Unused)
	writer.WriteByte(hdr.DstAddr)
	writer.WriteByte(hdr.SrcAddr)
}

func HandleIpv4(param *Param, reader *ByteReader) {
	hdr := ReadIpv4Hdr(reader)
	// 移出全部的头数据
	reader.ReadByte(int(hdr.HdrLen*4 - 20))
	CheckIpv4(hdr, reader)
	param.SrcAddr = hdr.SrcAddr
	param.DstAddr = hdr.DstAddr
	switch hdr.Protocol {
	case Ipv4ICMP:
		HandleICMP(param, reader)
	case Ipv4TCP:
		HandleTCP(param, reader)
	case Ipv4UDP:
		HandleUDP(param, reader)
	default:
		panic(fmt.Sprintf("unknown protocol %d", hdr.Protocol))
	}
}

func SendIpv4(param *Param, writer *ByteWriter, hdr *IPv4Hdr) {
	WriteIpv4Hdr(writer, hdr)
	hdr.Checksum = GetChecksum(writer.GetData()[:20]) // 只计算头部分的
	writer.Seek(20)
	WriteIpv4Hdr(writer, hdr)
	_, err := param.Inst.Write(writer.GetData())
	HandleErr(err)
}

func WriteIpv4Hdr(writer *ByteWriter, hdr *IPv4Hdr) {
	writer.WriteByte(hdr.DstAddr)
	writer.WriteByte(hdr.SrcAddr)
	writer.WriteU16(hdr.Checksum)
	writer.WriteU8(uint8(hdr.Protocol))
	writer.WriteU8(hdr.Ttl)
	writer.WriteU16((hdr.Offset << 3) | uint16(hdr.Flags&0x7))
	writer.WriteU16(hdr.ID)
	writer.WriteU16(hdr.Len)
	writer.WriteU8(hdr.Unused)
	writer.WriteU8((hdr.Version << 4) | (hdr.HdrLen & 0xF))
}

func CheckIpv4(hdr *IPv4Hdr, reader *ByteReader) {
	if hdr.Version != 4 {
		panic("invalid IPv4 version")
	}
	// TODO checksum计算
}

func ReadIpv4Hdr(reader *ByteReader) *IPv4Hdr {
	u8 := reader.ReadU8()
	unused := reader.ReadU8()
	len0 := reader.ReadU16()
	id := reader.ReadU16()
	u16 := reader.ReadU16()
	ttl := reader.ReadU8()
	protocol := reader.ReadU8()
	checksum := reader.ReadU16()
	srcAddr := reader.ReadByte(4)
	dstAddr := reader.ReadByte(4)
	return &IPv4Hdr{
		Version:  (u8 >> 4) & 0xF, // 注意顺序
		HdrLen:   u8 & 0xF,
		Unused:   unused,
		Len:      len0,
		ID:       id,
		Flags:    uint8(u16 & 0x7),
		Offset:   u16 >> 3,
		Ttl:      ttl,
		Protocol: Ipv4Protocol(protocol),
		Checksum: checksum,
		SrcAddr:  srcAddr,
		DstAddr:  dstAddr,
	}
}
