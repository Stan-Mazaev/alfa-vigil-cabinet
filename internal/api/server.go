package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"vigil/internal/lexer"
	"vigil/internal/mock"
	"vigil/internal/parser"
	"vigil/internal/scoring"
	"vigil/internal/security"
)

type LoanRequest struct {
	Company   string `json:"company"`
	INN       string `json:"inn"`
	Amount    int    `json:"amount"`
	Purpose   string `json:"purpose"`
	Role      string `json:"role"` // <-- НОВОЕ (client или analyst)
	DebugLeak bool   `json:"debug_leak"`
}

type LoanResponse struct {
	Amount          int                   `json:"amount"`
	ProductName     string                `json:"product_name,omitempty"`
	ApprovedAmount  int                   `json:"approved_amount,omitempty"`
	SuggestedAmount int                   `json:"suggested_amount,omitempty"`
	Rate            float64               `json:"rate,omitempty"`
	Error           string                `json:"error,omitempty"`
	Reason          string                `json:"reason,omitempty"`
	AIAdvice        string                `json:"ai_advice,omitempty"` // <-- Оставлено поле, но больше не заполняется
	AuditLog        []security.AuditEvent `json:"audit_log,omitempty"`
	VigilCode       string                `json:"vigil_code,omitempty"`
	ASTJSON         string                `json:"ast_json,omitempty"` // <-- НОВОЕ
}

func StartServer() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "web/index.html")
	})
	http.HandleFunc("/api/apply", handleApply)
	http.HandleFunc("/api/company-info", handleCompanyInfo)
	http.HandleFunc("/api/pre-approved", handlePreApproved)

	fmt.Println("🚀 Сервер Умного кредитного кабинета запущен на http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}

func handleCompanyInfo(w http.ResponseWriter, r *http.Request) {
	inn := r.URL.Query().Get("inn")
	if len(inn) != 10 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(mock.GetCompanyInfo(inn))
}

// handlePreApproved возвращает предварительно одобренный лимит и ставку по ИНН.
// Используется фронтендом для превью «вам доступно до N ₽ под X%» ещё до подачи заявки.
func handlePreApproved(w http.ResponseWriter, r *http.Request) {
	inn := r.URL.Query().Get("inn")
	w.Header().Set("Content-Type", "application/json")

	// Мок-логика: если ИНН московский (77), даем 3 млн, иначе 1 млн
	limit := 1000000
	rate := 17.0
	if len(inn) >= 2 && inn[:2] == "77" {
		limit = 3000000
		rate = 15.0
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"limit": limit,
		"rate":  rate,
	})
}

func handleApply(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req LoanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(LoanResponse{Error: "Неверный формат данных"})
		return
	}

	var vigilScript string
	if req.DebugLeak {
		vigilScript = `
stream applicant_data [level: secure]
stream public_analytics [level: public]
on applicant_data as client {
    transmit(client) to public_analytics [level: public]
}`
	} else {
		// И для клиента, и для аналитика генерируем одинаковый безопасный скрипт
		// Разница будет только в том, как движок проверит роль (allowed_roles)
		vigilScript = `
stream applicant_data [level: secure]
stream scoring_result [level: internal]
stream public_analytics [level: public]
on applicant_data as client {
    mask(client)
    transmit(client) to scoring_result [level: internal]
}
on scoring_result as score {
    mask(score)
    transmit(score) to public_analytics [level: public, allowed_roles: analyst]
}`
	}

	l := lexer.New(vigilScript)
	p := parser.New(l)
	program := p.ParseProgram()

	// Сериализуем AST в JSON для отображения на фронтенде
	astData, _ := json.MarshalIndent(program, "", "  ")
	astString := string(astData)

	secEngine := security.New()
	secEngine.UserRole = req.Role // <-- Устанавливаем роль (client или analyst)
	secEngine.Check(program)

	if len(secEngine.Errors) > 0 {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(LoanResponse{
			Error:     secEngine.Errors[0],
			AuditLog:  secEngine.AuditEvents,
			VigilCode: vigilScript,
			ASTJSON:   astString, // <-- ДОБАВИТЬ
		})
		return
	}

	scoreResult := scoring.Evaluate(req.INN, req.Amount, req.Purpose)

	if !scoreResult.Approved {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(LoanResponse{
			Error:           "Заявка отклонена банком",
			Reason:          scoreResult.Reason,
			SuggestedAmount: scoreResult.SuggestedAmount,
			AuditLog:        secEngine.AuditEvents,
			VigilCode:       vigilScript,
			ASTJSON:         astString, // <-- ДОБАВИТЬ (для консистентности)
		})
		return
	}

	json.NewEncoder(w).Encode(LoanResponse{
		Amount:         req.Amount,
		ProductName:    scoreResult.ProductName,
		ApprovedAmount: scoreResult.ApprovedAmount,
		Rate:           scoreResult.InterestRate,
		Reason:         scoreResult.Reason,
		AuditLog:       secEngine.AuditEvents,
		VigilCode:      vigilScript, // <-- Отдаем код
		ASTJSON:        astString,   // <-- ДОБАВИТЬ
	})
}
