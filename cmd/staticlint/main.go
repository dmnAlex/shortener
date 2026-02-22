/*
Package staticlint предоставляет инструмент для статического анализа кода (multichecker).

Механизм запуска:
1. Соберите анализатор:
   go build -o staticlint cmd/staticlint/main.go
2. Запустите проверку проекта:
   ./staticlint ./...

Список включенных анализаторов:
- printf: проверка соответствия форматов в функциях Printf.
- shadow: поиск затененных (shadowed) переменных.
- structtag: проверка валидности тегов в структурах.
- Все анализаторы класса SA пакета staticcheck.io (ошибки корректности).
- ST1005: cтиль сообщений об ошибках (staticcheck.io).
- unused: анализ неиспользуемого кода.
- bodyclose: проверка закрытия HTTP ответов.
- noexit: запрет os.Exit в main() пакета main (кастомный).
*/

package main

import (
	"strings"

	"github.com/dmnAlex/shortener/cmd/staticlint/noexit"
	"github.com/timakin/bodyclose/passes/bodyclose"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"honnef.co/go/tools/staticcheck"
	"honnef.co/go/tools/stylecheck"
	"honnef.co/go/tools/unused"
)

func main() {
	var analyzers []*analysis.Analyzer

	// Стандартные анализаторы golang.org/x/tools/go/analysis/passes
	analyzers = append(analyzers, printf.Analyzer, shadow.Analyzer, structtag.Analyzer)

	// Все анализаторы класса SA пакета staticcheck.io
	for _, v := range staticcheck.Analyzers {
		if strings.HasPrefix(v.Analyzer.Name, "SA") {
			analyzers = append(analyzers, v.Analyzer)
		}
	}

	// Один анализатор другого класса (ST1005 - стиль сообщений об ошибках)
	for _, v := range stylecheck.Analyzers {
		if v.Analyzer.Name == "ST1005" {
			analyzers = append(analyzers, v.Analyzer)
		}
	}

	// Два публичных анализатора
	analyzers = append(analyzers, unused.Analyzer.Analyzer)
	analyzers = append(analyzers, bodyclose.Analyzer)

	// Кастомный анализатор
	analyzers = append(analyzers, noexit.Analyzer)

	multichecker.Main(analyzers...)
}
