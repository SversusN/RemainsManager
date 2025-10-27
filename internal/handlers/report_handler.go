// package handlers
package handlers

import (
	"context"
	_ "embed"
	"fmt"
	"github.com/chromedp/cdproto/page"
	"html/template"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"RemainsManager/internal/models"
	"RemainsManager/internal/services"
	"github.com/chromedp/chromedp"
)

//go:embed templates/report_offer.html
var reportTemplate string

type ReportHandler struct {
	reportService *services.ReportService
	chromeCtx     context.Context // контекст для chromedp (можно кэшировать)
}

func NewReportHandler(reportService *services.ReportService) *ReportHandler {
	// Создаём общий контекст для chromedp (можно переиспользовать)
	ctx, _ := chromedp.NewContext(context.Background())
	return &ReportHandler{
		reportService: reportService,
		chromeCtx:     ctx,
	}
}

func (h *ReportHandler) GenerateOfferReport(w http.ResponseWriter, r *http.Request) {
	offerIDStr := r.URL.Query().Get("id")
	if offerIDStr == "" {
		http.Error(w, "missing 'id' query param", http.StatusBadRequest)
		return
	}

	offerID, err := strconv.ParseInt(offerIDStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid 'id' format", http.StatusBadRequest)
		return
	}

	report, err := h.reportService.BuildReport(r.Context(), offerID)
	if err != nil {
		http.Error(w, "failed to build report: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if report == nil {
		http.Error(w, "offer not found", http.StatusNotFound)
		return
	}

	// Сортируем группы по ключу (получателю)
	sortedGroups := make(map[string][]models.OfferItemReport)
	keys := make([]string, 0, len(report.Groups))
	for k := range report.Groups {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		sortedGroups[k] = report.Groups[k]
	}
	report.Groups = sortedGroups

	// Рендерим HTML
	tmpl := template.Must(template.New("report").Parse(reportTemplate))
	var htmlStr string
	buf := &strings.Builder{}
	err = tmpl.Execute(buf, report)
	if err != nil {
		http.Error(w, "template error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	htmlStr = buf.String()
	// 1. Создаём контекст с таймаутом (рекомендуется!)
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel() // ← обязательно!

	// 2. Создаём контекст chromedp на основе ctx
	ctx, cancel = chromedp.NewContext(ctx)
	defer cancel()
	// Генерируем PDF через chromedp
	var pdfBuf []byte
	err = chromedp.Run(ctx,
		chromedp.Navigate("data:text/html,"+url.PathEscape(htmlStr)),
		chromedp.WaitVisible("body", chromedp.ByQuery),
		chromedp.ActionFunc(func(ctx context.Context) error {
			var err error
			pdfBuf, _, err = page.PrintToPDF().WithLandscape(false).Do(ctx) // альбомная ориентация
			return err
		}),
	)
	if err != nil {
		http.Error(w, "PDF generation failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`inline; filename="offer_%d.pdf"`, offerID))
	w.Write(pdfBuf)
}
