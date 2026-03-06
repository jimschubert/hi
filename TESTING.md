# Testing

## Daemon

### buf curl

`buf curl` uses the Connect protocol (which supports both HTTP/1.1 and HTTP/2):

```shell
buf curl \
  --schema . \
  --unix-socket /tmp/hi.sock \
  http://localhost/hi.v1.HiService/Ping
```

### grpcurl

`grpcurl` uses the gRPC protocol (only HTTP/2).

The daemon mounts both `grpc.reflection.v1` and `grpc.reflection.v1alpha` handlers, so other tools with reflection should work.

**Get service details**
```shell
grpcurl \
  -plaintext \
  unix:///tmp/hi.sock describe hi.v1.HiService
```

**Make a request**
```shell
grpcurl \
  -plaintext \
  -d '{}' \
  unix:///tmp/hi.sock hi.v1.HiService/Ping
```
