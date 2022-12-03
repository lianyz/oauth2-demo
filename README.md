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

### 1. 浏览器代理->授权客户端

```
http://localhost:9094
```

### 2. 授权客户端->浏览器代理（第一次重定向）

```
redirect http://localhost:9096/oauth/authorize 302
```

### 3. 浏览器代理->授权服务端

```
http://localhost:9096/oauth/authorize?
  client_id=222222&
  code_challenge=Qn3Kywp0OiU4NK_AFzGPlmrcYJDJ13Abj_jdL08Ahg8=&
  code_challenge_method=S256&
  redirect_uri=http://localhost:9094/oauth2&
  response_type=code&
  scope=all&
  state=xy
```

### 4. 授权服务端->浏览器代理
```
redirect http://localhost:9096/login 302
```

### 5. 浏览器代理->授权服务端
```
http://localhost:9096/login
```

### 6. 授权服务端->浏览器代理
```
redirect http://localhost:9096/auth 302
```

### 7. 浏览器代理->授权服务端
```
http://localhost:9096/auth
```

### 8. 浏览器代理->授权服务端
```
http://localhost:9096/oauth/authorize
```

### 9. 授权服务端->浏览器代理（第二次重定向）
```
redirect http://localhost:9094/oauth2?code=NJNMNJFHZGMTNDY4NS0ZYTY2LWIWZGUTYWIXNTEYZWU3MJY5&state=xyz 302
```

### 10 浏览器代理->授权客户端
```
http://localhost:9094/oauth2?code=NJNMNJFHZGMTNDY4NS0ZYTY2LWIWZGUTYWIXNTEYZWU3MJY5&state=xyz
```

### 11 授权客户端->授权服务端（直接通信，不通过浏览器）
```
POST http://localhost:9096/oauth/token
     code:NJNMNJFHZGMTNDY4NS0ZYTY2LWIWZGUTYWIXNTEYZWU3MJY5
     state:xyz
```

```
{
"access_token": "MMI3NZRHMGQTNWNMMC0ZOWUXLTKXYMETNTQ3ZGEYMMQ1YJFI",
"token_type": "Bearer",
"refresh_token": "NWY1YJU1YTYTZTC5YI01YMZJLWEXYMQTOGYZNWQYNJNLNDFM",
"expiry": "2022-12-03T15:40:06.737177+08:00"
}
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


授权服务端执行日志
```
2022/12/03 13:37:13 Server is running at 9096 port.
2022/12/03 13:37:13 Point your OAuth client Auth endpoint to http://localhost:9096/oauth/authorize
2022/12/03 13:37:13 Point your OAuth client Token endpoint to http://localhost:9096/oauth/token
2022/12/03 13:38:13 [Request]: oauth/authorize. method: GET url: /oauth/authorize?client_id=222222&code_challenge=Qn3Kywp0OiU4NK_AFzGPlmrcYJDJ13Abj_jdL08Ahg8=&code_challenge_method=S256&redirect_uri=http://localhost:9094/oauth2&response_type=code&scope=all&state=xyz
2022/12/03 13:38:13 [Request]: oauth/authorize. Get store.ReturnUri: [map[]] and Delete store.ReturnUri
2022/12/03 13:38:13 [Handler]: userAuthorize. r.Form: [map[client_id:[222222] code_challenge:[Qn3Kywp0OiU4NK_AFzGPlmrcYJDJ13Abj_jdL08Ahg8=] code_challenge_method:[S256] redirect_uri:[http://localhost:9094/oauth2] response_type:[code] scope:[all] state:[xyz]]]
2022/12/03 13:38:13 [Handler]: userAuthorize. set store.ReturnUri [map[client_id:[222222] code_challenge:[Qn3Kywp0OiU4NK_AFzGPlmrcYJDJ13Abj_jdL08Ahg8=] code_challenge_method:[S256] redirect_uri:[http://localhost:9094/oauth2] response_type:[code] scope:[all] state:[xyz]]]
2022/12/03 13:38:13 [Request]: userAuthorize. redirect to /login
2022/12/03 13:38:13 [Handler]: userAuthorize. userid: [], Get store.LoggedInUserID is null
2022/12/03 13:38:13 [Request]: login. method: GET url: /login
2022/12/03 13:39:24 [Request]: login. method: POST url: /login
2022/12/03 13:39:24 [Request]: login. redirect to /auth
2022/12/03 13:39:24 [Request]: auth. method: GET url: /auth
2022/12/03 13:40:06 [Request]: oauth/authorize. method: POST url: /oauth/authorize
2022/12/03 13:40:06 [Request]: oauth/authorize. Get store.ReturnUri: [map[client_id:[222222] code_challenge:[Qn3Kywp0OiU4NK_AFzGPlmrcYJDJ13Abj_jdL08Ahg8=] code_challenge_method:[S256] redirect_uri:[http://localhost:9094/oauth2] response_type:[code] scope:[all] state:[xyz]]] and Delete store.ReturnUri
2022/12/03 13:40:06 [Handler]: userAuthorize. userid: [lianyanze], Delete store.LoggedInUserID
2022/12/03 13:40:06 [Request]: token. method: POST url: /oauth/token
```

授权客户端执行日志
```
2022/12/03 13:37:16 Client is running at :9094 port. Please open http://localhost:9094
2022/12/03 13:38:13 [Request]: /. method: GET url: /
2022/12/03 13:38:13 [Request]: /. redirect to http://localhost:9096/oauth/authorize?client_id=222222&code_challenge=Qn3Kywp0OiU4NK_AFzGPlmrcYJDJ13Abj_jdL08Ahg8%3D&code_challenge_method=S256&redirect_uri=http%3A%2F%2Flocalhost%3A9094%2Foauth2&response_type=code&scope=all&state=xyz
2022/12/03 13:40:06 [Request]: token. method: GET url: /oauth2?code=NJNMNJFHZGMTNDY4NS0ZYTY2LWIWZGUTYWIXNTEYZWU3MJY5&state=xyz
```

浏览器代理获得的结果
```
{
"access_token": "MMI3NZRHMGQTNWNMMC0ZOWUXLTKXYMETNTQ3ZGEYMMQ1YJFI",
"token_type": "Bearer",
"refresh_token": "NWY1YJU1YTYTZTC5YI01YMZJLWEXYMQTOGYZNWQYNJNLNDFM",
"expiry": "2022-12-03T15:40:06.737177+08:00"
}
```
