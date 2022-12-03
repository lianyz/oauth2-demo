# OAuth2 Demo


## run

### 1. 修改Makefile

如果需要，修改Makefile中SERVER_ADDR、CLIENT_ADDR的IP地址

### 2. 运行服务端

```

```




* [Build your Own OAuth2 Server in Go: Client Credentials Grant Flow](https://medium.com/@cyantarek/build-your-own-oauth2-server-in-go-7d0f660732c3)
* [go-oauth2/oauth2](https://github.com/go-oauth2/oauth2/tree/master/example)



## OAuth2.0 核心
OAuth2.0授权的核心是颁发访问令牌、使用访问令牌

## OAuth2.0 四种许可类型
* 授权码许可（Authorization Code）,是最经典、最完备、最安全、应用最广泛的许可类型
* 隐式许可（Implicit）
* 客户端凭证许可（Client Credentials）
* 资源拥有者凭证许可（Resource Owner Credentials）

## OAuth2.0 五种角色

* 资源拥有者
* 浏览器代理（代理资源拥有者）
* 授权客户端（第三方软件）
* 授权服务端
* 受保护资源

## OAuth2.0 授权码许可流程

* 第一步：获取授权码 （授权客户端与授权服务端**间接通信**，中间通过了浏览器代理）
* 第二步：获取访问令牌（授权客户端与授权服务端**直接通信**）

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

### 5. 浏览器代理->授权服务端，让用户登录
```
http://localhost:9096/login
```

### 6. 授权服务端->浏览器代理
```
redirect http://localhost:9096/auth 302
```

### 7. 浏览器代理->授权服务端，让用户确认授权
```
http://localhost:9096/auth
```

### 8. 浏览器代理->授权服务端，生成授权码
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

### 11 授权客户端->授权服务端（直接通信，不通过浏览器，生成访问令牌）
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
