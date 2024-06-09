/*
@author: sk
@date: 2024/6/7
*/
package main

import "fmt"

type UDPHdr struct {
	SrcPort, DstPort uint16
	Total            uint16
	Checksum         uint16
}

type UDPHandler func(param *Param, reader *ByteReader)

var (
	udpHandlers = make(map[uint16]UDPHandler)
)

func RegisterUDPHandler(port uint16, handler UDPHandler) {
	udpHandlers[port] = handler
}

func HandleUDP(param *Param, reader *ByteReader) {
	hdr := ReadUDPHdr(reader)
	param.SrcPort = hdr.SrcPort
	param.DstPort = hdr.DstPort
	CheckUDP(hdr, reader)
	if handler, ok := udpHandlers[hdr.DstPort]; ok {
		handler(param, reader)
	} else { // 应该返回端口不可达
		panic(fmt.Sprintf("invalid UDP port %d", hdr.DstPort))
	}
}

func SendUDP(param *Param, writer *ByteWriter, hdr *UDPHdr) {
	WriteUDPHdr(writer, hdr)
	fakeHdr := &FakeHdr{ // 伪头部信息
		SrcAddr:  param.DstAddr,
		DstAddr:  param.SrcAddr,
		Protocol: Ipv4UDP,
		Total:    writer.Len(),
	}
	WriteFakeHdr(writer, fakeHdr)
	hdr.Checksum = GetChecksum(writer.GetData())
	writer.Seek(8 + 12)      // 移出伪头部与 udp 头部
	WriteUDPHdr(writer, hdr) // 写入正确的 udp头
	hdr0 := &IPv4Hdr{
		Version:  4,
		HdrLen:   5,
		Len:      writer.Len() + 20,
		Ttl:      64,
		Protocol: Ipv4UDP,
		SrcAddr:  param.DstAddr,
		DstAddr:  param.SrcAddr,
	}
	SendIpv4(param, writer, hdr0)
}

func WriteUDPHdr(writer *ByteWriter, hdr *UDPHdr) {
	writer.WriteU16(hdr.Checksum)
	writer.WriteU16(hdr.Total)
	writer.WriteU16(hdr.DstPort)
	writer.WriteU16(hdr.SrcPort)
}

func CheckUDP(hdr *UDPHdr, reader *ByteReader) {
	// TODO 校验和计算
}

func ReadUDPHdr(reader *ByteReader) *UDPHdr {
	srcPort := reader.ReadU16()
	dstPort := reader.ReadU16()
	total := reader.ReadU16()
	checksum := reader.ReadU16()
	return &UDPHdr{
		SrcPort:  srcPort,
		DstPort:  dstPort,
		Total:    total,
		Checksum: checksum,
	}
}

func EchoUDPHandler(param *Param, reader *ByteReader) {
	writer := NewByteWriter()
	writer.WriteByte(reader.ReadLast())
	hdr := &UDPHdr{
		SrcPort: param.DstPort,
		DstPort: param.SrcPort,
		Total:   writer.Len() + 8,
	}
	SendUDP(param, writer, hdr)
}
