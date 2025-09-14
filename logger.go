// gologen v.0.1.2 - simple logger in golang, async logging and graceful shutdown

// author: github.com/mysokolsky
// t.me/timeforpeople

package log

import (
	"bufio"
	"fmt"
	"os"
	// "os/signal"
	"strings"
	"sync"
	"sync/atomic"
	// "syscall"
	"time"
)

const (
	reset = "\033[0m" // сброс цветовых настроек шрифта и фона до стандартных

	// основные настройки стиля в консоли
	bold      = "\033[01m" // жирный
	italic    = "\033[03m" // наклонный
	underline = "\033[04m" // подчёркнутый

	// дополнительные настройки стиля в консоли
	inverse         = "\033[07m" // флип цветов: цвет текста <-> цвет фона
	fade            = "\033[02m" // тусклый (отдельно настройки яркого нет. Вместо этого используются прямые коды цветов: обычные (текст 30-37, фон 40-47), яркие (90-97, 100-107))
	strike          = "\033[09m" // зачёркнутый
	hide            = "\033[08m" // скрытый (цвет шрифта совпадает с цветом фона)
	blink           = "\033[05m" // мигающий (!не везде работает)
	doubleUnderline = "\033[21m" // двойное подчёркивание
	overline        = "\033[53m" // надчёркнутый

	// Используемые цвета
	extraRed   = "\033[38;05;196m" // светло-красный цвет текста из диапазона 256 цветов
	orange     = "\033[38;05;178m" // оранжевый цвет текста из диапазона 256 цветов
	extraWhite = "\033[97m"        // яркий белый цвет текста из стандартного диапазона
	gray       = "\033[38;05;244m" // серый цвет текста из диапазона 256 цветов
	lightGray  = "\033[38;05;250m" // светло-серый цвет текста из диапазона 256 цветов
	grayBG     = "\033[48;05;238m"
	darkRedBG  = "\033[48;05;88m"
	red        = "\033[31m"
)

// структура для записи атрибутов текста
type style struct {
	attrs []string
}

// уровни логирования
type loglevel uint

// уровни логирования
const (
	info loglevel = iota
	warn
	err
	fatal
)

// структура для хранения настроек для каждого уровня логирования
type levelconfig struct {
	timestamp style
	lvl_name  string
	lvl_style style
	message   style
}

// основная структура для хранения всех настроек уровней логгера
type logger struct {
	writer  *bufio.Writer            // буфер для строки
	ch      chan string              // канал для записи строк
	configs map[loglevel]levelconfig // мапа для хранения настроек всех уровней логов

	wg     sync.WaitGroup
	once   sync.Once // для одиночного закрытия, чтоб не было паники при повторном вызове
	closed atomic.Bool
}

// Основная функция-конструктор, которая создаёт объект логера и записывает нужные параметры в поля
func newLogger() *logger {

	l := &logger{ // ссылка на новый объект логирования

		writer: bufio.NewWriter(os.Stdout), // буфер для вывода строки

		// создаём буферизированный канал на 1000 строк для записи и вывода в консоль
		ch: make(chan string, 1000),

		// создаём пресет настройки стилей для вывода разных уровней логирования
		configs: map[loglevel]levelconfig{
			info: { // настройки стиля уровня info
				timestamp: style{attrs: []string{gray}}, // настройка текстового стиля для временной метки. В фигурных скобках после обозначения слайса строк через запятые укажите названия констант конкретной настройки текста. Например, ' []string { Red, Bold } ' будет означать красный цвет текста
				lvl_name:  "  INF  ",                    // название уровня, которое будет отображатьс на экране. Может быть любым
				lvl_style: style{attrs: []string{gray}}, // настройка текстового стиля отображения названия уровня
				message:   style{attrs: []string{gray}}, // настройка текстового стиля отображения сообщения
			},
			warn: { // настройки стиля уровня warning
				timestamp: style{attrs: []string{lightGray}},
				lvl_name:  "  WRN  ",
				lvl_style: style{attrs: []string{lightGray}},
				message:   style{attrs: []string{lightGray}},
			},
			err: { // настройки стиля уровня error
				timestamp: style{attrs: []string{}},
				lvl_name:  "  ERR  ",
				lvl_style: style{attrs: []string{orange, bold}},
				message:   style{attrs: []string{}},
			},
			fatal: { // настройки стиля уровня fatal
				timestamp: style{attrs: []string{extraWhite, grayBG}},
				lvl_name:  " FATAL ",
				lvl_style: style{attrs: []string{red, bold}},
				message:   style{attrs: []string{extraWhite, grayBG}},
			},
		},
	}

	// запускаем горутины
	l.run()

	return l // возвращаем ссылку на новый объект логирования

}

