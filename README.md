Docker-dns
==========

Very simple resolver for docker, a lot to improve until I can call it a "Discovery" system.

Installation
============

```
go get -v github.com/rsampaio/docker-dns/...
```

Usage
=====

Run

```
$GOPATH/bin/docker-dns -i docker0
```

Start a container and get the ip from *.docker domain

```
# docker run -name mycontainer ubuntu /bin/bash

# host mycontainer.docker 172.17.42.1
Using domain server:
Name: 172.17.42.1
Address: 172.17.42.1#53
Aliases: 

mycontainer.docker has address 172.17.0.3
```

TODO
====

* SRV entries for each container port
* Dynamicaly remove entries based on docker events