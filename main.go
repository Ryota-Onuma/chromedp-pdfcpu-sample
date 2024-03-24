package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"text/template"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
)

type Handler struct{}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	val := map[string]string{
		"あいさつ":     "こんにちは",
		"Greeting": "Hello",
	}

	// ファイルじゃなくてもいいが、HTMLを読み込む
	rawHTMLBytes, err := os.ReadFile("sample.html")
	if err != nil {
		log.Fatal(err)
	}

	tmp, err := template.New("template").Parse(string(rawHTMLBytes))
	if err != nil {
		log.Fatal(err)
	}

	buf := new(bytes.Buffer)
	if err := tmp.Execute(buf, val); err != nil {
		log.Fatal(err)
	}

	fmt.Fprintf(w, string(buf.String()))
}

func main() {
	ctx := context.Background()
	var opts []chromedp.ExecAllocatorOption
	opts = append(opts,
		chromedp.Flag("headless", true),
	)
	allocCtx, _ := chromedp.NewExecAllocator(ctx, opts...)
	chromdpCtx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	srv := &http.Server{
		Addr:    ":9999",
		Handler: nil,
	}

	http.Handle("/pdf/generate", &Handler{})
	go srv.ListenAndServe()

	defer func(ctx context.Context) {
		if err := srv.Shutdown(chromdpCtx); err != nil {
			log.Fatal(err)
		}
	}(chromdpCtx)

	if err := generatePDF(chromdpCtx); err != nil {
		fmt.Println(err)
	}
}

func generatePDF(chromdpCtx context.Context) error {
	if err := chromedp.Run(chromdpCtx,
		chromedp.Navigate("http://localhost:9999/pdf/generate"),
		chromedp.ActionFunc(func(ctx context.Context) error {
			buf, _, err := page.PrintToPDF().
				WithPreferCSSPageSize(true).
				WithPrintBackground(true).
				WithScale(1).
				WithPaperWidth(8.3).
				WithPaperHeight(11.7).
				Do(ctx)
			if err != nil {
				return err
			}
			r := bytes.NewReader(buf)
			out, err := os.Create("A4.pdf")
			if err != nil {
				return err
			}
			defer out.Close()

			nup := model.DefaultNUpConfig()
			nup.PageSize = "A4P"
			nup.PageGrid = false
			nup.Border = false

			if err := pdfcpu.ParseNUpGridDefinition(1, 1, nup); err != nil {
				return err
			}
			conf := model.NewDefaultConfiguration()
			api.NUp(r, out, nil, nil, nup, conf)
			return nil
		}),
	); err != nil {
		return err
	}
	return nil
}
