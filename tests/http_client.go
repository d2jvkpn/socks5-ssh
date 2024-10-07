package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
)

func main() {
	var (
		err   error
		proxy string

		urlAddr   string
		urlProxy  *url.URL
		transport *http.Transport
		client    *http.Client
		request   *http.Request
		response  *http.Response
		body      []byte
	)

	flag.StringVar(&proxy, "proxy", "socks5://127.0.0.1:1080", "proxy address")
	flag.StringVar(&urlAddr, "urlAddr", "https://icanhazip.com", "request url")

	flag.Parse()

	defer func() {
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}
	}()

	if urlProxy, err = url.Parse(proxy); err != nil {
		return
	}

	transport = &http.Transport{
		Proxy:           http.ProxyURL(urlProxy),
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client = &http.Client{Transport: transport}

	if request, err = http.NewRequest("GET", urlAddr, nil); err != nil {
		return
	}
	// response, err = client.Get(urlAddr)
	if response, err = client.Do(request); err != nil {
		return
	}
	defer response.Body.Close()

	if body, err = ioutil.ReadAll(response.Body); err != nil {
		return
	}

	fmt.Printf("==> response:\n%s\n", body)
}
