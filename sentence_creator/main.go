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
	mhs "github.com/g13n4/LuteSentencePicker/sentence_creator/creator/mhs"
	"github.com/g13n4/LuteSentencePicker/sentence_creator/creator/simple"
	"github.com/g13n4/LuteSentencePicker/sentence_creator/db"
	mw "github.com/g13n4/LuteSentencePicker/sentence_creator/middleware"
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
	mhsExec := mhs.NewExecutor(stateSingleton.Pool)
	simpleExec := simple.NewExecutor(stateSingleton.Pool)

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
	e.Use(mw.InjectQueryHelperMiddleware)

	e.Renderer = &TemplateRenderer{
		templates: template.Must(template.ParseGlob("src/*.html")),
	}

	e.GET("/", func(c *echo.Context) error {
		return c.Render(http.StatusOK, "index.html", buttonsData)
	})

	e.POST("/sentences", func(c *echo.Context) error {
		val := c.Get(mw.ContextKey)

		qh, ok := val.(*mw.QueryHelper)
		if !ok {
			return c.String(http.StatusBadRequest, "middleware doesn't work correctly")
		}

		timeNow := time.Now().Format("2006-01-02 15:04:05")
		fn := filepath.Join(sentencePath, "output-"+qh.String()+timeNow+".txt")

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

		if qh.UseMHS {
			err = mhsExec.GetSentences(context.Background(), file, *qh, 1000)
		} else {
			err = simpleExec.GetSentences(context.Background(), file, *qh, 1000)
		}
		if err != nil {
			panic(err)
		}

		return c.File(fn)
	})

	if err := e.Start(":9999"); err != nil {
		panic(err)
	}
}
