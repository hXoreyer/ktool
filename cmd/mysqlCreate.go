/*
Copyright © 2023 hxoreyer

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"fmt"
	"os"
	"text/template"

	"github.com/spf13/cobra"
)

const tpl = `package databases

import (
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var Db *gorm.DB

func init() {
	dsn := "{{.SQLName}}/{{.Name}}?charset=utf8mb4&loc=Local&parseTime=true"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Println(err)
		return
	}
	Db = db
}`

// mysqlCreateCmd represents the mysqlCreate command
var mysqlCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "MYSQL代码",
	Long:  `创建MYSQL初始化代码`,
	Example: `ktool mysql create -u hxoreyer -p password -a 123.11.22.321:1234 -d tableName
ktool mysql create --user hxoreyer --passwd password --address 123.11.22.321:1234 --database tableName`,
	Run: func(cmd *cobra.Command, args []string) {
		data := map[string]string{
			"user":     "",
			"passwd":   "",
			"address":  "",
			"database": "",
		}
		b := true
		for k := range data {
			val, err := cmd.Flags().GetString(k)
			if err != nil || val == "" {
				fmt.Printf("请填写 %s", k)
				b = false
				break
			}
			data[k] = val
		}

		if b {
			type info struct {
				SQLName string
				Name    string
			}
			inf := &info{
				SQLName: fmt.Sprintf("%s:%s@(%s)", data["user"], data["passwd"], data["address"]),
				Name:    data["database"],
			}
			file, err := os.Create("./databases/mysql.go")
			if err != nil {
				panic(err)
			}
			defer file.Close()
			tmpl, _ := template.New("mysql").Parse(tpl)
			tmpl.Execute(file, inf)
			fmt.Println("gen success")
		}
	},
}

func init() {
	mysqlCmd.AddCommand(mysqlCreateCmd)
	mysqlCreateCmd.Flags().StringP("user", "u", "", "数据库用户名")
	mysqlCreateCmd.Flags().StringP("passwd", "p", "", "数据库密码")
	mysqlCreateCmd.Flags().StringP("address", "a", "", "数据库地址")
	mysqlCreateCmd.Flags().StringP("database", "d", "", "数据库名称")
}
