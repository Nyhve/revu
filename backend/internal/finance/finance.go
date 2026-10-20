// Package finance implements income/expense transactions and statistics (/api/finance).
package finance

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/Nyhve/revu/backend/internal/app"
	"github.com/Nyhve/revu/backend/internal/auth"
	"github.com/Nyhve/revu/backend/internal/decimal"
	"github.com/Nyhve/revu/backend/internal/httpx"
	"github.com/Nyhve/revu/backend/internal/store"
)

var (
	types      = []string{"INCOME", "EXPENSE"}
	categories = []string{
		"SALARY", "FREELANCE", "OTHER_INCOME",
		"FOOD", "TRANSPORT", "HOUSING", "HEALTH", "ENTERTAINMENT",
		"SHOPPING", "EDUCATION", "OTHER_EXPENSE",
	}
	monthNames = [12]string{"Jan", "Feb", "Mär", "Apr", "Mai", "Jun", "Jul", "Aug", "Sep", "Okt", "Nov", "Dez"}
	minAmount  = decimal.MustParse("0.01")
)

type CreateRequest struct {
	Amount      *decimal.Decimal `json:"amount"`
	Description *string          `json:"description"`
	Category    *string          `json:"category"`
	Type        *string          `json:"type"`
	Date        *httpx.Date      `json:"date"`
}

func (r CreateRequest) Validate() error {
	var v httpx.Validator
	if v.Required("amount", r.Amount != nil, "Amount must not be null") {
		v.DecimalMin("amount", r.Amount, minAmount, "Amount must be greater than 0")
		v.DecimalFits("amount", r.Amount, 12, 2, "Amount must not exceed 9999999999.99")
	}
	v.NotBlank("description", r.Description, "Description must not be blank")
	v.MaxLen("description", r.Description, 255, "Description must not exceed 255 characters")
	if v.Required("category", r.Category != nil, "Category must not be null") {
		v.OneOf("category", r.Category, categories)
	}
	if v.Required("type", r.Type != nil, "Type must not be null") {
		v.OneOf("type", r.Type, types)
	}
	v.Required("date", r.Date != nil, "Date must not be null")
	return v.Err()
}

type TransactionResponse struct {
	ID          int64               `json:"id"`
	Amount      decimal.Decimal     `json:"amount"`
	Description string              `json:"description"`
	Category    string              `json:"category"`
	Type        string              `json:"type"`
	Date        httpx.Date          `json:"date"`
	CreatedAt   httpx.LocalDateTime `json:"createdAt"`
}

type MonthlyStatsResponse struct {
	Year          int             `json:"year"`
	Month         int             `json:"month"`
	TotalIncome   decimal.Decimal `json:"totalIncome"`
	TotalExpenses decimal.Decimal `json:"totalExpenses"`
	Balance       decimal.Decimal `json:"balance"`
}

type MonthlySummary struct {
	Year       int             `json:"year"`
	Month      int             `json:"month"`
	MonthLabel string          `json:"monthLabel"`
	Income     decimal.Decimal `json:"income"`
	Expenses   decimal.Decimal `json:"expenses"`
}

func toResponse(t store.Transaction) TransactionResponse {
	return TransactionResponse{
		ID:          t.ID,
		Amount:      decimal.FromNumeric(t.Amount),
		Description: t.Description,
		Category:    t.Category,
		Type:        t.Type,
		Date:        httpx.Date(t.Date),
		CreatedAt:   httpx.LocalDateTime(t.CreatedAt),
	}
}

type Handler struct{ *app.Deps }

func Register(r *gin.RouterGroup, d *app.Deps) {
	h := Handler{d}
	g := r.Group("/finance")
	g.GET("", httpx.H(h.list))
	g.POST("", httpx.H(h.create))
	g.DELETE("/:id", httpx.H(h.delete))
	g.GET("/stats/:year/:month", httpx.H(h.monthlyStats))
	g.GET("/summary", httpx.H(h.summary))
}

