@echo off

IF "%1"=="runserver" (
    go run cmd\server\server.go --listenaddr 127.0.0.1:8080
) ELSE IF "%1"=="runclient" (
    go run cmd\client\client.go
) ELSE IF "%1"=="runwingman" (
    go run cmd\wingman\wingman.go 
) ELSE IF "%1"=="server" (
    go build -o server.exe cmd\server\server.go 
    SET GOOS=linux
    go build -o server.elf cmd\server\server.go 
) ELSE IF "%1"=="client" (
    REM -H windowsgui will disable the go window
    REM go build -o client.exe -ldflags "-H windowsgui" cmd\client\client.go
    go build -o client.exe cmd\client\client.go
    SET GOOS=linux
    go build -o client.elf cmd\client\client.go
) ELSE IF "%1"=="wingman" (
    REM go build -o wingman.exe -ldflags "-H windowsgui" cmd\wingman\wingman.go 
    go build -o wingman.exe cmd\wingman\wingman.go 
    SET GOOS=linux
    go build -o wingman.elf cmd\wingman\wingman.go 
) ELSE IF "%1"=="wingmandll" (
    go build -o wingman.dll -buildmode=c-shared .\cmd\wingman\wingman.go
) ELSE IF "%1"=="deploy" (
    .\makewin.bat client
    .\makewin.bat server
    .\makewin.bat wingman
    .\makewin.bat wingmandll
    mkdir build\upload
    mkdir build\static
    mkdir build\webui
    copy server.elf build\
    copy server.exe build\
    copy client.exe build\static\
    copy client.elf build\static\
    copy wingman.exe build\static\
    copy wingman.elf build\static\
    copy wingman.dll build\static\
    copy webui\* build\webui\
) ELSE IF "%1"=="coverage" (
    go test .\... -coverprofile="coverage.out"
    go tool cover -html="coverage.out"
) ELSE (
    echo "Unknown: %1"
)
