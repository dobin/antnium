
runserver: 
	go run cmd\server\server.go --listenaddr 127.0.0.1:8080

runclient: 
	go run cmd\client\client.go

runexecutor:
	go run cmd\executor\executor.go 

compile: server client executor
	
server:
	go build cmd\server\server.go 

client:
	go build cmd\client\client.go

executor:
	go build cmd\executor\executor.go 

test:
	go test .\...

prodclient:
	go build -ldflags="-s -w" cmd\client\client.go

clean:
	rm server.exe client.exe executor.exe
