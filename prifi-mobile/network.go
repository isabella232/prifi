package prifimobile

import (
	"github.com/parnurzeal/gorequest"
	"strconv"
	"time"
)

// Used for latency test
type HTTPRequestResult struct {
	Latency    int64
	StatusCode int
	BodySize   int
}

// Creates an HTTPRequestResult
func NewHTTPRequestResult() *HTTPRequestResult {
	return &HTTPRequestResult{0, 0, 0}
}

/*
 * Request google home page through PriFi
 *
 * It is a method instead of a function due to the type restriction of gomobile.
 */
func (result *HTTPRequestResult) RetrieveHTTPResponseThroughPrifi(targetURLString string, timeout int, throughPrifi bool) error {
	// Get the localhost PriFi server port
	prifiPort, err := GetPrifiPort()
	if err != nil {
		return err
	}

	// Construct the proxy host address
	proxyURL := "socks5://127.0.0.1:" + strconv.Itoa(prifiPort)

	// Construct a request object with proxy and timeout value
	var request *gorequest.SuperAgent
	if throughPrifi {
		request = gorequest.New().Proxy(proxyURL).Timeout(time.Duration(timeout) * time.Second)
	} else {
		request = gorequest.New().Timeout(time.Duration(timeout) * time.Second)
	}

	// Used for latency test
	start := time.Now()
	resp, bodyBytes, errs := request.Get(targetURLString).EndBytes()
	elapsed := time.Since(start)

	if len(errs) > 0 {
		return errs[0]
	}

	result.Latency = elapsed.Nanoseconds()
	result.StatusCode = resp.StatusCode
	result.BodySize = len(bodyBytes)

	return nil
}
