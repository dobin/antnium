# Linux Makefile

# For debugging
runserver: 
	go run cmd/server/server.go --listenaddr 127.0.0.1:8080

runclient: 
	go run cmd/client/client.go

rundownstreamclient:
	go run cmd/downstreamclient/downstreamclient.go 


# all
compile: server client downstreamclient
	
server:
	go build cmd/server/server.go 

client:
	GOOS=linux GOARCH=amd64 go build -o client.elf cmd/client/client.go
	GOOS=windows GOARCH=amd64 go build -o client.exe cmd/client/client.go
	

downstreamclient:
	go build cmd/downstreamclient/downstreamclient.go 


# Deploy
deploy:
	go build -ldflags="-s -w" cmd/client/client.go


# Utilities
test:
	go test ./...


clean:
	rm server.exe client.exe downstreamclient.exe
