# 一款GIN的CLI工具

安装

```shell
go install github.com/hxoreyer/ktool
```

创建应用

```shell
ktool init -n app
```

创建应用后会自动生成文件夹和mod等文件

```shell
cd app
```

创建路由

1. 修改yaml文件的内容

   ```yaml
   groups:
   - name: api/v1
     infos: 
       - method: POST
         path: /register
         function: Register
         file: user
       - method: GET
         path: /login
         function: Login
         file: user
   - name: api/v2
     infos: 
       - method: POST
         path: /register2
         function: Register2
         file: admin
       - method: GET
         path: /login2
         function: Login2
         file: admin
   ```

2. 生成controller文件，默认为 config.yaml

   ```shell
   ktool router -p config.yaml
   ```

数据库

1. 初始化数据库

   ```shell
   ktool mysql create -u hxoreyer -p password -a 123.11.22.321:1234 -d tableName
   ```

2. 通过model文件生成service文件

   ```go
   // 在models文件夹下创建user.go
   package models
   
   import (
   	"gorm.io/gorm"
   )
   
   type User struct {
   	gorm.Model
   	UserName string `gorm:"column:username"`
   	Password string `gorm:"column:password"`
   }
   ```

   ```shell
   #生成service文件(-n 为models里的文件名 不加后缀)
   ktool mysql new -n user
   ```

运行

```shell
go mod tidy
go run .
# 地址为localhost:5201
```

