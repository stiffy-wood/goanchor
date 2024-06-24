set shell := ["powershell.exe", "-c"]
default:
   just --list 

runDocker:
    go run .\src\docker\ .\Dockerfile.example
