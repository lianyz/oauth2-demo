# OAuth2 Demo

```
make run
```

```
curl http://localhost:9096/credentials
{"CLIENT_ID":"25e29f3a","CLIENT_SECRET":"ad08b7aa"}
```

```
curl http://localhost:9096/token\?grant_type\=client_credentials\&client_id\=25e29f3a\&client_secret\=ad08b7aa\&scope\=all
{"access_token":"PXK95V2MPVKPRCVYVPLY-Q","expires_in":7200,"scope":"all","token_type":"Bearer"}
```

```
curl http://localhost:9096/protected\?access_token\=PXK95V2MPVKPRCVYVPLY-Q
Hello, I'm protected!
```
