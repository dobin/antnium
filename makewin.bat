@echo off

IF "%1"=="runserver" (
    go run cmd\server\server.go --listenaddr 127.0.0.1:8080
) ELSE IF "%1"=="runclient" (
    go run cmd\client\client.go
) ELSE IF "%1"=="rundownstreamclient" (
    go run cmd\downstreamclient\downstreamclient.go 
) ELSE IF "%1"=="server" (
    go build -o server.exe cmd\server\server.go 
    SET GOOS=linux
    go build -o server.elf cmd\server\server.go 
) ELSE IF "%1"=="client" (
    REM -H windowsgui will disable the go window
    go build -o client.exe -ldflags "-H windowsgui" cmd\client\client.go
    SET GOOS=linux
    go build -o client.elf cmd\client\client.go
) ELSE IF "%1"=="downstreamclient" (
    go build -o downstreamclient.exe -ldflags "-H windowsgui" cmd\downstreamclient\downstreamclient.go 
    SET GOOS=linux
    go build -o downstreamclient.elf cmd\downstreamclient\downstreamclient.go 
) ELSE IF "%1"=="deploy" (
    .\makewin.bat client
    .\makewin.bat server
    .\makewin.bat downstreamclient
    mkdir build\upload
    mkdir build\static
    mkdir build\webui
    copy server.elf build\
    copy server.exe build\
    copy client.exe build\static\
    copy client.elf build\static\
    copy downstreamclient.exe build\static\
    copy downstreamclient.elf build\static\
    copy webui\* build\webui\
) ELSE IF "%1"=="coverage" (
    go test -coverprofile="coverage.out"
    go tool cover -html="coverage.out"
) ELSE (
    echo "Unknown: %1"
)