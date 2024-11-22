#!/usr/bin/env bash
set -eu -o pipefail # -x
_wd=$(pwd); _path=$(dirname $0 | xargs -i readlink -f {})

function show_help() {
    >&2 echo -e "Usage $(basename $0):\n     socks5_ssh.sh <remote_host> [127.0.0.1:1081]"
}

if [ $# -eq 0 ] || [ "$1" == "-h" ] || [ "$1" == "--help" ]; then
    show_help
    exit 0
fi

remote_host="$1"
address="${2:-127.0.0.1:1081}"

config=${config:-""}

[ ! -z "$(netstat -tulpn 2>/dev/null | grep -w "$address")" ] && {
    >&2 echo '!!!'" address is occupied: $address"
    exit 0
}

>&2 echo "[$(date +%FT%T%:z)] socks5 proxy: remote_host=$remote_host, address=$address, config=$config"

# autossh -f
# -p 22
# -i ~/.ssh/id_rsa
# -o "UserKnownHostsFile ~/.ssh/known_hosts"
if [[ ! -z "$config" ]]; then
    ssh -NC -F "$config" -D "$address" "$remote_host"
else
    ssh -NC -D \
      -o "ServerAliveInterval 5" \
      -o "ServerAliveCountMax 3" \
      -o "ExitOnForwardFailure yes" \
      "$address" "$remote_host"
fi

exit 0
cat /path/to/ssh.conf <<EOF
Host remote_host
	#ProxyJump another_host
	HostName 127.0.0.1
	User account
	Port 22
	IdentityFile ~/.ssh/id_rsa
	UserKnownHostsFile ~/.ssh/known_hosts
	Compression yes
	LogLevel INFO
	ServerAliveInterval 5
	ServerAliveCountMax 3
	ExitOnForwardFailure yes
EOF

ssh -NC -F /path/to/ssh.conf -D $address $remote_host

exit 0
proxy=socks5h://127.0.0.1:1081 # proxy=socks5h://username:password@127.0.0.1:1081

https_proxy=$proxy git pull
https_proxy=$proxy curl https://icanhazip.com
curl -x $proxy https://icanhazip.com

# neither firefox or chromium support socks5 with auth
chromium --disable-extensions --incognito --proxy-server="$proxy"

firefox -p proxy

https_proxy=$proxy remmina
