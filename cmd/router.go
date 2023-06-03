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
	"io/ioutil"
	"os"
	"regexp"
	"text/template"

	"github.com/spf13/cobra"
	v2 "gopkg.in/yaml.v2"
)

const routerContent = `package routers

import (
	"{{.Name}}/controllers"
	"{{.Name}}/middlewares"

	"github.com/gin-gonic/gin"
)

func SetRouters() {
	r := gin.Default()
	r.Use(middlewares.Cors())
	{{range $k, $v := .Groups}}
	v{{$k}} := r.Group("{{$v.Group}}")
	{
		{{range $n, $m := $v.Methods}}
        v{{$k}}.{{$m.Method}}("{{$m.Path}}", controllers.{{$m.Function}})
        {{end}}
	}
	{{end}}
	r.Run(":5201")
}
`

const controllerContent = `package controllers

import (

	"github.com/gin-gonic/gin"
)
{{range $k, $v := .Functions}}
func {{$v.FunctionName}}(c *gin.Context) {
	c.JSON(200, gin.H{
		"code": 200,
		"msg":  "hello gin",
	})
}
{{end}}
`

const existContent = `
{{range $k, $v := .Functions}}
func {{$v.FunctionName}}(c *gin.Context) {
	c.JSON(200, gin.H{
		"code": 200,
		"msg":  "hello gin",
	})
}
{{end}}`

type info struct {
	Method   string `yaml:"method"`
	Path     string `yaml:"path"`
	Function string `yaml:"function"`
	File     string `yaml:"file"`
}

type group struct {
	Name  string `yaml:"name"`
	Infos []info `yaml:"infos"`
}

type Config struct {
	Groups []group `yaml:"groups"`
}

type Method struct {
	Method   string
	Path     string
	Function string
	File     string
}

type Group struct {
	Group   string
	Methods []Method
}

type logic struct {
	FunctionName string
}

// routerCmd represents the router command
var routerCmd = &cobra.Command{
	Use:   "router",
	Short: "router",
	Long:  `路由逻辑绑定`,
	Run: func(cmd *cobra.Command, args []string) {
		path, err := cmd.Flags().GetString("path")
		if err != nil {
			fmt.Println("解析yaml地址错误 err:", err)
			return
		}
		conf := loadYaml(path)
		parseFile(conf)
		fmt.Println("gen success")
	},
}

func init() {
	rootCmd.AddCommand(routerCmd)
	routerCmd.Flags().StringP("path", "p", "./config.yaml", "yaml文件地址")
}

func loadYaml(path string) *Config {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Printf("err: %v\n", err)
		return nil
	}

	conf := new(Config)
	if err := v2.Unmarshal(data, conf); err != nil {
		fmt.Printf("err: %v\n", err)
		return nil
	}
	return conf
}

func parseFile(conf *Config) {
	groups := make([]Group, len(conf.Groups))
	for k, v := range conf.Groups {
		funcs := make([]Method, len(v.Infos))
		for k, z := range v.Infos {
			funcs[k].Method = z.Method
			funcs[k].Path = z.Path
			funcs[k].Function = z.Function
			funcs[k].File = z.File
		}
		groups[k].Group = v.Name
		groups[k].Methods = funcs
	}
	go parseLogic(groups)
	var na string
	if na = getModName(); na == "" {
		fmt.Println("请先初始化项目")
		return
	}
	datas := struct {
		Name   string
		Groups []Group
	}{
		Name:   na,
		Groups: groups,
	}
	file, err := os.Create("./routers/router.go")
	if err != nil {
		panic(err)
	}
	defer file.Close()
	tmpl, _ := template.New("router").Parse(routerContent)
	tmpl.Execute(file, datas)
	if err != nil {
		fmt.Println("render template failed err =", err)
		return
	}
	initMain()
}

func getModName() string {
	f, err := os.Open("./go.mod")
	if err != nil {
		return ""
	}
	defer f.Close()

	r := bufio.NewReader(f)
	bt, _, err := r.ReadLine()
	if err == io.EOF {
		return ""
	}

	str := string(bt)

	return str[7:]
}

func parseLogic(groups []Group) {

	mp := make(map[string][]logic)
	for _, v := range groups {
		for _, m := range v.Methods {
			if _, ok := mp[m.File]; !ok {
				mp[m.File] = make([]logic, 0)
			}
			if functionIsExists(m.Function, m.File) {
				continue
			}
			mp[m.File] = append(mp[m.File], logic{FunctionName: m.Function})
		}
	}

	for k, v := range mp {
		go parse(k, v)
	}
}

func parse(name string, data []logic) {
	if len(data) == 0 {
		return
	}
	datas := struct {
		Functions []logic
	}{
		Functions: data,
	}
	tmpl := &template.Template{}
	if isExists(fmt.Sprintf("./controllers/%s.go", name)) {
		f, _ := os.ReadFile(fmt.Sprintf("./controllers/%s.go", name))
		content := string(f) + existContent
		tmpl, _ = template.New("controller").Parse(content)
	} else {
		tmpl, _ = template.New("controller").Parse(controllerContent)
	}
	file, err := os.Create(fmt.Sprintf("./controllers/%s.go", name))
	if err != nil {
		panic(err)
	}
	defer file.Close()
	tmpl.Execute(file, datas)
	if err != nil {
		fmt.Println("render template failed err =", err)
		return
	}
}

func initMain() {
	name := getModName()
	str := fmt.Sprintf(`package main

	import "%s/routers"
	
	func main() {
		routers.SetRouters()
	}
	`, name)
	_, err := os.Stat(fmt.Sprintf("./%s.go", name))
	if os.IsNotExist(err) {
		f, _ := os.Create(fmt.Sprintf("./%s.go", name))
		defer f.Close()

		bts := bufio.NewReader(bytes.NewBufferString(str))
		io.Copy(f, bts)
	}
}

func functionIsExists(name, path string) bool {
	_, err := os.Stat(fmt.Sprintf("./controllers/%s.go", path))
	if os.IsNotExist(err) {
		return false
	}

	f, _ := os.ReadFile(fmt.Sprintf("./controllers/%s.go", path))

	regex := regexp.MustCompile(`func\s+(\w+)\(c\s+\*gin.Context\)`)

	// 使用正则表达式进行匹配
	matches := regex.FindAllStringSubmatch(string(f), -1)
	if matches != nil {
		for _, match := range matches {
			functionName := match[1]
			if functionName == name {
				return true
			}
		}
	} else {
		return false
	}
	return false
}
