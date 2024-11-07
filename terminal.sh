#!/usr/bin/env bash
set -eu -o pipefail # -x
_wd=$(pwd); _path=$(dirname $0 | xargs -i readlink -f {})

cd ${_path}

container=$(yq .services.socks5-openvpn.container_name docker-compose.yaml)

docker exec -it $container bash
