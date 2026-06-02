package main

import (
	"context"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync/atomic"
	"time"

	"github.com/g13n4/LuteSentencePicker/sentence_creator/application"
	"github.com/g13n4/LuteSentencePicker/sentence_creator/db"
	"github.com/g13n4/LuteSentencePicker/sentence_creator/mhs"
	"github.com/g13n4/LuteSentencePicker/sentence_creator/repository"
	"github.com/g13n4/LuteSentencePicker/sentence_creator/state"
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

	sentencePath := os.Getenv("OUTPUT_FOLDER")
	if sentencePath == "" {
		panic("no shared folder found")
	}

	sp := db.NewStatusPool()
	mhsRepo := repository.NewMHSRepository(stateSingleton.Pool)

	var lastStepValue int64
	for lastStepValue < 6 {
		lastStepValue = dbStatus.Load()
		status, ok := sp.PopStatus(lastStepValue)
		if ok {
			log.Println(status.Message)
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

	e.POST("/sentences", func(c *echo.Context) error {
		qh, err := mhs.NewQueryHelper(c)
		if err != nil {
			return c.String(http.StatusBadRequest, err.Error())
		}

		fn := filepath.Join(sentencePath, "output-"+qh.String()+time.Now().String()+".txt")

		file, err := os.Create(fn)
		defer func() {
			err := file.Close()
			if err != nil {
				panic(err)
			}
		}()

		if err != nil {
			panic("have no access to output file")
		}

		err = mhsRepo.GetSentences(context.Background(), file, *qh, 5, 1000)
		if err != nil {
			panic(err)
		}

		return c.File(fn)
	})

	if err := e.Start(":9999"); err != nil {
		panic(err)
	}
}
