FROM golang:1.16 AS build
WORKDIR /playproxy
ADD . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/shadowproxy .

FROM nginx:stable-alpine

WORKDIR /root

COPY --from=build /shadowproxy/bin/shadowproxy /root/shadowproxy
COPY --from=build /shadowproxy/docker/config.yaml /root/config/config.yaml

CMD nginx && /root/shadowproxy -c /root/config/config.yaml
