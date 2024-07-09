module github.com/Fan-Fuse/spotify-service

go 1.21.4

require (
	github.com/Fan-Fuse/artist-service v0.0.0-20240709200758-a970aa4e7426
	github.com/Fan-Fuse/config-service v0.0.0-20240705130120-98f1060bcd87
	github.com/Fan-Fuse/user-service v0.0.0-20240709024251-7e60dd68c16c
	github.com/rabbitmq/amqp091-go v1.10.0
	github.com/zmb3/spotify/v2 v2.4.2
	go.uber.org/zap v1.27.0
	golang.org/x/oauth2 v0.20.0
	google.golang.org/grpc v1.65.0
	google.golang.org/protobuf v1.34.2
)

require (
	github.com/golang/protobuf v1.5.4 // indirect
	go.uber.org/multierr v1.10.0 // indirect
	golang.org/x/net v0.25.0 // indirect
	golang.org/x/sys v0.20.0 // indirect
	golang.org/x/text v0.15.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240528184218-531527333157 // indirect
)
