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


[Build your Own OAuth2 Server in Go: Client Credentials Grant Flow](https://medium.com/@cyantarek/build-your-own-oauth2-server-in-go-7d0f660732c3)




## 授权码

### client向server发送授权请求

```
/oauth/authorize?client_id=222222
&code_challenge=Qn3Kywp0OiU4NK_AFzGPlmrcYJDJ13Abj_jdL08Ahg8=
&code_challenge_method=S256
&redirect_uri=http://localhost:9094/oauth2
&response_type=code
&scope=all
&state=xyz
```


## OAuth2 访问流程

* 资源拥有者
* 浏览器代理（代理资源拥有者）
* 授权客户端（第三方软件）
* 授权服务端
* 受保护资源

### 1. 浏览器代理->第三方软件

```
http://localhost:9094
```

### 2. 第三方软件->浏览器代理（第一次重定向）

```
redirect http://localhost:9096/oauth/authorize 302

http://localhost:9096/oauth/authorize?
  client_id=222222&
  code_challenge=Qn3Kywp0OiU4NK_AFzGPlmrcYJDJ13Abj_jdL08Ahg8=&
  code_challenge_method=S256&
  redirect_uri=http://localhost:9094/oauth2&
  response_type=code&
  scope=all&
  state=xy
```

### server调用UserAuthorizationHandler处理
client->server
/oauth/authorize 如果未设置存储store.LoggedInUserID，则返回302，且在Header中设置Location: /login
                 否则，跳转至client的http://localhost:9094/oauth2
/login 登录成功后，设置store.LoggedInUserID为username, 并返回302，且在Header中设置Location: /auth
/auth 如果未设置存储store.LoggedInUserID，则返回302，且在Header中设置Location: /login
      否则，显示授权确认的页面，点击授权按钮后，跳转至/oauth/authorize

server->client
/oauth2?code=MDFHNMFHYZMTZDQ3MC0ZZDFKLWE2NDMTOTHJNWY0ODGXNDHJ&state=xyz

client0>server
/oauth/token



