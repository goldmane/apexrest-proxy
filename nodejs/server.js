require('dotenv').config();

const util = require('util');
const jsforce = require('jsforce');

const express = require('express');
const app = express();
const httpProxy = require('http-proxy');

const proxy_configs = JSON.parse(process.env.PROXY_CONFIGS);
const proxy = httpProxy.createProxyServer({});

const connections = {};

function getConfigByKey(key) {
    return proxy_configs[key]
}

async function getSalesforceAccessTokenByConfigKey(config_key) {
    // get the config object
    const config = getConfigByKey(config_key)
    // connect to salesforce org
    const conn = new jsforce.Connection({
        instanceUrl: config.instance_url,
        oauth2: {
            clientId: config.oauth.client_id,
            clientSecret: config.oauth.client_secret,
            loginUrl: config.instance_url
        },
        loginUrl: config.login_url
    });
    const userInfo = await conn.authorize({ grant_type: "client_credentials" })
    // get the auth token
    const access_token = conn.accessToken;
    return access_token;
}

proxy.on('proxyReq', async function(proxyReq, req, res, options) {
});

proxy.on('proxyRes', function (proxyRes, req, res) {
});

app.all('/:config_key/', async (req, res) => {
    const config_key = req.params['config_key'] || undefined;
    const config = proxy_configs[config_key];
    if (!config) {
        res.sendStatus(400).end();
    } else {
        const access_token = await getSalesforceAccessTokenByConfigKey(config_key);
        if (access_token) {
            req.headers['Authorization'] = `Bearer ${access_token}`;
            req.headers['X-Proxy-Auth'] = access_token;
        }
        const proxy_target = proxy_configs[config_key].proxy_target
        console.log(`Proxy to ${proxy_target}`)
        proxy.web(req, res, {
            target: proxy_target,
            secure: false,
            ignorePath: true,
            changeOrigin: true
        }, (e) => {
            console.error(e);
        });
    }
});

const PORT = process.env.PORT || 3000;
app.listen(PORT, () => {
    console.log(`Server started on port ${PORT}`);
});