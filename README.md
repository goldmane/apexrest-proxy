# ApexRest Proxy

Why? This proxy will provide two main functions:

1. Allow us to use ApexRest for things like Google API Webhooks which provide some form of Authentication header. This is because the platform will reject requests that might have a random JWT which it cannot verify.
2. Allow us to make authenticated calls into the ApexRest route. This is important due to the limitations of the guest user profile.

## How it works:
- A configuration is built based on a few items:
    - `proxy_target` - this is the base URL to target on the proxy.
    - `instance_url` - 
    - `login_url` - 
    - Oauth
        - `client_id` - 
        - `client_secret` - 

## How it's run
The app runs inside of a Docker container on Heroku. 

## Setting up the Connected App
- Assuming you want to use simple Client Credentials:
    - Allow client credentials
    - Add the (api) scope
    - Set the running user to a user with [API Enabled]