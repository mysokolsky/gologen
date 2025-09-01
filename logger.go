// gologen v.0.1.0 - simple logger in golang,

// author: github.com/mysokolsky
// t.me/timeforpeople

package log

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"
)

const (
	Reset = "\033[0m" // сброс цветовых настроек шрифта и фона до стандартных

	// основные настройки стиля в консоли
	Bold      = "\033[01m" // жирный
	Italic    = "\033[03m" // наклонный
	Underline = "\033[04m" // подчёркнутый

	// дополнительные настройки стиля в консоли
	Inverse         = "\033[07m" // флип цветов: цвет текста <-> цвет фона
	Fade            = "\033[02m" // тусклый (отдельно настройки яркого нет. Вместо этого используются прямые коды цветов: обычные (текст 30-37, фон 40-47), яркие (90-97, 100-107))
	Strike          = "\033[09m" // зачёркнутый
	Hide            = "\033[08m" // скрытый (цвет шрифта совпадает с цветом фона)
	Blink           = "\033[05m" // мигающий (!не везде работает)
	DoubleUnderline = "\033[21m" // двойное подчёркивание
	Overline        = "\033[53m" // надчёркнутый

	// Используемые цвета
	ExtraRed   = "\033[38;05;196m" // светло-красный цвет текста из диапазона 256 цветов
	Orange     = "\033[38;05;178m" // оранжевый цвет текста из диапазона 256 цветов
	ExtraWhite = "\033[97m"        // яркий белый цвет текста из стандартного диапазона
	Gray       = "\033[38;05;244m" // серый цвет текста из диапазона 256 цветов
	LightGray  = "\033[38;05;250m" // светло-серый цвет текста из диапазона 256 цветов
	GrayBG     = "\033[48;05;238m"
	DarkRedBG  = "\033[48;05;88m"
	Red        = "\033[31m"
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
	*bufio.Writer                          // указатель на буферизованный писальщик
	configs       map[loglevel]levelconfig // мапа для хранения настроек всех уровней логов
}

// Основная функция-конструктор, которая создаёт объект логера и записывает нужные параметры в поля
func newLogger() *logger {

	return &logger{ // возвращаем ссылку на новый объект логирования

		// создаём буфер для записи и вывода в консоль
		Writer: bufio.NewWriter(os.Stdout),

		// создаём пресет настройки стилей для вывода разных уровней логирования
		configs: map[loglevel]levelconfig{
			info: { // настройки стиля уровня info
				timestamp: style{attrs: []string{Gray}}, // настройка текстового стиля для временной метки. В фигурных скобках после обозначения слайса строк через запятые укажите названия констант конкретной настройки текста. Например, ' []string { Red, Bold } ' будет означать красный цвет текста
				lvl_name:  "  INF  ",                    // название уровня, которое будет отображатьс на экране. Может быть любым
				lvl_style: style{attrs: []string{Gray}}, // настройка текстового стиля отображения названия уровня
				message:   style{attrs: []string{Gray}}, // настройка текстового стиля отображения сообщения
			},
			warn: { // настройки стиля уровня warning
				timestamp: style{attrs: []string{LightGray}},
				lvl_name:  "  WRN  ",
				lvl_style: style{attrs: []string{LightGray}},
				message:   style{attrs: []string{LightGray}},
			},
			err: { // настройки стиля уровня error
				timestamp: style{attrs: []string{}},
				lvl_name:  "  ERR  ",
				lvl_style: style{attrs: []string{Orange, Bold}},
				message:   style{attrs: []string{}},
			},
			fatal: { // настройки стиля уровня fatal
				timestamp: style{attrs: []string{ExtraWhite, GrayBG}},
				lvl_name:  " FATAL ",
				lvl_style: style{attrs: []string{Red, Bold}},
				message:   style{attrs: []string{ExtraWhite, GrayBG}},
			},
		},
	}

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
	os.Exit(1)
}

// возвращает строку со всеми атрибутами стиля для настройки текста в консоли
func (st *style) getFullAttrStyle() string {
	return strings.Join(st.attrs, "")
}

// собираем полную строку лога и выводим в консоль
func (log *logger) print(level loglevel, format string, args ...interface{}) {

	// подстрока временной метки (таймстамп)
	t := time.Now().Format("2006/01/02 15:04:05") + " "

	var msg string // переменная для сборки подстроки сообщения лога
	if len(args) == 0 {
		msg = format
	} else if strings.Contains(format, "%") {
		msg = fmt.Sprintf(format, args...)
	} else {
		msg = strings.TrimSuffix(fmt.Sprintln(append([]interface{}{format}, args...)...), "\n")
	}

	cfg := log.configs[level] // алиас (сокращённая временная переменная для удобства)

	// конечная строка суммируется из 3х подстрок
	// каждая подстрока тоже состоит из 3х составляющих:
	// 1) сначала идёт подподстрока настройки атрибутов стиля,
	// 2) потом текст,
	// 3) потом атрибут сброса стиля
	str := fmt.Sprintf("%s%s%s%s%s%s%s%s%s\n",
		cfg.timestamp.getFullAttrStyle(), t, Reset,
		cfg.lvl_style.getFullAttrStyle(), cfg.lvl_name, Reset,
		cfg.message.getFullAttrStyle(), " "+msg+" ", Reset,
	)

	// запись конечной строки в буфер
	log.WriteString(str)
	// вывод накопившихся строк из буфера.
	// (Вообще, по хорошему вывод в проде надо делать в асинхронной функции, скажем каждые 100 миллисеккунд)
	log.Flush()
}
