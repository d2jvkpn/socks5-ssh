#!/usr/bin/env bash
set -eu -o pipefail # -x
_wd=$(pwd); _path=$(dirname $0 | xargs -i readlink -f {})

exit
docker exec -it socks5-openvpn supervisorctl status
# supervisorctl reread
# supervisorctl update

docker exec socks5-openvpn curl https://icanhazip.com

curl -x sock5h://127.0.0.1:1090 https://icanhazip.com

# docker exec -it socks5-openvpn bash

# docker exec -it socks5-openvpn ssh -F configs/ssh.conf remote_host
