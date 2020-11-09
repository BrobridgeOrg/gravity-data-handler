module github.com/BrobridgeOrg/gravity-data-handler

go 1.13

require (
	github.com/BrobridgeOrg/gravity-adapter-nats v0.0.0-20201102201841-dd912fffd409 // indirect
	github.com/BrobridgeOrg/gravity-api v0.2.2
	github.com/cfsghost/gosharding v0.0.2
	github.com/cfsghost/parallel-chunked-flow v0.0.2
	github.com/golang/protobuf v1.4.2
	github.com/json-iterator/go v1.1.10
	github.com/lithammer/go-jump-consistent-hash v1.0.1
	github.com/nats-io/nats.go v1.10.0
	github.com/sirupsen/logrus v1.6.0
	github.com/soheilhy/cmux v0.1.4
	github.com/spf13/viper v1.7.1
	go.uber.org/automaxprocs v1.3.0
	golang.org/x/net v0.0.0-20200301022130-244492dfa37a
	golang.org/x/sys v0.0.0-20200202164722-d101bd2416d5 // indirect
	google.golang.org/grpc v1.32.0
	google.golang.org/grpc/examples v0.0.0-20200807164945-d3e3e7a46f57 // indirect
	google.golang.org/protobuf v1.25.0 // indirect
	gopkg.in/yaml.v2 v2.2.8 // indirect
)

//replace github.com/BrobridgeOrg/gravity-api => ../gravity-api

//replace github.com/cfsghost/grpc-connection-pool => /Users/fred/works/opensource/grpc-connection-pool
//replace github.com/cfsghost/parallel-chunked-flow => /Users/fred/works/opensource/parallel-chunked-flow
