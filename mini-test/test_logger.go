// mini-test/test_logger.go

// В консоли перейди в папку с этим файлом и запусти команду:
// go run test_logger.go

package main

import log "github.com/mysokolsky/gologen"
import "fmt"

func testLogger() {

log.Info("Привет, мир!")
log.Warn("Внимание!", "Предупреждение!")

// Пример с ошибкой
err := fmt.Errorf("Файл конфигурации не найден!")
log.Error("Критическая ошибка: %v Проверьте настройки!", err)

// имитация критической ошибки и выход из программы
log.Fatalf("После вывода этой строки произойдёт выход из программы!")
}

func main() {
testLogger()
}
