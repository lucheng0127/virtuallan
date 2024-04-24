# virtuallan
## Description
virtuallan is a l2vpn. It can setup an virtual Ethernet LAN network in WAN.

## Features
* Basic auth for vpn endpoint
* AES encrypt for ethernet traffic
* Ethernet traffic in udp

## How it work
![architecture](./docs/statics/architecture.png)
* server create a linux bridge for each virtual ethernet network
* server create a tap interface for each authed endpoint
* client create a tap interface
* encrypt ethernet traffic that on tap interface and send to udp conn
* receive udp stream from conn and decrypt then send to tap interface

An udp connection just like a cable connect dc and ep taps. And the taps became to a pair linux veth peer, connected to a linux bridge.

## Build

```
➜  virtuallan git:(master) ✗ make
go generate pkg/cipher/cipher.go
go build -o virtuallan main.go
```

## Getting started

**Server**
```
➜  virtuallan git:(master) ✗ ./virtuallan server -h
NAME:
   virtuallan server - run virtuallan server

USAGE:
   virtuallan server [command options] [arguments...]

OPTIONS:
   --config-dir value, -d value  config directory to launch virtuallan server, conf.yaml as config file, users as user storage
   --help, -h                    show help
```

config dir files:
* config.yaml: server config file
* users: user database csv format \<username>,\<user passwd base64 encode>

**Endpoint**
```
➜  virtuallan git:(master) ✗ ./virtuallan client -h
NAME:
   virtuallan client - connect to virtuallan server

USAGE:
   virtuallan client [command options] [arguments...]

OPTIONS:
   --target value, -t value  socket virtuallan server listened on
   --addr value, -a value    ipv4 address of current endpoint
   --user value, -u value    username of virtuallan endpoint
   --passwd value, -p value  password of virtuallan endpoint user
   --help, -h                show help
```

If not set -u and -p flags, you need to input user name and passwd in console