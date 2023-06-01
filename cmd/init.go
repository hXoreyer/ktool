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
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"text/template"

	"github.com/spf13/cobra"
)

const mod = `module {{.Name}}

go {{.Version}}
`

const cors = `package middlewares

import "github.com/gin-gonic/gin"

func Cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}`

const configYaml = `groups:
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
      file: admin`

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "项目模块名称",
	Long:  `项目模块名称`,
	Run: func(cmd *cobra.Command, args []string) {
		name, err := cmd.Flags().GetString("name")
		if err != nil {
			fmt.Println("请输入项目名称")
			return
		}
		initMod(name)
		paths := []string{
			"controllers",
			"databases",
			"middlewares",
			"models",
			"routers",
			"services",
		}
		for _, v := range paths {
			err := os.Mkdir(fmt.Sprintf("./%s/%s", name, v), os.ModePerm)
			if err != nil {
				fmt.Printf("创建文件夹%s失败, err: %v\n", v, err)
			}
		}
		initCors(name)
		creatYaml(name)
		fmt.Println("gen success")
	},
}

func init() {
	rootCmd.AddCommand(initCmd)

	initCmd.Flags().StringP("name", "n", "kgin", "项目名称")
}

func initMod(name string) {
	os.Mkdir(name, os.ModePerm)
	type modType struct {
		Name    string
		Version string
	}
	m := &modType{
		Name: name,
	}
	v := runtime.Version()
	str := strings.Split(v[2:], ".")
	m.Version = fmt.Sprintf("%s.%s", str[0], str[1])

	file, err := os.Create(fmt.Sprintf("./%s/go.mod", name))
	if err != nil {
		panic(err)
	}
	defer file.Close()
	tmpl, _ := template.New("go.mod").Parse(mod)
	tmpl.Execute(file, m)
}

func initCors(name string) {
	file, err := os.Create(fmt.Sprintf("./%s/middlewares/cors.go", name))
	if err != nil {
		panic(err)
	}
	defer file.Close()
	_, err = io.WriteString(file, cors)
	if err != nil {
		fmt.Println("cors文件创建错误 err:", err)
		return
	}

}

func creatYaml(name string) {
	f, _ := os.Create(fmt.Sprintf("./%s/config.yaml", name))
	defer f.Close()

	bts := bufio.NewReader(bytes.NewBufferString(configYaml))
	io.Copy(f, bts)
}
