@echo off

IF "%1"=="runserver" (
    go run cmd\server\server.go --listenaddr 127.0.0.1:8080
) ELSE IF "%1"=="runclient" (
    go run cmd\client\client.go
) ELSE IF "%1"=="runexecutor" (
    go run cmd\executor\executor.go 
) ELSE IF "%1"=="server" (
    go build cmd\server\server.go 
    REM Doesnt work
    REM SET GOOS=linux
    REM go build -o server.elf cmd\server\server.go 
) ELSE IF "%1"=="client" (
    go build cmd\client\client.go
) ELSE IF "%1"=="executor" (
    go build cmd\executor\executor.go 
) ELSE IF "%1"=="deploy" (
    echo "bla"
) ELSE (
    echo "Unknown: %1"
)