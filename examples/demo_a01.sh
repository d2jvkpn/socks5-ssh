#!/usr/bin/env bash
set -eu -o pipefail # -x
_wd=$(pwd); _path=$(dirname $0 | xargs -i readlink -f {})

# echo "Hello, world!"

proxy=socks5://127.0.0.1:1081
# proxy=socks5://username:password@127.0.0.1:1081

https_proxy=$proxy git pull

https_proxy=$proxy curl -4 https://icanhazip.com

# neither firefox or chromium support socks5 with auth
chromium --disable-extensions --incognito --proxy-server="$proxy"

firefox -p proxy
