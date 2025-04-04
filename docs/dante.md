# Title
---
```meta
version: 0.1.0
authors: ["Jane Doe<jane.doe@noreply.local>"]
date: 1970-01-01
```


#### Ch01. 
1. links
- https://www.inet.no/dante/
- https://www.inet.no/dante/doc/
- https://www.inet.no/dante/doc/1.4.x/config/ipv6.html

2. install
```bash
sudo apt update
sudo apt install dante-server

apk update
apk add dante-server
```

3. config
```path=/etc/danted.conf
logoutput: /var/log/dante-server.log # syslog stdout stderr
user.privileged: root
user.unprivileged: nobody
user.libwrap: nobody

internal: eth0 port = 1080
external: tun0 # wg0

clientmethod: none
socksmethod: none

client pass {
    #from: 0.0.0.0/0 to: 0.0.0.0/0 # ipv4
    #from: ::/0 to: ::/0           # ipv6
    from: 0/0 to: 0/0
    log: connect disconnect error
}

socks pass {
    #from: 0.0.0.0/0 to: 0.0.0.0/0 # ipv4
    #from: ::/0 to: ::/0           # ipv6
    from: 0/0 to: 0/0
    log: connect disconnect error
}
```

4. services
```
danted -f configs/dante-server.conf -N 2
# sockd -f configs/dante-server.conf -N 2 # alpine

systemctl status danted
systemctl start/stop danted
systemctl enable/disable danted
```
