set shell := ["powershell.exe", "-c"]
default:
   just --list 

build:
    go build .\src\

buildAndDeploy:
    go build -o "C:\Program Files\Docker\Docker\resources\bin\docker.exe" .\src\

buildDeployAndRun: buildAndDeploy
    docker
    
