package mock

import "strings"

type CompanyInfo struct {
	INN       string `json:"inn"`
	Name      string `json:"name"`
	Director  string `json:"director"`
	Status    string `json:"status"`
	RiskLevel string `json:"risk_level"` // green, yellow, red
}

// GetCompanyInfo имитирует запрос к сервису ФНС "Прозрачный бизнес"
func GetCompanyInfo(inn string) CompanyInfo {
	// Для примера: если ИНН начинается на 77 - компания супер, если на 11 - рискованная
	if strings.HasPrefix(inn, "77") {
		return CompanyInfo{
			INN: inn, Name: "ООО Ромашка", Director: "Иванов И.И.",
			Status: "Действующее", RiskLevel: "green",
		}
	} else if strings.HasPrefix(inn, "11") {
		return CompanyInfo{
			INN: inn, Name: "ООО Риск", Director: "Петров П.П.",
			Status: "В процессе ликвидации", RiskLevel: "red",
		}
	}
	// Дефолтный ответ
	return CompanyInfo{
		INN: inn, Name: "ООО Неизвестная Компания", Director: "Неизвестно",
		Status: "Действующее", RiskLevel: "yellow",
	}
}
