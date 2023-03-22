# biyesheji

[TOC]

## 账号的注册

账号内容为以下四项：

- 用户名
- 密码
- 权限：admin，user
- 工作路径

### 数据

1、用户的数据格式以json格式保存在本地

### 说明

1、每个集群只能拥有一个admin，平台在被**初次**部署时创建admin用户，此后只能创建user用户。

2、admin用户不需要输入工作路径

3、用户名和用户路径绑定，且唯一



## 账号的登录

登录只需要三个部分：

- 用户名
- 密码

### 数据

1、用户名

### 说明

1、权限和工作路径在用户输入完以上两个部分后，会得到自动显示，且不可更改。

2、如返回的是权限是admin，则可以使用注册按钮；若是user，则注册按钮锁死
