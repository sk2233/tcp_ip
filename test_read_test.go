/*
@author: sk
@date: 2024/6/8
*/
package main

import (
	"fmt"
	"os"
	"testing"
)

func TestRead(t *testing.T) {
	open, err := os.Open("/Users/bytedance/Documents/go/tcp_ip/res/test.html")
	fmt.Println(open, err)
}
