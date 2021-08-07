# Linux Makefile

# For debugging
runserver: 
	go run cmd/server/server.go --listenaddr 127.0.0.1:8080

runclient: 
	go run cmd/client/client.go

runexecutor:
	go run cmd/executor/executor.go 


# all
compile: server client executor
	
server:
	go build cmd/server/server.go 

client:
	GOOS=linux GOARCH=amd64 go build -o client.elf cmd/client/client.go
	GOOS=windows GOARCH=amd64 go build -o client.exe cmd/client/client.go
	

executor:
	go build cmd/executor/executor.go 


# Deploy
deploy:
	go build -ldflags="-s -w" cmd/client/client.go


# Utilities
test:
	go test ./...


clean:
	rm server.exe client.exe executor.exe
