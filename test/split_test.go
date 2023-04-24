package test

import (
	"fmt"
	"strings"
	"testing"
)

/**
 * @Author Charles
 * @Date 9:19 PM 10/9/2022
 **/

const defaultBasePath = "/_geecache/"

func TestSplit(t *testing.T) {
	p := "/_geecache/scores/abc/_gee"
	parts := strings.SplitN(p[len(defaultBasePath):], "/", 2)
	fmt.Println(parts, len(defaultBasePath))
	if len(parts) != 2 {
		fmt.Println("nooooooo")
	}
}
