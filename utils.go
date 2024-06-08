/*
@author: sk
@date: 2024/6/4
*/
package main

import "encoding/binary"

func HandleErr(err error) {
	if err != nil {
		panic(err)
	}
}

func GetChecksum(data []byte) uint16 {
	res := 0
	for i := 1; i < len(data); i += 2 {
		res += int(binary.BigEndian.Uint16(data[i-1 : i+1]))
	}
	if len(data)%2 == 1 {
		res += int(data[len(data)-1]) << 8
	}
	for (res >> 16) > 0 { // 若是大于 u16就进行削减
		res = (res >> 16) + (res & 0xFFFF)
	}
	return ^uint16(res)
}
