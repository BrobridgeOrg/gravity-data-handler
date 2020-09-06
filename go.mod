module gravity-data-handler

go 1.13

require (
	github.com/BrobridgeOrg/gravity-api v0.0.0-20200816194343-a91dd0fa3335
	github.com/golang/protobuf v1.4.2
	github.com/lithammer/go-jump-consistent-hash v1.0.1
	github.com/nats-io/nats-streaming-server v0.17.0 // indirect
	github.com/nats-io/nats.go v1.9.1
	github.com/nats-io/stan.go v0.6.0
	github.com/prometheus/common v0.4.0
	github.com/sirupsen/logrus v1.4.2
	github.com/sony/sonyflake v1.0.0
	github.com/spf13/viper v1.6.2
	golang.org/x/net v0.0.0-20200301022130-244492dfa37a
	google.golang.org/grpc v1.31.0
)

//replace github.com/BrobridgeOrg/gravity-api => ../gravity-api
