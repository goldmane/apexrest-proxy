require('dotenv').config();

const express = require('express');
const app = express();
const httpProxy = require('http-proxy');

const proxies = JSON.parse(process.env.PROXY_CONFIGS);
const proxy = httpProxy.createProxyServer({});

app.all('/:clientId/', (req, res) => {
    const clientId = req.params['clientId'] || undefined;
    const config = proxies[clientId];
    if (!config) {
        res.sendStatus(400).end();
    } else {
        const domain = proxies['local'].target_domain
        console.log(`Proxy to ${domain}`)
        proxy.web(req, res, {
            target: domain,
            secure: false
        });
    }
});

const PORT = process.env.PORT || 3000;
app.listen(PORT, () => {
    console.log(`Server started on port ${PORT}`);
});