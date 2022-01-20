Deployweb-operator演示文档

### 测试环境

minikube v1.13.0

go 1.16.13 linux/amd64

kubectl v1.19.0

docker v20.10.5

操作系统 CentOS7.6

### 功能介绍

通过该operator，可实现在k8s集群自动部署web前端镜像，并可指定对外暴露服务访问端口

### 功能实现

通过创建 ReplicationController 和 service 实现前端web应用镜像的部署

dw == Deployweb

![img](https://www.webkey5.com/media/editor/dw%20flow_20220120090501091361.png)

### 使用教程

1. make deploy IMG=ccr.ccs.tencentyun.com/oldcc/k8s-operator:1.16-alpine

2. vim deployweb.yml

   ```
   apiVersion: web.czy.io/v1
   kind: Deployweb
   metadata:
     name: deployweb-sample
   spec:
     image: ccr.ccs.tencentyun.com/oldcc/k8s-nginx-test:v1
     port: 30010
     replicas: 2
   ```

3. kubectl apply -f deployweb.yaml

4. 使用浏览器访问 http://host:30010 查看





