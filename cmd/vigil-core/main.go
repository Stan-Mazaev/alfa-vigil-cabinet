package main

import (
	"fmt"
	"os"
	"vigil/internal/api" // <-- ДОБАВИЛИ
	"vigil/internal/ast"
	"vigil/internal/lexer"
	"vigil/internal/parser"
	"vigil/internal/security"
)

func main() {
	// Если запустили без аргументов, или с аргументом "server" - запускаем веб-сервер
	if len(os.Args) < 2 || os.Args[1] == "server" {
		api.StartServer()
		return
	}

	// Дальше старый код для консольного режима...
	filePath := os.Args[1]
	content, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Printf("Ошибка чтения файла: %v\n", err)
		os.Exit(1)
	}

	// 1. Лексер
	l := lexer.New(string(content))

	// 2. Парсер
	p := parser.New(l)
	program := p.ParseProgram()

	// Проверяем на ошибки синтаксиса
	if len(p.Errors()) != 0 {
		fmt.Println("❌ Ошибки синтаксиса Vigil:")
		for _, msg := range p.Errors() {
			fmt.Println("  -", msg)
		}
		os.Exit(1)
	}

	// 3. ПРОВЕРКА БЕЗОПАСНОСТИ
	secEngine := security.New()
	secEngine.Check(program)

	if len(secEngine.Errors) > 0 {
		fmt.Println("🛑 [CRITICAL] СБОРКА ЗАБЛОКИРОВАНА! Обнаружены нарушения безопасности:")
		for _, msg := range secEngine.Errors {
			fmt.Println("  ❌", msg)
		}
		// Выводим JSON аудит
		fmt.Println("\n📋 AUDIT LOG (JSON):")
		fmt.Println(secEngine.GenerateReport())
		os.Exit(1) // Завершаем с ошибкой, как это делает компилятор
	}

	fmt.Println("✅ Скрипт Vigil безопасен! Структура (AST):")
	printNode(program, "")
}

// Рекурсивная функция для красивого вывода дерева
func printNode(node ast.Node, indent string) {
	switch n := node.(type) {
	case *ast.Program:
		for _, stmt := range n.Statements {
			printNode(stmt, indent)
		}
	case *ast.StreamStatement:
		fmt.Printf("%s📦 Stream: %s (Уровень доступа: %s)\n", indent, n.Name.Value, n.Level.Value)
		if len(n.Tags) > 0 {
			fmt.Printf("%s   🏷️  Теги: %v\n", indent, n.Tags)
		}
	case *ast.ReactorStatement:
		fmt.Printf("%s⚡ Реактор на поток: %s (как переменная: %s)\n", indent, n.Stream.Value, n.Variable.Value)
		printNode(n.Body, indent+"    ")
	case *ast.BlockStatement:
		for _, stmt := range n.Statements {
			printNode(stmt, indent)
		}
	case *ast.ExpressionStatement:
		printNode(n.Expression, indent)
	case *ast.CallExpression:
		fmt.Printf("%s🛡️ Вызов функции: %s(", indent, n.Function.TokenLiteral())
		for i, arg := range n.Arguments {
			if i > 0 {
				fmt.Print(", ")
			}
			fmt.Print(arg.TokenLiteral())
		}
		fmt.Println(")")
	case *ast.Identifier:
		fmt.Printf("%s🔹 Идентификатор: %s\n", indent, n.Value)
	case *ast.TransmitStatement:
		fmt.Printf("%s🚀 Передача данных: %s в %s (Уровень: %s)\n", indent, n.Data.TokenLiteral(), n.Channel.Value, n.Level.Value)
		if len(n.AllowedRoles) > 0 {
			fmt.Printf("%s   👥 Разрешённые роли: %v\n", indent, n.AllowedRoles)
		}
	}
}
