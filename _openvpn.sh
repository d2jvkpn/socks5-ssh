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

exit
cat > configs/supervisord.conf <<EOF
# path: /etc/supervisord.conf

[unix_http_server]
file=/var/run/supervisord.sock

[supervisorctl]
serverurl=unix:///var/run/supervisord.sock
# port=127.0.0.1:9001

[rpcinterface:supervisor]
supervisor.rpcinterface_factory = supervisor.rpcinterface:make_main_rpcinterface

[supervisord]
nodaemon=true

[include]
# files = /etc/supervisor.d/*.ini
# relative paths to configs/supervisord.conf
files = supervisor.ini
EOF

exit
cat configs/supervisor.ini <<EOF
# path: /etc/supervisor.d/*.ini
[program:http_server]
command=python3 -m http.server 8080
autostart=true
autorestart=true
stdout_logfile=/var/log/http_server.stdout.log
stderr_logfile=/var/log/http_server.stderr.log

[program:openvpn]
user=root
command=openvpn --config configs/openvpn.ovpn --auth-user-pass configs/openvpn.auth
autostart=true
autorestart=true
stdout_logfile=logs/openvpn.log
stderr_logfile=logs/openvpn.error

[program:proxy]
command=target/main server --addr=:1090
autostart=true
autorestart=true
stdout_logfile=logs/proxy_server.log
stderr_logfile=logs/proxy_server.error
EOF
