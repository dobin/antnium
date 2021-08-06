
runserver: 
	go build cmd\server\server.go 
	./server.exe 

runclient: 
	go build cmd\client\client.go
	./client.exe

runexecutor:
	go build cmd\executor\executor.go 
	./executor.exe

compile: 
	go build cmd\client\client.go
	go build cmd\server\server.go 
	go build cmd\executor\executor.go 

test:
	go test .\...

prodclient:
	go build -ldflags="-s -w" cmd\client\client.go

clean:
	rm server.exe client.exe executor.exe
