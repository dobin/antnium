# Antnium 

```
Anti Tanium
```

There are two components: 
* client.exe: The actual trojan
* server.exe: C2 infrastructure 


## Quick How to use

Download and install go (and git).

Decide on a C2 IP or domain. We use `127.0.0.1:8080` here, as we start both client.exe and server.exe
on the same host (and OS). This is also the default, no need to change anything. 

Check your campaign in `model/campaign.go`: 
* `serverUrl = "http://127.0.0.1:8080"`

Access the WebUI by opening the following URL in the browser after starting server.exe:
```
http://localhost:8080/
```

### Quick Windows

Go: https://golang.org/doc/install

Compile client.exe and server.exe: 
```
> go get all
> .\makewin.bat server
> .\makewin.bat client
```

Start server.exe:
```
> .\server.exe --listenaddr localhost:8080
```

Start client.exe:
```
> .\client.exe
```


### Quick Linux

Go: `apt install golang`

Compile client and server.exe: 
```
$ go get all
$ make server
$ make client
```

Start server.exe:
```
$ ./server --listenaddr localhost:8080
```

Start client.exe:
```
$ ./client
```


## Campaign configuration

`campaign.go` connects a compiled client.exe with a specific server.exe, which forms a campaign. 
A campaign has individual encryption and authentication keys, which are shared between
server and client. 

The relevant parts of a campagin is depicted here:
```
type Campaign struct {
	ApiKey      string  // Key used to access client facing REST
	AdminApiKey string  // Key used to access admin facing REST
	EncKey      []byte  // Key used to encrypt packets between server/client
	ServerUrl   string  // URL of the server, as viewed from the clients
}
```

Note that `ServerUrl` is the URL used by the client for all interaction with the server. 
It is the public server URL, e.g. `http://totallynotmalware.ch`. The actual server.exe may
be behind a reverse proxy, and started with `server.exe --listenaddr 127.0.0.1:8080` (ServerUrl != listenaddr). 


## Details

```
$ git clone https://github.com/dobin/antnium
$ cd antnium
$ go get all
```

## Server

Tested on: 
* Works: Ubuntu 20.04 LTS, Go 1.13.8
* Wors: Windows 10, Go 1.16.6
* Compile fail: Ubuntu 16.04 LTS, Go 1.6.2

```
$ make server
$ ./server --listenaddr 0.0.0.0:8080
```

Put a reverse proxy before it, make sure it supports websockets!

Result is `server.exe`. Make sure to run it in the directory where you have or expect: 
* upload/
* static/
* db.*.json

as working directory.


## Client

Tested on: 
* Windows 10
* Ubuntu 20.04

Compile on windows:
```
> .\makewin.bat client
```

Deploy it on your target.


## Testing

```
go test ./...
go test ./server
```
