package server

import (
	"fmt"
	"net/http"
)

func serveHiddenPage(res http.ResponseWriter, authErr error) {
	const hiddenPage = `<html>
<head>
  <title>ShadowProxy</title>
</head>
<body>
<h1>ShadowProxy Hidden Proxy Page!</h1>
%s<br/>
</body>
</html>`
	const AuthFail = "Please authenticate yourself to the proxy."
	const AuthOk = "Congratulations, you are successfully authenticated to the proxy! Go browse all the things!"

	res.Header().Set("Content-Type", "text/html")
	if authErr != nil {
		res.Header().Set("Proxy-Authenticate", "Basic realm=\"Secure Web Proxy\"")
		res.WriteHeader(http.StatusProxyAuthRequired)
		res.Write([]byte(fmt.Sprintf(hiddenPage, AuthFail)))
		return
	}
	res.Write([]byte(fmt.Sprintf(hiddenPage, AuthOk)))
	return
}
