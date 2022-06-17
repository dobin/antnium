# Antnium 

```
Anti Tanium
```

A C2 framework and RAT written in Go. 

There are two components: 
* client.exe: The actual RAT / beacon / agent / implant
* server.exe: C2 server

## Features

* HTTP/S and Websocket communication channel
* Proxy support (manual, windows, authenticated and kerberos)
* Command execution
  * Direct LOLbins
	* Copy file first
	* Process hollowing
  * Interactive cmd.exe/Powershell shell
  * Remote managed and unmanaged code
	* Using donut
	* PE to shellcode
	* Encrypted
	* AMSI bypass
* EDR bypass with Reflexxion (ntdll.dll restore)
* Encrypted communication
* Malleable C2
* File upload / download
* File browser


## Quick How to use

Download and install go (and git).

We use `127.0.0.1:8080` as C2 domain here (localhost as we start both client.exe and server.exe
on the same host). This is also the default, no need to change anything. 

Check campaign in `campaign/campaign.go`: 
* `serverUrl = "http://127.0.0.1:8080"`

Build it on windows: 
```
> .\makewin.bat deploy
```

Build it on linux: 
```
$ make deploy
```

Start server, and client: 
```
cd build\
.\server.exe
.\static\client.exe
```

Access the WebUI by opening the following URL in the browser after starting server.exe:
```
http://localhost:8080/webui/
```

## Directories

### `static/`: Public directory for tools

Put files there you want to download on other machines. Like `client.exe`, `wingman.exe`. 
And your tools, like `mimikatz.exe`, or `seatbelt.exe`. But use more inconspicuous file names. 

The files are also available via the `/secure` API requested with encrypted filenames, and encrypted+base64 encoded file as response.

dotNet files can be execute by using `remote` execution option (accessed via `/secure`).

### `upload/`: Private directory for data exfiltration 

File uploads from the client will be stored there. 



## Detailed build instructions

Go install: 
* Windows: https://golang.org/doc/install
* Linux: `apt install golang gcc-mingw-w64`

Compile client.exe and server.exe: 
```
> .\makewin.bat deploy
```

This will create: 
* /build/server.exe
* /build/server.elf
* /build/static/client.exe
* /build/static/client.elf
* /build/static/wingman.exe
* /build/upload/
* /build/webui/

Start server.exe:
```
> cd build
> .\server.exe

Antnium 0.1
Loaded 0 packets from db.packets.json
Loaded 0 clients from db.clients.json 
Periodic DB dump enabled
Starting webserver on 127.0.0.1:8080  
```

Start client.exe:
```
> .\build\static\client.exe

Antnium 0.1
time="2021-09-02T21:48:16+02:00" level=info msg="UpstreamHttp: Use WS"
time="2021-09-02T21:48:16+02:00" level=info msg="Connecting to WS succeeded"
time="2021-09-02T21:48:16+02:00" level=info msg=Send 1_computerId=c4oil02sdke2sp3nfngg 2_packetId=0 3_downstreamId=client 4_packetType=ping 5_arguments="map[]" 6_response=...
time="2021-09-02T21:48:16+02:00" level=info msg=Send 1_computerId=c4oil02sdke2sp3nfngg 2_packetId=0 3_downstreamId=client 4_packetType=ping 5_arguments="map[]" 6_response=...
```

## Notes on Campaign configuration

`pkg/campaign/campaign.go` connects a compiled client.exe with a specific server.exe, which forms a campaign. 
A campaign has individual encryption- and authentication keys, which are shared between
server and client. 

```
type Campaign struct {
	ApiKey      string  // Key used to access client facing REST
	EncKey      []byte  // Key used to encrypt packets between server/client

	ServerUrl   string  // URL of the server, as viewed from the clients
}
```

And admin UI / operator key in `pkt/server/config.go`:
```
type Config struct {
	AdminApiKey string
}```

Note that `ServerUrl` is the URL used by the client for all interaction with the server. 
It is the public server URL, e.g. `http://totallynotmalware.ch`. The actual server.exe may
be behind a reverse proxy, and started with `server.exe --listenaddr 127.0.0.1:8080` (so `ServerUrl` is not necessarily equal `listenaddr`). 

## Notes on server access

When first connecting to the server, you need to access and configure the UI first. 

The Angular UI files are publicly accessible. Lets assume `ServerUrl="http://localhost:8080"` and `listenaddr=0.0.0.0:8080`. You can either: 
* use the integrated antniumui, available as `http://localhost:8080/webui` on your browser
* or `ng serve` from antniumui directory, and then open `http://localhost:4200` on your browser

When connecting to the UI in the browser, you need first to configure the server IP and its password:
* AdminApiKey (default: "Secret-AdminApi-Key", like in config default)
* ServerIP (default: "http://localhost:8080")
* User (optional, can be chosen randomly)


## Client

Tested on: 
* Windows 10
* Ubuntu 20.04 LTS

Compile on windows:
```
> .\makewin.bat client
```

Deploy it on your target.


## Server

Tested on: 
* Works: Ubuntu 20.04 LTS, Go 1.13.8
* Works: Windows 10, Go 1.16.6
* Compile FAIL: Ubuntu 16.04 LTS, Go 1.6.2

On Linux:
```
$ make server
$ mkdir -p static upload
$ ./server --listenaddr 0.0.0.0:8080
```

Result is `server.exe`. Make sure to run it in the directory where you have or expect: 
* upload/
* static/
* db.*.json
as working directory.

It will start a REST server on that port, providing: 
* `/`: REST for the clients
* `/ws`: Websocket for the clients
* `/admin`: REST for admin interface (add packet, get clients)
* `/adminws`: Websocket for admin interface (push packets)
* `/webui`: HTML files for admin interface (Angular source and html, accesses REST and websocket)

Put a reverse proxy before it (make sure it supports websockets!) or forward ports.


## Options

For manual proxy, use full HTTP url:
```
client.exe -proxy http://proxy:8080
```

or via environment variables:
```
export PROXY http://localhost:8080
./client
```


## Wingman

Wingman is basically the Client, but without direct connection to the C2. 
It can connect to an existing client on localhost:50000 (make sure its started, if Campaign.AutoStartDownstreams is false)

Connects to localhost port 50000:
```
wingman.exe
```

Or use rundll32.exe to load the dll (the 64 bit rundll32 version in system32, not the 32 bit version in C:\Windows\SysWOW64\rundll32.exe):
```
C:\Windows\System32\rundll32.exe .\wingman.dll,Start
```

It will appear as downstream `net#0`.


## Testing

```
go test ./...
```

There may still be race conditions. If it fails once, just execute it again. 
