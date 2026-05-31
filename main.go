package main

import (
	"fmt"
	"html/template"
	"io"
	"net/http"
	"sync/atomic"

	"github.com/g13n4/LuteSentencePicker/application"
	"github.com/g13n4/LuteSentencePicker/db"
	"github.com/g13n4/LuteSentencePicker/state"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

type TemplateRenderer struct {
	templates *template.Template
}

func (t *TemplateRenderer) Render(c *echo.Context, w io.Writer, name string, data any) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func main() {
	stateSingleton := state.GetStateSingleton()
	var dbStatus atomic.Int64

	go loadDB(stateSingleton, &dbStatus)

	sp := db.NewStatusPool()
	var lastStepValue int64
	for lastStepValue < 6 {
		lastStepValue = dbStatus.Load()
		status, ok := sp.PopStatus(lastStepValue)
		if ok {
			fmt.Println(status.Message)
		}
	}

	fd := application.NewButtonsFrontend(stateSingleton.Pool)
	buttonsData, err := fd.GetIndexButtons()
	if err != nil {
		panic(err)
	}

	e := echo.New()

	e.Use(middleware.RequestLogger())

	e.Renderer = &TemplateRenderer{
		templates: template.Must(template.ParseGlob("src/*.html")),
	}

	e.GET("/", func(c *echo.Context) error {
		return c.Render(http.StatusOK, "index.html", buttonsData)
	})

	e.GET("/db-state", func(c *echo.Context) error {
		return c.Render(http.StatusOK, "index.html", nil)
	})

	if err := e.Start(":9999"); err != nil {
		panic(err)
	}
}
