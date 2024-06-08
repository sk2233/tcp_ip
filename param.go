/*
@author: sk
@date: 2024/6/5
*/
package main

import "github.com/songgao/water"

type Param struct { // 随tcp_ip协议栈向下传递的参数，会不断补充
	Inst    *water.Interface
	SrcAddr []byte // 4
	DstAddr []byte // 4
	SrcPort uint16
	DstPort uint16
	TCPHdr  *TCPHdr
}
