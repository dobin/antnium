
server: 
	go build .\cmd\server\server.go 
	./server.exe 


client: 
	go build .\cmd\client\client.go
	./client.exe

executor:
	go build .\cmd\executor\executor.go 
	./executor.exe

clean:
	rm server.exe client.exe executor.exe
