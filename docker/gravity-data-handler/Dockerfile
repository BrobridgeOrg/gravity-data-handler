FROM golang:1.15.3-alpine3.12 AS builder
WORKDIR $GOPATH/src/github.com/BrobridgeOrg/gravity-data-handler/
COPY . .

RUN apk add --update git
RUN CGO_ENABLED=0 GO111MODULE=on go build -installsuffix cgo -o /gravity-data-handler ./cmd/gravity-data-handler/gravity-data-handler.go

FROM alpine:3.12
COPY --from=builder /gravity-data-handler /gravity-data-handler
COPY ./config/config.toml /config.toml
COPY ./rules /rules
COPY ./docker/gravity-data-handler/startup.sh /startup.sh
EXPOSE 44445
CMD ["/bin/sh", "/startup.sh"]
