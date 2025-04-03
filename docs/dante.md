# Title
---
```meta
version: 0.1.0
authors: ["Jane Doe<jane.doe@noreply.local>"]
date: 1970-01-01
```


#### Ch01. 
1. install
```bash
sudo apt update
sudo apt install dante-server

apk update
apk add dante-server
```

2. config
```path=/etc/danted.conf
logoutput: /var/log/danted-server.log syslog
user.privileged: root
user.unprivileged: nobody
internal: 127.0.0.1 port = 1080
external: wg0

socksmethod: none
socks pass {
    from: 0.0.0.0/0 to: 0.0.0.0/0
    #log: connect disconnect error
}
```
