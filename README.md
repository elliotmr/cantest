# Setup
This module is a basic test for SocketCAN functionality in Go. This will only work with a patched version of 
[golang.org/x/sys/unix](https://godoc.org/golang.org/x/sys/unix) (see patch in `unix.diff`).

After patching you must also set up a virtual can bus to run this code. In your shell type the following commands.

```bash
sudo modprobe vcan
sudo ip link add dev vcan0 type vcan
sudo ip link set up vcan0
```

# Running the Test
Simply run `go test -v github.com/elliotmr/cantest`