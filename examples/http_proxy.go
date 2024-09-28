// https://stackoverflow.com/questions/63656117/cannot-use-socks5-proxy-in-golang-read-connection-reset-by-peer/63661992#63661992
package main

import (
	"context"
	"fmt"
	"golang.org/x/net/proxy"
	"io/ioutil"
	"net"
	"net/http"
	"runtime"
	"time"
)

func main() {
	proxyUrl := "127.0.0.1:1080"
	dialer, err := proxy.SOCKS5(
		"tcp", proxyUrl, &proxy.Auth{User: "hello", Password: "world"}, proxy.Direct,
	)

	dialContext := func(ctx context.Context, network, address string) (net.Conn, error) {
		return dialer.Dial(network, address)
	}

	transport := &http.Transport{
		// Proxy:                 http.ProxyFromEnvironment,
		DialContext:           dialContext,
		DisableKeepAlives:     true,
		MaxIdleConns:          10,
		IdleConnTimeout:       10 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		MaxIdleConnsPerHost:   runtime.GOMAXPROCS(0) + 1,
	}
	cl := &http.Client{Transport: transport}

	// resp, err := cl.Get("https://wtfismyip.com/json")
	resp, err := cl.Get("https://www.google.com")
	if err != nil {
		// TODO handle me
		panic(err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	// TODO work with the response
	if err != nil {
		fmt.Println("body read failed")
	}
	fmt.Println(string(body))
}