func (h Handler) list(c *gin.Context) error {
	rows, err := h.Q.ListTransactions(c, auth.UserID(c))
	if err != nil {
		return err
	}
	out := make([]TransactionResponse, 0, len(rows))
	for _, t := range rows {
		out = append(out, toResponse(t))
	}
	return httpx.OK(c, out)
}

func (h Handler) create(c *gin.Context) error {
	var req CreateRequest
	if err := httpx.Bind(c, &req); err != nil {
		return err
	}
	t, err := h.Q.CreateTransaction(c, store.CreateTransactionParams{
		UserID:      auth.UserID(c),
		Amount:      req.Amount.Numeric(),
		Description: *req.Description,
		Category:    *req.Category,
		Type:        *req.Type,
		Date:        req.Date.Time(),
		CreatedAt:   h.Clock.Now(),
	})
	if err != nil {
		return err
	}
	return httpx.Created(c, toResponse(t))
}

func (h Handler) delete(c *gin.Context) error {
	id, err := httpx.PathID(c, "id")
	if err != nil {
		return err
	}
	n, err := h.Q.DeleteTransaction(c, store.DeleteTransactionParams{ID: id, UserID: auth.UserID(c)})
	if err != nil {
		return err
	}
	if n == 0 {
		return httpx.NotFound("Transaction", id)
	}
	return httpx.NoContent(c)
}

func (h Handler) monthlyStats(c *gin.Context) error {
	year, err := httpx.PathInt(c, "year")
	if err != nil {
		return err
	}
	month, err := httpx.PathInt(c, "month")
	if err != nil {
		return err
	}
	if year < 1 || year > 9999 || month < 1 || month > 12 {
		return httpx.BadRequest("Invalid year or month: %d/%d", year, month)
	}
	from, to := monthBounds(year, time.Month(month))
	sums, err := h.Q.SumTransactionsBetween(c, store.SumTransactionsBetweenParams{
		UserID: auth.UserID(c), FromDate: from, ToDate: to,
	})
	if err != nil {
		return err
	}
	income := decimal.FromNumeric(sums.Income)
	expenses := decimal.FromNumeric(sums.Expenses)
	return httpx.OK(c, MonthlyStatsResponse{
		Year:          year,
		Month:         month,
		TotalIncome:   income,
		TotalExpenses: expenses,
		Balance:       income.Sub(expenses),
	})
}

// summary returns income/expenses for the last ?months=6 calendar months
// (oldest first, current month last).
func (h Handler) summary(c *gin.Context) error {
	months, err := httpx.QueryInt(c, "months", 6, 0, 120)
	if err != nil {
		return err
	}
	out := make([]MonthlySummary, 0, months)
	if months == 0 {
		return httpx.OK(c, out)
	}
	today := h.Clock.Today()
	current := time.Date(today.Year(), today.Month(), 1, 0, 0, 0, 0, time.UTC)
	first := current.AddDate(0, -(months - 1), 0)
	_, last := monthBounds(current.Year(), current.Month())

	rows, err := h.Q.MonthlyTransactionSums(c, store.MonthlyTransactionSumsParams{
		UserID: auth.UserID(c), FromDate: first, ToDate: last,
	})
	if err != nil {
		return err
	}
	type key struct{ y, m int }
	byMonth := make(map[key]store.MonthlyTransactionSumsRow, len(rows))
	for _, r := range rows {
		byMonth[key{int(r.Year), int(r.Month)}] = r
	}
	for i := 0; i < months; i++ {
		m := first.AddDate(0, i, 0)
		s := MonthlySummary{
			Year:       m.Year(),
			Month:      int(m.Month()),
			MonthLabel: fmt.Sprintf("%s %d", monthNames[m.Month()-1], m.Year()),
			Income:     decimal.Zero,
			Expenses:   decimal.Zero,
		}
		if r, ok := byMonth[key{m.Year(), int(m.Month())}]; ok {
			s.Income = decimal.FromNumeric(r.Income)
			s.Expenses = decimal.FromNumeric(r.Expenses)
		}
		out = append(out, s)
	}
	return httpx.OK(c, out)
}

func monthBounds(year int, month time.Month) (time.Time, time.Time) {
	from := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
	return from, from.AddDate(0, 1, -1)
}
