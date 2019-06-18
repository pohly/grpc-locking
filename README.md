This test case demonstrates a false (?) negative in the Go race
detector.

The data race is in the implementation of the SayHello method: because
gRPC invokes that in goroutines (one per request), the implementation
must protect against concurrent access to the shared "s.names" struct
member.

The following invocation tests that without involving gRPC and it
reliably reports the data race, regardless of the delay:

``` shell
$ go test -race -run=TestAPI . -args -delay=1ms
...
==================
WARNING: DATA RACE
Read at 0x00c00000e580 by goroutine 9:
  help.(*server).SayHello()
      /work/grpc-locking/main.go:48 +0x1a0
  help.TestAPI.func1()
      /work/grpc-locking/main_test.go:27 +0xf4

Previous write at 0x00c00000e580 by goroutine 7:
  help.(*server).SayHello()
      /work/grpc-locking/main.go:48 +0x234
  help.TestAPI.func1()
      /work/grpc-locking/main_test.go:27 +0xf4

Goroutine 9 (running) created at:
  help.TestAPI()
      /work/grpc-locking/main_test.go:31 +0x228
  testing.tRunner()
      /nvme/gopath/go/src/testing/testing.go:865 +0x163

Goroutine 7 (finished) created at:
  help.TestAPI()
      /work/grpc-locking/main_test.go:29 +0x1c5
  testing.tRunner()
      /nvme/gopath/go/src/testing/testing.go:865 +0x163
==================
```

But when going through gRPC, the race is only detected for delays
smaller than (roughly) 5ms (your mileage may vary):

``` shell
$ go test -v -race -run=TestGRPC . -args -delay=10ms
=== RUN   TestGRPC
2019/06/18 19:49:40 Received: foo in goroutine 23 [running]:
runtime/debug.Stack(0xa794cd, 0x0, 0x0)
	/nvme/gopath/go/src/runtime/debug/stack.go:24 +0xab
help.(*server).SayHello(0xc0000f6500, 0xc5fe60, 0xc0001b64b0, 0xc0001b64e0, 0xc0000f6500, 0xc00023e190, 0xc0000ba000)
	/work/grpc-locking/main.go:45 +0x45
google.golang.org/grpc/examples/helloworld/helloworld._Greeter_SayHello_Handler(0xaf85a0, 0xc0000f6500, 0xc5fe60, 0xc0001b64b0, 0xc00023e190, 0x0, 0xc5fe60, 0xc0001b64b0, 0xc0001c406a, 0x5)
	/nvme/gopath/pkg/mod/google.golang.org/grpc@v1.21.1/examples/helloworld/helloworld/helloworld.pb.go:158 +0x2fd
google.golang.org/grpc.(*Server).processUnaryRPC(0xc0000d0600, 0xc647c0, 0xc0000d0d80, 0xc000284100, 0xc0000e1ef0, 0x104a130, 0x0, 0x0, 0x0)
	/nvme/gopath/pkg/mod/google.golang.org/grpc@v1.21.1/server.go:998 +0x937
google.golang.org/grpc.(*Server).handleStream(0xc0000d0600, 0xc647c0, 0xc0000d0d80, 0xc000284100, 0x0)
	/nvme/gopath/pkg/mod/google.golang.org/grpc@v1.21.1/server.go:1278 +0x1352
google.golang.org/grpc.(*Server).serveStreams.func1.1(0xc0001c4010, 0xc0000d0600, 0xc647c0, 0xc0000d0d80, 0xc000284100)
	/nvme/gopath/pkg/mod/google.golang.org/grpc@v1.21.1/server.go:717 +0xad
created by google.golang.org/grpc.(*Server).serveStreams.func1
	/nvme/gopath/pkg/mod/google.golang.org/grpc@v1.21.1/server.go:715 +0xb9
2019/06/18 19:49:40 Received: bar in goroutine 40 [running]:
runtime/debug.Stack(0xa794cd, 0x0, 0x0)
	/nvme/gopath/go/src/runtime/debug/stack.go:24 +0xab
help.(*server).SayHello(0xc0000f6500, 0xc5fe60, 0xc0001a03f0, 0xc0001a0420, 0xc0000f6500, 0xc000112730, 0xc00008ea80)
	/work/grpc-locking/main.go:45 +0x45
google.golang.org/grpc/examples/helloworld/helloworld._Greeter_SayHello_Handler(0xaf85a0, 0xc0000f6500, 0xc5fe60, 0xc0001a03f0, 0xc000112730, 0x0, 0xc5fe60, 0xc0001a03f0, 0xc0000e8858, 0x5)
	/nvme/gopath/pkg/mod/google.golang.org/grpc@v1.21.1/examples/helloworld/helloworld/helloworld.pb.go:158 +0x2fd
google.golang.org/grpc.(*Server).processUnaryRPC(0xc0000d0600, 0xc647c0, 0xc0000d0d80, 0xc00017c200, 0xc0000e1ef0, 0x104a130, 0x0, 0x0, 0x0)
	/nvme/gopath/pkg/mod/google.golang.org/grpc@v1.21.1/server.go:998 +0x937
google.golang.org/grpc.(*Server).handleStream(0xc0000d0600, 0xc647c0, 0xc0000d0d80, 0xc00017c200, 0x0)
	/nvme/gopath/pkg/mod/google.golang.org/grpc@v1.21.1/server.go:1278 +0x1352
google.golang.org/grpc.(*Server).serveStreams.func1.1(0xc0001c4010, 0xc0000d0600, 0xc647c0, 0xc0000d0d80, 0xc00017c200)
	/nvme/gopath/pkg/mod/google.golang.org/grpc@v1.21.1/server.go:717 +0xad
created by google.golang.org/grpc.(*Server).serveStreams.func1
	/nvme/gopath/pkg/mod/google.golang.org/grpc@v1.21.1/server.go:715 +0xb9
--- PASS: TestGRPC (0.01s)
PASS
ok  	help	1.032s

$ go test -v -race -run=TestGRPC . -args -delay=1ms
=== RUN   TestGRPC
...
==================
WARNING: DATA RACE
Read at 0x00c0000e8520 by goroutine 26:
  help.(*server).SayHello()
      /work/grpc-locking/main.go:48 +0x1a0
  google.golang.org/grpc/examples/helloworld/helloworld._Greeter_SayHello_Handler()
      /nvme/gopath/pkg/mod/google.golang.org/grpc@v1.21.1/examples/helloworld/helloworld/helloworld.pb.go:158 +0x2fc
  google.golang.org/grpc.(*Server).processUnaryRPC()
      /nvme/gopath/pkg/mod/google.golang.org/grpc@v1.21.1/server.go:998 +0x936
  google.golang.org/grpc.(*Server).handleStream()
      /nvme/gopath/pkg/mod/google.golang.org/grpc@v1.21.1/server.go:1278 +0x1351
  google.golang.org/grpc.(*Server).serveStreams.func1.1()
      /nvme/gopath/pkg/mod/google.golang.org/grpc@v1.21.1/server.go:717 +0xac
...
```

As the debug output shows, the function does get called from different
goroutines (24 and 40 in the example above).

Perhaps gRPC does something while waiting for new requests that causes
the race detector to believe that there is some ordering that prevents
races.
