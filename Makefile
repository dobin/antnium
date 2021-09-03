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
	go build -o server.elf cmd/server/server.go 

client:
	GOOS=linux GOARCH=amd64 go build -o client.elf cmd/client/client.go
	GOOS=windows GOARCH=amd64 go build -o client.exe -ldflags "-H windowsgui"  cmd/client/client.go
	#GOOS=darwin GOARCH=amd64 go build -o client.darwin cmd/client/client.go

downstreamclient:
	GOOS=windows GOARCH=amd64 go build -o downstreamclient.exe cmd/downstreamclient/downstreamclient.go 


deploy: compile
	# client
	GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o client.elf cmd/client/client.go
	GOOS=windows GOARCH=amd64 go build -ldflags="-s -w -H windowsgui" -o client.exe cmd/client/client.go
	# GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o client.darwin cmd/client/client.go
	GOOS=windows GOARCH=amd64 go build -ldflags="-s -w -H windowsgui" -o downstreamclient.exe cmd/downstreamclient/downstreamclient.go 

	# server
	GOOS=linux GOARCH=amd64 go build cmd/server/server.go 

	# directory structure
	mkdir -p build/static build/upload

	cp server.elf build/
	cp client.elf client.exe build/static/
	cp downstreamclient.exe build/static/
	cp -R webui/* build/webui/


# Utilities
test:
	go test ./...

clean:
	rm server.exe client.exe client.elf client.darwin downstreamclient.exe
