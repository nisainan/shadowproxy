FROM golang:1.16 AS build
WORKDIR /playproxy
ADD . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/playproxy .

FROM nginx:stable-alpine

WORKDIR /root

COPY --from=build /playproxy/bin/playproxy /root/playproxy
COPY --from=build /playproxy/docker/config.yaml /root/config/config.yaml

CMD nginx && /root/playproxy -c /root/config/config.yaml
