/*
@author: sk
@date: 2024/6/5
*/
package main

import "fmt"

type LoopHdr struct {
	Type LoopType
}

func HandleLoop(reader *ByteReader) {
	hdr := ReadLoopHdr(reader)
	fmt.Println(hdr)
}

func ReadLoopHdr(reader *ByteReader) *LoopHdr {
	type0 := reader.ReadU32()
	return &LoopHdr{
		Type: LoopType(type0),
	}
}
