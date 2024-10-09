package proxy

import (
	"context"
	"crypto/tls"
	// "fmt"
	"net"
	"net/http"
	"net/url"
	"time"
)

func NewHttpClient(proxyAddr string, tlsInsecureSkipVerify bool) (client *http.Client, err error) {
	var (
		transport *http.Transport
		urlProxy  *url.URL
	)

	if urlProxy, err = url.Parse(proxyAddr); err != nil {
		return nil, err
	}

	transport = &http.Transport{
		Proxy: http.ProxyURL(urlProxy),
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: tlsInsecureSkipVerify,
		},
		DialContext: (&net.Dialer{
			Timeout: 2 * time.Second,
		}).DialContext,
		ResponseHeaderTimeout: 2 * time.Second,
	}

	client = &http.Client{
		Transport: transport,
		Timeout:   3 * time.Second,
	}

	return client, nil
}

func HttpTest(client *http.Client, method, addr string) (statudCode int, err error) {
	var (
		ctx      context.Context
		cancel   func()
		request  *http.Request
		response *http.Response
	)

	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if request, err = http.NewRequestWithContext(ctx, method, addr, nil); err != nil {
		return 0, err
	}

	// response, err = client.Get(urlAddr)
	if response, err = client.Do(request); err != nil {
		return 0, err
	}
	defer response.Body.Close()

	//var bts []byte
	//bts, err = ioutil.ReadAll(response.Body)
	//fmt.Printf("==> response_body:\n%s\n", bts)

	return response.StatusCode, nil
}
