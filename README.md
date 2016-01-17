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

Setup SSH key for login
```
ssh-keygen -t rsa -b 4096 -f keys/id_rsa
```

Setup your endpoints.yml
```
$ cat endpoints.yml
- host: 1.2.3.4
  port: 22
  user: admin
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
