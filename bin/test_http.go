package bin

import (
	"flag"
	"fmt"
	//"io/ioutil"
	"net/http"
	"os"

	"github.com/d2jvkpn/socks5-proxy/pkg/proxy"
)

func TestProxy(args []string) {
	var (
		err                   error
		proxyAddr             string
		urlAddr               string
		tlsInsecureSkipVerify bool
		fSet                  *flag.FlagSet

		client     *http.Client
		statusCode int
	)

	fSet = flag.NewFlagSet("proxy", flag.ContinueOnError) // flag.ExitOnError

	fSet.StringVar(&proxyAddr, "proxy", "socks5h://127.0.0.1:1080", "proxy address")
	fSet.StringVar(&urlAddr, "url", "https://icanhazip.com", "request url")

	fSet.BoolVar(
		&tlsInsecureSkipVerify,
		"tlsInsecureSkipVerify",
		false,
		"tls insecure skip verify",
	)

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

	if client, err = proxy.NewHttpClient(proxyAddr, tlsInsecureSkipVerify); err != nil {
		return
	}

	if statusCode, err = proxy.HttpTest(client, "GET", urlAddr); err != nil {
		return
	}
	fmt.Printf("==> status_code: %d\n", statusCode)
}
