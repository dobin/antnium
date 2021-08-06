# Antnium 

```
Anti Tanium
```

There are two components: 
* Client: The actual trojan
* Server: C2 infrastructure 


## Quick How to use

Decide on a C2 IP or domain. We use `127.0.0.1:8080` here. 

Configure your campaign in `model/campaign.go`, minimum: 
* `serverUrl = "http://127.0.0.1:8080"`

Compile client and server: 
```
go get all
make compile
```

Start server:
```
server.exe --listenaddr 0.0.0.0:8080
```

Access the WebUI by opening the following URL in the browser:
```
http://localhost:8080/
```

Start client:
```
client.exe
```




Start a client on your target. 

### Notes on deployment

`campaign.go` connects a client with a specific server, which forms a campaign. 
A campaign has individual encryption and authentication keys, which are shared between
server and client. 

* Replace `127.0.0.1:8080` with your domain, e.g. `totallynotmalware.ch`.
* Put server behind a reverse proxy.


## Install 

```
$ git clone https://github.com/dobin/antnium
$ cd antnium
$ go get all
$ make compile
```

### Server

Tested on: 
* Works: Ubuntu 20.04 LTS, Go 1.13.8
* Wors: Windows 10, Go 1.16.6
* Doesnt work: Ubuntu 16.04 LTS, Go 1.6.2

```
$ make server
$ ./server --listenaddr 0.0.0.0:8080
```

Put a reverse proxy before it, make sure it supports websockets.

Result is `server.exe`. Make sure to run it in the directory where you have or expect: 
* upload/
* static/
* db.*.json

### Client

Configure your campaign in `model/campaign.go`: 
* `serverUrl = "http://totallynotmalware.ch"`

Compile on windows:
```
make prodclient
```

Deploy it on your target.


### Commands

* exec: Execute a file
* fileupload: upload a file 
* filedownload: download a file 
* dir: directory content

For a complete list, see `doc/protocol.md`.

## Server

* Runs on a specific port
* uploads files from client via REST to `./upload/`
* serves directory `./static/`

### DB

The server stores its data in the files: 
* db.packets.json
* db.clients.json
regularly. It will load it on start automatically. 

Use: 
* `server.exe --dbReadOnly` to only read but not update
* `server.exe --dbWriteOnly` to only write but not read



## Security 

The client and server share a static encryption key, and a API key. 

If the blue team manages to extract the API key from a HTTP proxy or client binary, they
gain access to the server API, which enables them to:
* flood the server with fake clients 
* observe public executed commands easily 

If the blue team manages to extract the encryption key from a client binary, they can: 
* decrypt all past communications of all client instances (if they have proxy log)
* Issue new commands to existing clients (if they can perform HTTP MITM on proxy)

This is intentional. The campgain is only protected against outsiders, not a motivated blue team. 

The admin API is protected by a separate AdminApi key, not found in the client. 

## Packet Types

From client to server:

* Ping
  * sent from client to server upon start
  * only packet not initiated by server
  * only packettype the server knows about
  * not logged or broadcasted (anti-spam)
  * contains host info
  * data available through /admin/clients API

From server to client: 
* exec
* iOpen
* iIssue
* fileUpload
* fileDownload


## Testing

```
go test ./...
go test ./server
```