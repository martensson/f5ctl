# f5ctl

*A REST API Proxy to quick and easy control clusters of F5 BIGIPs*

f5ctl is built in Go for easy deployment, concurrency and simplicity.

It serves a simple REST API that can be used to get status on nodes and to change their status. It will auto-detect the Bigip that is Active in a cluster.

### Installing from source

#### Dependencies

* Git
* Go 1.4+
* F5 LTM 11.4+ with REST API Enabled

#### Clone and Build locally:

``` sh
git clone https://github.com/martensson/f5ctl.git
cd f5ctl
go build
```

### Create a config.yml file and add all your bigips:

``` yaml
---
apiuser: "admin"
apipass: "admin"
f5:
    external:
        user: "admin"
        pass: "admin"
        ltm:
            - "ltmext01"
            - "ltmext02"
    internal:
        user: "admin"
        pass: "admin"
        ltm:
            - "ltmint01"
            - "ltmint02"
```

### Running f5ctl

``` sh
./f5ctl -p 5000 -f /path/to/config.yml
```
or just

``` sh
./f5ctl
```

### CURL Examples

#### Get info about all nodes

``` sh
curl -u admin:admin -i localhost:5000/v1/nodes/internal/
```

#### Get info about one node

``` sh
curl -u admin:admin -i localhost:5000/v1/nodes/internal/mynginxserver
```

#### Enable a node

``` sh
curl -u admin:admin -i localhost:5000/v1/nodes/internal/mynginxserver -X PUT -d '{"State":"enabled"}'
```

#### Force offline a node

``` sh
curl -u admin:admin -i localhost:5000/v1/nodes/internal/mynginxserver -X PUT -d '{"State":"forced-offline"}'
```
