# Spatch

## Installation

Fetch dependencies:
```
go get -u
```

Setup SSH host keys
```
ssh-keygen -t rsa -b 4096 -f keys/ssh_host_rsa_key
ssh-keygen -t dsa -b 1024 -f keys/ssh_host_dsa_key
```

## Running

```
go run *.go
```

## Connecting

```
ssh -p 8080 test@127.0.0.1
```
Password: `test`
