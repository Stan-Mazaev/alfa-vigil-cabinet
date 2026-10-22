package scoring

import (
	"fmt"
	"strings"
)

type Result struct {
	Approved        bool    `json:"approved"`
	ProductName     string  `json:"product_name"`
	ApprovedAmount  int     `json:"approved_amount"`
	SuggestedAmount int     `json:"suggested_amount,omitempty"`
	InterestRate    float64 `json:"interest_rate"`
	Reason          string  `json:"reason"`
}

func Evaluate(inn string, requestedAmount int, purpose string) Result {
	// 1. Базовая проверка (Регион)
	if !strings.HasPrefix(inn, "77") && !strings.HasPrefix(inn, "50") {
		return Result{
			Approved: false,
			Reason:   "Отказ: Высокий риск. Регион регистрации не входит в зону покрытия.",
		}
	}

	// 2. Логика от цели
	switch purpose {
	case "working_capital":
		if requestedAmount > 5000000 {
			return Result{
				Approved:        false,
				Reason:          "Для оборотных средств лимит 5 млн ₽.",
				SuggestedAmount: 5000000,
			}
		}
		return Result{
			Approved:       true,
			ProductName:    "Оборотный кредит",
			ApprovedAmount: requestedAmount,
			InterestRate:   21.0, // <-- ОБНОВЛЕНО (было 16.5)
			Reason:         "Одобрено. Подходит для пополнения оборотных средств.",
		}

	case "equipment":
		if requestedAmount > 15000000 {
			return Result{
				Approved:        false,
				Reason:          "Для оборудования лимит 15 млн ₽.",
				SuggestedAmount: 15000000,
			}
		}
		return Result{
			Approved:       true,
			ProductName:    "Инвестиционный кредит",
			ApprovedAmount: requestedAmount,
			InterestRate:   18.5, // <-- ОБНОВЛЕНО (было 14.0)
			Reason:         "Одобрено. Требуется залог приобретаемого оборудования.",
		}

	case "cash_gap":
		if requestedAmount > 1000000 {
			return Result{
				Approved:        false,
				Reason:          "Овердрафт доступен на сумму до 1 млн ₽.",
				SuggestedAmount: 1000000,
			}
		}
		return Result{
			Approved:       true,
			ProductName:    "Овердрафт",
			ApprovedAmount: requestedAmount,
			InterestRate:   24.0, // <-- ОБНОВЛЕНО (было 19.0)
			Reason:         "Одобрено. Лимит установлен сроком на 6 месяцев.",
		}
	}

	return Result{Approved: false, Reason: "Неизвестная цель финансирования."}
}

func FormatRate(r float64) string {
	return fmt.Sprintf("%.1f", r)
}
