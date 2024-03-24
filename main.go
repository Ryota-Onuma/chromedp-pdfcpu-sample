package main

import (
	"bytes"
	"context"
	"errors"
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
		// chromiumのsandboxは、ホストマシン上への変更や機密情報へのアクセスを防ぐためのもの
		// CAP_SYS_ADMINのcapabilityをつけないと、namespaceを使ってchromium sandboxプロセスを分離することができない
		// CAP_SYS_ADMINをつけると、影響範囲が広いので、つけない方法を考える
		// 隔離してホストマシンを守るための機能なので、別の方法で隔離できてればOKなはず
		// Docker for MacやRancher Desktopはdockerdが動くVMがセキュリティ境界になっているため、--no-sandboxあっても問題ない
		//    SEE: https://docs.docker.com/desktop/mac/permission-requirements/#containers-running-as-root-within-the-linux-vm
		// 	  SEE: https://docs.rancherdesktop.io/references/architecture
		// OrbStackは厳密にはどうやら独立したVMで動かしてはいないようだが、セキュリティに自信ありそうなので大丈夫でしょうおそらく
		// 	  SEE: https://docs.orbstack.dev/architecture#security
		chromedp.Flag(("no-sandbox"), true),
	)
	allocCtx, _ := chromedp.NewExecAllocator(ctx, opts...)
	chromdpCtx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	srv := &http.Server{
		Addr:    ":9999",
		Handler: nil,
	}

	http.Handle("/pdf/generate", &Handler{})
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				log.Fatal(err)
			}
		}
	}()

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
