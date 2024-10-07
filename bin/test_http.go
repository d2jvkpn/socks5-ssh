package bin

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"time"
)

func TestProxy(args []string) {
	var (
		err     error
		proxy   string
		urlAddr string
		fSet    *flag.FlagSet

		urlProxy *url.URL
		ctx      context.Context
		cancel   func()

		client  *http.Client
		request *http.Request

		response *http.Response
		body     []byte
	)

	fSet = flag.NewFlagSet("proxy", flag.ContinueOnError) // flag.ExitOnError

	fSet.StringVar(&proxy, "proxy", "socks5h://127.0.0.1:1080", "proxy address")
	fSet.StringVar(&urlAddr, "urlAddr", "https://icanhazip.com", "request url")

	if err = fSet.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "ssh exit: %v\n", err)
		os.Exit(1)
		return
	}

	defer func() {
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}
	}()

	if urlProxy, err = url.Parse(proxy); err != nil {
		return
	}
	client = newClient(urlProxy)

	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if request, err = http.NewRequestWithContext(ctx, "GET", urlAddr, nil); err != nil {
		return
	}

	// response, err = client.Get(urlAddr)
	if response, err = client.Do(request); err != nil {
		return
	}
	defer response.Body.Close()

	body, err = ioutil.ReadAll(response.Body)
	fmt.Printf("==> response: status_code=%d, body=\n%s\n", response.StatusCode, body)
}

func newClient(urlProxy *url.URL) (client *http.Client) {
	var transport *http.Transport

	transport = &http.Transport{
		Proxy:           http.ProxyURL(urlProxy),
		TLSClientConfig: &tls.Config{InsecureSkipVerify: false},
		DialContext: (&net.Dialer{
			Timeout: 2 * time.Second,
		}).DialContext,
		ResponseHeaderTimeout: 2 * time.Second,
	}

	client = &http.Client{
		Transport: transport,
		Timeout:   5 * time.Second,
	}

	return client
}
