/*
@author: sk
@date: 2024/6/4
*/
package main

import "fmt"

type EtherHdr struct {
	DstMac []byte // 6
	SrcMac []byte // 6
	Type   EtherType
}

func HandleEther(reader *ByteReader) {
	hdr := ReadEtherHdr(reader)
	fmt.Println("Ether:", hdr.DstMac, hdr.SrcMac, hdr.Type)
}

func ReadEtherHdr(reader *ByteReader) *EtherHdr {
	dstMac := reader.ReadByte(6)
	srcMac := reader.ReadByte(6)
	type0 := reader.ReadU16()
	return &EtherHdr{
		DstMac: dstMac,
		SrcMac: srcMac,
		Type:   EtherType(type0),
	}
}
