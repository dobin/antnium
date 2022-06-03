# Linux Makefile

# LDFLAGS = "-s -w"
# LDFLAGS = "-H windowsgui"
LDFLAGS=""

# For debugging
runserver: 
	go run cmd/server/server.go --listenaddr 127.0.0.1:8080

runclient: 
	go run cmd/client/client.go

runwingman:
	go run cmd/wingman/wingman.go 


# all
compile: server client wingman
	
server:
	go build -o server.elf cmd/server/server.go 

client:
	GOOS=linux GOARCH=amd64 go build -o client.elf cmd/client/client.go
	CGO_ENABLED=1 GOOS=windows GOARCH=amd64 CC=x86_64-w64-mingw32-gcc CGO_LDFLAGS="-L /usr/x86_64-w64-mingw32/lib/ -lpsapi" go build -o client.exe -ldflags $(LDFLAGS) cmd/client/client.go

wingman:
	CGO_ENABLED=1 GOOS=windows GOARCH=amd64 CC=x86_64-w64-mingw32-gcc CGO_LDFLAGS="-L /usr/x86_64-w64-mingw32/lib/ -lpsapi" go build -o wingman.exe -ldflags $(LDFLAGS) cmd/wingman/wingman.go 

deploy: compile
	# directory structure
	mkdir -p build/static build/upload

	cp server.elf build/
	cp client.elf client.exe build/static/
	cp wingman.exe build/static/
	cp -R webui/* build/webui/


# Utilities
test:
	go test ./...

clean:
	rm -f server.exe client.exe client.elf client.darwin wingman.exe