// создаём глобальный объект логера
var lg = newLogger()

// функция для вывода лога уровеня Info
func Info(format string, args ...interface{}) {
	lg.print(info, format, args...)
}

// вывод уровня Warn
func Warn(format string, args ...interface{}) {
	lg.print(warn, format, args...)
}

// вывод уровня Error
func Error(format string, args ...interface{}) {
	lg.print(err, format, args...)
}

// Fatalf выводится и завершает программу
func Fatalf(format string, args ...interface{}) {
	lg.print(fatal, format, args...)
	lg.shutdown()
	os.Exit(1)
}

func (log *logger) shutdown() {
	log.once.Do(func() {
		log.closed.Store(true) // ставим флаг
		close(log.ch)          // завершаем приём
		log.wg.Wait()          // ждём писателя
		log.writer.Flush()
	})
}

// возвращает строку со всеми атрибутами стиля для настройки текста в консоли
func (st *style) getFullAttrStyle() string {
	return strings.Join(st.attrs, "")
}

// // собираем полную строку лога и выводим в консоль
// func (log *logger) print(level loglevel, format string, args ...interface{}) {

// 	// подстрока временной метки (таймстамп)
// 	t := time.Now().Format("2006/01/02 15:04:05") + " "

// 	var msg string // переменная для сборки подстроки сообщения лога
// 	if len(args) == 0 {
// 		msg = format
// 	} else if strings.Contains(format, "%") {
// 		msg = fmt.Sprintf(format, args...)
// 	} else {
// 		msg = strings.TrimSuffix(fmt.Sprintln(append([]interface{}{format}, args...)...), "\n")
// 	}

// 	cfg := log.configs[level] // алиас (сокращённая временная переменная для удобства)

// 	// конечная строка суммируется из 3х подстрок
// 	// каждая подстрока тоже состоит из 3х составляющих:
// 	// 1) сначала идёт подподстрока настройки атрибутов стиля,
// 	// 2) потом текст,
// 	// 3) потом атрибут сброса стиля
// 	str := fmt.Sprintf("%s%s%s%s%s%s%s%s%s\n",
// 		cfg.timestamp.getFullAttrStyle(), t, reset,
// 		cfg.lvl_style.getFullAttrStyle(), cfg.lvl_name, reset,
// 		cfg.message.getFullAttrStyle(), " "+msg+" ", reset,
// 	)

// 	// если shutdown уже начался — просто дропаем
// 	select {
// 	case log.ch <- str:
// 	default:
// 	}
// }

func (log *logger) format(level loglevel, format string, args ...interface{}) string {
	t := time.Now().Format("2006/01/02 15:04:05") + " "

	var msg string
	if len(args) == 0 {
		msg = format
	} else if strings.Contains(format, "%") {
		msg = fmt.Sprintf(format, args...)
	} else {
		msg = strings.TrimSuffix(fmt.Sprintln(append([]interface{}{format}, args...)...), "\n")
	}

	cfg := log.configs[level]

	return fmt.Sprintf("%s%s%s%s%s%s%s%s%s\n",
		cfg.timestamp.getFullAttrStyle(), t, reset,
		cfg.lvl_style.getFullAttrStyle(), cfg.lvl_name, reset,
		cfg.message.getFullAttrStyle(), " "+msg+" ", reset,
	)
}

func (log *logger) print(level loglevel, format string, args ...interface{}) {
	str := log.format(level, format, args...)

	if log.closed.Load() {
		// аварийный принт в консоль — тот же формат
		fmt.Print(str)
		return
	}

	// обычный асинхронный режим
	select {
	case log.ch <- str:
	default: // если буфер переполнен — дропаем
	}
}

// run запускает горутины логгера
func (l *logger) run() {

	l.wg.Add(1)
	go func() {
		defer l.wg.Done()
		for msg := range l.ch {
			l.writer.WriteString(msg)
			l.writer.Flush()
		}
	}()

	// // graceful shutdown по сигналам
	// go func() {
	// 	sigc := make(chan os.Signal, 1)
	// 	signal.Notify(sigc, syscall.SIGINT, syscall.SIGTERM)
	// 	<-sigc
	// 	l.shutdown()
	// }()
}
