package events

import (
	"html/template"
	"os"
)

func Create(appName string) {
	err := os.Mkdir(appName+"/events", 0755)
	if err != nil {
		panic(err)
	}

	createdFile, err := os.Create(appName + "/events/example.go")
	if err != nil {
		panic(err)
	}

	t, err := template.ParseFiles("internal/cli/templates/events/base.tmpl")
	if err != nil {
		panic(err)
	}

	err = t.Execute(createdFile, nil)
	if err != nil {
		panic(err)
	}
}
