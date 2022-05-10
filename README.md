![logo](logo.png)

[![License](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)  [![GoDoc](https://godoc.org/github.com/cloudflare/cfssl?status.svg)](https://pkg.go.dev/github.com/nisainan/shadowproxy)

A proxy based on native https protocal. But can response a http2 website that you configured without authorization to hide your proxy.

## Features

- Native proxy
- TLS support
- Authorization 
- Camouflage traffic

## Installing

~~~shell
$ git clone https://github.com/nisainan/shadowproxy.git
$ cd shadowproxy
$ make 
~~~

You can set GOOS and GOARCH environment variables to allow Go to cross-compile alternative platforms.

The resulting binaries will be in the bin folder:

~~~shell
$ tree bin
bin
├── shadowproxy
~~~

Edit  `config.yaml` with your own data

~~~yaml
listen-address: "0.0.0.0:443" # listen address
username: "username" # authorization username
password: "password" # authorization password
probe-resist-domain: "shengtao.link" # authentication url
cert-file: "xxxx" # cert file localtion
key-file: "xxxx" # key file localtion
cheat-host: "127.0.0.1:80" # cheat-host, make sure this server works
~~~

~~~shell
$ shadowproxy -c config.yaml
~~~

## Usage

1. Use [SwitchyOmega](chrome-extension://padekgcemlokbadohgkifijomclgjgif/options.html#!/about) in your browser
2. Add a https proxy.Don't forget filling in username and passowrd
3. Access `probe-resist-domain` in your browser
4. Congratulations,Go browse all the things!

## License

ShadowProxy source code is available under the MIT [License](https://github.com/nisainan/shadowproxy/blob/master/LICENSE).

## Thanks

[forwardproxy](https://github.com/caddyserver/forwardproxy.git)
