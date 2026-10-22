package security

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
	"vigil/internal/ast"
)

type AuditEvent struct {
	Timestamp time.Time `json:"timestamp"`
	Severity  string    `json:"severity"`
	Rule      string    `json:"rule"`
	Message   string    `json:"message"`
}

type Engine struct {
	Streams     map[string]*ast.StreamStatement // Храним весь стейт потока (уровень + теги)
	Errors      []string
	AuditEvents []AuditEvent
	UserRole    string // Роль того, кто выполняет скрипт (для RBAC)
}

func New() *Engine {
	return &Engine{
		Streams:     make(map[string]*ast.StreamStatement),
		Errors:      []string{},
		AuditEvents: []AuditEvent{},
		UserRole:    "client", // По умолчанию клиент
	}
}

func (e *Engine) Check(program *ast.Program) {
	for _, stmt := range program.Statements {
		if s, ok := stmt.(*ast.StreamStatement); ok {
			e.Streams[s.Name.Value] = s
			e.logAudit("INFO", "STREAM_DECL", fmt.Sprintf("Объявлен поток %s (Уровень: %s, Теги: %v)", s.Name.Value, s.Level.Value, s.Tags))
		}
	}

	for _, stmt := range program.Statements {
		if r, ok := stmt.(*ast.ReactorStatement); ok {
			sourceStream := e.Streams[r.Stream.Value]
			if sourceStream != nil {
				e.checkBlock(r.Body, r.Variable.Value, sourceStream)
			}
		}
	}
}

func (e *Engine) checkBlock(block *ast.BlockStatement, varName string, sourceStream *ast.StreamStatement) {
	currentLevel := sourceStream.Level.Value
	isMasked := false

	for _, stmt := range block.Statements {
		switch s := stmt.(type) {
		case *ast.ExpressionStatement:
			if call, ok := s.Expression.(*ast.CallExpression); ok {
				if call.Function.TokenLiteral() == "mask" && len(call.Arguments) > 0 {
					if call.Arguments[0].TokenLiteral() == varName {
						isMasked = true
						currentLevel = "public" // Маска делает данные публичными
						e.logAudit("INFO", "MASK_APPLIED", fmt.Sprintf("Данные %s замаскированы", varName))
					}
				}
			}
		case *ast.TransmitStatement:
			targetLevel := s.Level.Value

			// ПРАВИЛО 1: Базовое (secure нельзя в public)
			if !e.isAllowed(currentLevel, targetLevel) {
				e.violation(fmt.Sprintf("SECURITY VIOLATION: Попытка передать [%s] данные в [%s] канал! (Строка %d)", currentLevel, targetLevel, s.Token.Line), "152-ФЗ Ст.7 / Базовый контур")
				continue
			}

			// ПРАВИЛО 2: Банковская тайна (ст. 857 ГК РФ). Даже маскированные данные нельзя в паблик.
			if targetLevel == "public" && e.hasTag(sourceStream, "bank_secret") {
				e.violation(fmt.Sprintf("LEGAL VIOLATION: Данные с тегом 'bank_secret' запрещено передавать в публичный канал! (Строка %d)", s.Token.Line), "ГК РФ Ст.857")
				continue
			}

			// ПРАВИЛО 3: PII (152-ФЗ). Если тег pii, в паблик можно только с маской.
			if targetLevel == "public" && e.hasTag(sourceStream, "pii") && !isMasked {
				e.violation(fmt.Sprintf("GDPR/152-FZ VIOLATION: Передача PII данных без маскировки! (Строка %d)", s.Token.Line), "152-ФЗ Ст.7")
				continue
			}

			// ПРАВИЛО 4: RBAC (Ролевая модель)
			if len(s.AllowedRoles) > 0 && !e.hasRole(s.AllowedRoles) {
				e.violation(fmt.Sprintf("RBAC VIOLATION: Роль '%s' не имеет доступа к каналу %s! (Строка %d)", e.UserRole, s.Channel.Value, s.Token.Line), "152-ФЗ Ст.9 (Доступ ролей)")
				continue
			}

			e.logAudit("INFO", "TRANSMIT_OK", fmt.Sprintf("Успешная передача в %s", s.Channel.Value))
		}
	}
}

func (e *Engine) isAllowed(source, target string) bool {
	if source == target {
		return true
	}
	if source == "secure" && (target == "internal" || target == "public") {
		return false
	}
	if source == "internal" && target == "public" {
		return false
	}
	return true
}

func (e *Engine) hasTag(stream *ast.StreamStatement, tag string) bool {
	for _, t := range stream.Tags {
		if t == tag {
			return true
		}
	}
	return false
}

func (e *Engine) hasRole(allowed []string) bool {
	for _, r := range allowed {
		if r == e.UserRole {
			return true
		}
	}
	return false
}

func (e *Engine) violation(msg string, rule string) {
	e.Errors = append(e.Errors, msg)
	e.logAudit("CRITICAL", rule, msg)

	// Делаем запись в файл в отдельной горутине (асинхронно), чтобы не блокировать сервер
	go func() {
		f, err := os.OpenFile("security_incidents.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err == nil {
			defer f.Close()
			f.WriteString(fmt.Sprintf("[%s] %s | %s\n", time.Now().Format(time.RFC3339), rule, msg))
		}
	}()
}

func (e *Engine) logAudit(severity, rule, msg string) {
	event := AuditEvent{time.Now(), severity, rule, msg}
	e.AuditEvents = append(e.AuditEvents, event)
}

// GenerateReport выводит JSON аудит-лог
func (e *Engine) GenerateReport() string {
	report, _ := json.MarshalIndent(e.AuditEvents, "", "  ")
	return string(report)
}
