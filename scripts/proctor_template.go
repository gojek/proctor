package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

func parse(templateFile string, outFile string, version string, sha string) {
	t, err := template.ParseFiles(templateFile)
	if err != nil {
		log.Print(err)
		return
	}

	f, err := os.Create(outFile)
	if err != nil {
		log.Println("create file: ", err)
		return
	}

	config := map[string]string{
		"Tag": version,
		"SHA": sha,
	}

	err = t.Execute(f, config)
	if err != nil {
		log.Print("execute: ", err)
		return
	}
	f.Close()
}

func main() {
	if len(os.Args) != 3 {
		log.Print("Required arguments not passed. Usage `go run main.go <version> <sha>`")
		return
	}
	currentDir, err := filepath.Abs("./")
	if err != nil {
		log.Print(err)
		return
	}
	outFile := fmt.Sprintf("%s/scripts/proctor.rb", currentDir)
	templateFile := fmt.Sprintf("%s/scripts/proctor.rb.tpl", currentDir)
	_, err = os.Stat(outFile)
	if os.IsExist(err) {
		err = os.Remove("proctor.rb")
		if err != nil {
			log.Print("Unable to remove file proctor.rb")
		}
	}
	parse(templateFile, outFile, strings.Replace(os.Args[1], "v", "", 1), os.Args[2])
}
