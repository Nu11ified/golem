module golem-test-app

go 1.23.0

toolchain go1.24.2

replace github.com/Nu11ified/golem => ../

require github.com/Nu11ified/golem v0.1.0

require (
	golang.org/x/net v0.38.0 // indirect
	golang.org/x/sys v0.31.0 // indirect
	golang.org/x/text v0.23.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250324211829-b45e905df463 // indirect
	google.golang.org/grpc v1.73.0 // indirect
	google.golang.org/protobuf v1.36.6 // indirect
	nhooyr.io/websocket v1.8.17 // indirect
)
