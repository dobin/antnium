# Description

## Commands

* ping: special, initiated by client, and not logged. Handled by server. Contains host info. 
* exec: Execute a file
* fileUpload: upload a file 
* fileDownload: download a file 
* dir: directory content
* iOpen: interactive shell open
* iIssue: interactive shell issue data


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

This is intentional. The campaign is only protected against outsiders, not a motivated blue team. 

The admin API is protected by a separate AdminApi key, not found in the client. 

