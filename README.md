# OAuth2 Demo


## run

### 1. 修改Makefile

如果需要，修改Makefile中SERVER_ADDR、CLIENT_ADDR的IP地址

### 2. 运行服务端

```
# make run
mkdir -p bin/amd64
building server...
CGO_ENABLED=0 GOARCH=amd64 go build -o bin/amd64 ./server
building client...
CGO_ENABLED=0 GOARCH=amd64 go build -o bin/amd64 ./client
copy static files...
mkdir -p bin/amd64/static
cp ./server/static/*.html ./bin/amd64/static/
./bin/amd64/server -d=false -ip=192.168.34.2
2022/12/04 00:31:58 Server is running at 9096 port.
2022/12/04 00:31:58 Point your OAuth client Auth endpoint to http://192.168.34.2:9096/oauth/authorize
2022/12/04 00:31:58 Point your OAuth client Token endpoint to http://192.168.34.2:9096/oauth/token
```

### 3. 运行客户端

```
# make run.client
curl "http://192.168.34.2:9096/register?clientId=CLIENT_12345&clientSecret=CLIENT_xxxxx&clientAddr=http://192.168.34.2:9094"
{"CLIENT_ID":"CLIENT_12345","CLIENT_SECRET":"CLIENT_xxxxx"}
./bin/amd64/client -id CLIENT_12345 -secret CLIENT_xxxxx -addr http://192.168.34.2:9094 -server http://192.168.34.2:9096
ClientId: CLIENT_12345 ClientSecret: CLIENT_xxxxx ClientAddr: http://192.168.34.2:9094
2022/12/04 00:33:01 Client is running at http://192.168.34.2:9094. Please open http://192.168.34.2:9094
```

### 4. 在浏览器中访问http://192.168.34.2:9094

在页面中输入用户名密码, 点击同意授权按钮，成功后，浏览器显示如下内容
```
{
  "access_token": "OWFLNZEXNGETMJUXNY0ZMZE2LWI1NJYTYMEWNJQZNZAWM2JJ",
  "token_type": "Bearer",
  "refresh_token": "NWRHMZA3ZTGTZJVLMC01NWVLLWEZZTATMJUZYJDHMWIYMDJI",
  "expiry": "2022-12-04T02:34:09.473358159+08:00"
}
```

### 5. 将$(access_token) 拷贝至配置文件
```
# echo " OWFLNZEXNGETMJUXNY0ZMZE2LWI1NJYTYMEWNJQZNZAWM2JJ" >> ./webhook-config/config
```

### 6. 备份并修改k8s的配置文件
```
# make install.webhook
add token to webhook-config/config under oauth2-user
cp ~/.kube/config ~/.kube/config.bak
cp ./webhook-config/config ~/.kube/
cp /etc/kubernetes/manifests/kube-apiserver.yaml /etc/kubernetes/manifests/kube-apiserver.yaml.bak
cp ./webhook-config/kube-apiserver.yaml /etc/kubernetes/manifests/
cp /etc/config/webhook-config.json /etc/config/webhook-config.json.bak
cp ./webhook-config/webhook-config.json /etc/config/
```

### 7. 等待并查看k8s ApiServer是否重启成功

```
# k get po
The connection to the server 192.168.34.2:6443 was refused - did you specify the right host or port?

...

# k get po
NAME                          READY   STATUS      RESTARTS        AGE
centos-578b69b65f-jl9ww       0/1     Running     27 (3d5h ago)   90d
config-volume-pod             0/1     Completed   0               90d
envoy-6958c489d9-hmj7n        1/1     Running     24 (3d5h ago)   86d
hello-volume                  1/1     Running     27 (3d5h ago)   90d
hostnames-7fb5498f8d-bkwvt    1/1     Running     24 (3d5h ago)   54d
hostnames-7fb5498f8d-sb4ql    1/1     Running     24 (3d5h ago)   54d
hostnames-7fb5498f8d-sr9kt    1/1     Running     25 (3d5h ago)   54d
myapp-pod                     1/1     Running     105 (13m ago)   56d
nginx                         2/2     Running     49 (3d5h ago)   61d
patch-demo-68fc587f7c-5zlvw   1/1     Running     106 (12m ago)   56d
patch-demo-68fc587f7c-mjlt8   1/1     Running     106 (10m ago)   56d
```

### 8. 使用config中配置的oauth2-user用户访问k8s

```
# k get po --user oauth2-user
Error from server (Forbidden): pods is forbidden: User "lianyanze" cannot list resource "pods" in API group "" in the namespace "default"
```

用户获取成功，但没有权限

### 9. 配置用户在k8s中的权限

```
# make auth
kubectl delete -f ./webhook-config/role.yaml
role.rbac.authorization.k8s.io "example-role" deleted
kubectl delete -f ./webhook-config/rolebinding-user.yaml
rolebinding.rbac.authorization.k8s.io "example-rolebinding" deleted
kubectl apply -f ./webhook-config/role.yaml
role.rbac.authorization.k8s.io/example-role created
kubectl apply -f ./webhook-config/rolebinding-user.yaml
rolebinding.rbac.authorization.k8s.io/example-rolebinding created
```

### 10. 再次使用config中配置的oauth2-user用户访问k8s

```
# k get po --user oauth2-user
NAME                          READY   STATUS      RESTARTS        AGE
centos-578b69b65f-jl9ww       0/1     Running     27 (4d ago)     91d
config-volume-pod             0/1     Completed   0               90d
envoy-6958c489d9-hmj7n        1/1     Running     24 (4d ago)     87d
hello-volume                  1/1     Running     27 (4d ago)     91d
hostnames-7fb5498f8d-bkwvt    1/1     Running     24 (4d ago)     55d
hostnames-7fb5498f8d-sb4ql    1/1     Running     24 (4d ago)     55d
hostnames-7fb5498f8d-sr9kt    1/1     Running     25 (4d ago)     55d
myapp-pod                     1/1     Running     106 (21m ago)   56d
nginx                         2/2     Running     49 (4d ago)     62d
patch-demo-68fc587f7c-5zlvw   1/1     Running     107 (21m ago)   56d
patch-demo-68fc587f7c-mjlt8   1/1     Running     107 (19m ago)   56d
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
