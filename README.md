# Antnium 

```
Anti Tanium
```

There are two components: 
* Client: The actual trojan
* Server: C2 infrastructure 


## How to use

Configure your campaign in: 
* model/campaign.go 

Compile client, server and executor: 
* make compile

Which produces: 
* client.exe
* server.exe
* executor.exe


Deploy server on the URL you defined in the campaign. Start a client somewhere. 

Default server address is `127.0.0.1:4444`. 

## Client 

Commands: 
* exec: Execute a file
* fileupload: upload a file 
* filedownload: download a file 
* dir: directory content

For a complete list, see `doc/protocol.md`.

## Server

* Runs on a specific port
* uploads files from client via REST to `./upload/`
* serves directory `./static/`

## DB

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