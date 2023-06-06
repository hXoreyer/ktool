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
	"regexp"
	"text/template"

	"github.com/spf13/cobra"
)

const tags = `package services

import (
	"{{.Mod}}/databases"
	"{{.Mod}}/models"
	"log"
)

type {{.Stname}} struct {
	{{range $k, $v := .Infos}}
	{{$v.Name}} {{$v.Type}} ` + "`json:\"{{$v.Tag}}\"" + "`" + `
	{{end}}
}

func init() {
	if !databases.Db.Migrator().HasTable(&models.{{.Stname}}{}) {
		err := databases.Db.Migrator().CreateTable(&models.{{.Stname}}{})
		if err != nil {
			log.Println(err)
			return
		}
	}
}`

type Info struct {
	Name string
	Type string
	Tag  string
}

type Res struct {
	Stname string
	Mod    string
	Infos  []Info
}

var res = &Res{}

// mysqlNewCmd represents the mysqlNew command
var mysqlNewCmd = &cobra.Command{
	Use:   "new",
	Short: "new sql",
	Long:  `创建sql的service文件`,
	Run: func(cmd *cobra.Command, args []string) {
		name, err := cmd.Flags().GetString("name")
		if err != nil || name == "" {
			fmt.Println("请输入文件名")
			return
		}
		readFileAndFirstParse(name)
	},
}

func init() {
	mysqlCmd.AddCommand(mysqlNewCmd)
	mysqlNewCmd.Flags().StringP("name", "n", "", "models里的文件名")
}

func readFileAndFirstParse(name string) {
	f, _ := os.ReadFile(fmt.Sprintf("./models/%s.go", name))

	// 定义正则表达式
	regex := regexp.MustCompile(`type\s+(.*)\s+struct\s*{([^}]*)}`)

	// 使用正则表达式进行匹配
	match := regex.FindStringSubmatch(string(f))
	if len(match) > 0 {
		res.Stname = match[1]
		res.Mod = getModName()
		res.Infos = make([]Info, 0)
		parseLoop(match[2])
		toFile(name)
	} else {
		fmt.Println("User struct definition not found")
	}
}

func parseLoop(str string) {
	regex := regexp.MustCompile(`(\w+)\s+(\w+)\s` + "`gorm:\"column:(\\w+).*\"" + "`([\\s\\S]*)")
	match := regex.FindStringSubmatch(str)
	if len(match) > 0 {
		res.Infos = append(res.Infos, Info{
			Name: match[1],
			Type: match[2],
			Tag:  match[3],
		})
		if len(match) >= 4 {
			parseLoop(match[4])
		} else {
			return
		}
	}
}

func toFile(name string) {
	if isExists(fmt.Sprintf("./services/%s.go", name)) {
		return
	}
	file, err := os.Create(fmt.Sprintf("./services/%s.go", name))
	if err != nil {
		panic(err)
	}
	defer file.Close()
	tmpl, _ := template.New("controller").Parse(tags)
	tmpl.Execute(file, res)
	if err != nil {
		fmt.Println("render template failed err =", err)
		return
	}
	fmt.Println("gen success")
}

func isExists(name string) bool {
	_, err := os.Stat(name)
	return !os.IsNotExist(err)
}
