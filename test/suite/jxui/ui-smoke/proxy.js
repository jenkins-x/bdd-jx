var http = require('http'),
    httpProxy = require('http-proxy'),
    url = require('url'),
    targetBaseURL = url.parse(process.env.TARGET_BASE_URL || 'http://127.0.0.1:9000', false),
    proxyPort = process.env.PROXY_PORT || 9007,
    username = process.env.JENKINS_USERNAME || 'admin',
    password = process.env.JENKINS_PASSWORD,
    auth = "Basic " + Buffer.from(username + ":" + password).toString('base64'),
    authInfo = (password ? '' : ' without') + ' injecting basic auth credentials',
    injectBasicAuthCreds = req => {
        if (password && !req.headers.Authorization) {
    	        req.headers.Authorization = auth;
        }
    }
    proxy = httpProxy.createProxyServer({
        target: {
            host: targetBaseURL.hostname,
            port: targetBaseURL.port
        }
    }),
    proxyServer = http.createServer((req, res) => {
    	    if (req.url == '/shutdown-proxy') {
            console.log('Terminating proxy');
            process.exit();
        }
        injectBasicAuthCreds(req);
        proxy.web(req, res);
    });

// Proxy WebSocket requests as well
proxyServer.on('upgrade', (req, socket, head) => {
    injectBasicAuthCreds(req);
    proxy.ws(req, socket, head);
});

console.log('Proxying ' + targetBaseURL.hostname + ':' + targetBaseURL.port + ' on 127.0.0.1:' + proxyPort + authInfo);
proxyServer.listen(proxyPort);
