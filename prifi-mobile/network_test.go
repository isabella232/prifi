package prifimobile

import (
	"fmt"
	"testing"
)

func TestMakeHttpRequestThroughPrifi(t *testing.T) {
	r := NewHTTPRequestResult()
	fmt.Println(r)

	e := r.RetrieveHTTPResponseThroughPrifi("http://128.178.151.111", 5, false)
	if e != nil {
		fmt.Println(e)
	}
	fmt.Println(r)
}
