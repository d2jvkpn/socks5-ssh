# socks5-proxy
---
*socks5 proxying through ssh tunnel and vpn*

#### C01. Solved Problems
1. Creating a SOCKS5 Proxy Server via SSH Tunnel with Optional Authentication
- language: Golang
- packages:
  - github.com/armon/go-socks5
  - golang.org/x/crypto/ssh

2. DNS Resolver (socksh://)
- uses the SSH remote command(crypto/ssh.Client.NewSession): dig +short <hostname>

3. Containers
- docker, docker-compose
- make

4. Autoheal
- configure healthcheck in compose.yaml
- use the autoheal project(https://github.com/willfarrell/docker-autoheal)

#### C02. Usage
1. configuration(configs/local.yaml)
```yaml
ssh:
  ssh_host: remote_host
  ssh_port: 22
  ssh_user: account
  ssh_password: password
  ssh_private_key: /home/account/.ssh/id_rsa
  ssh_known_hosts: /home/account/.ssh/known_hosts
  socks5_user: hello
  socks5_password: world

noauth:
  ssh_host: remote_host
  ssh_port: 22
  ssh_user: account
  ssh_password: password
  ssh_private_key: /home/account/.ssh/id_rsa
  ssh_known_hosts: /home/account/.ssh/known_hosts
```

2. compile
```bash
make build
```

3. run
```bash
./target/main ssh -config=configs/local.yaml -addr=127.0.0.1:1081
```

4. release
```bash
make release
```

5. deployment(docker-compose)
- build image socks5-proxy:dev:
```bash
make image-dev
```
- create compose.yaml
see deploy/compose.sh and deploy/compose.template.yaml

#### C03. Applications
1. commandlines with socks5 proxying
```bash
# proxy=socks5://hello:world@127.0.0.1:1081
proxy=socks5h://hello:world@127.0.0.1:1081

https_proxy=$proxy git pull
https_proxy=$proxy git push

https_proxy=$proxy curl -4 https://icanhazip.com
curl -x "$proxy" https://icanhazip.com
```

2. web browser with sock5 proxying
(**Neither Firefox nor Chromium supports SOCKS5 with authentication**)
```bash
proxy=socks5h://127.0.0.1:1081

chromium --disable-extensions --incognito --proxy-server="$proxy"

# mannual config proxy in settings of firefox
firefox -p proxy
```

#### C04. Run an openvpn client in container and expose a sock5 proxy
1. config and debug
- _openvpn.sh

2. deployment
- deploy/supervisord.compose.yaml

3. openvpn server in container
- https://github.com/d2jvkpn/playground/tree/master/container/openvpn


#### C05. TODO
1. block target ips and domains
2. ~~socks5 through ssh~~
3. ~~vpn~~
4. wireguard: apk add wireguard-tools, --cap-add=NET_ADMIN, --cap-add=SYS_MODULE
3. stunnel
