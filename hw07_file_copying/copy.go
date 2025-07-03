package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

// Объявление ошибок для специфических случаев
var (
	ErrUnsupportedFile       = errors.New("unsupported file")
	ErrOffsetExceedsFileSize = errors.New("offset exceeds file size")
)

// optimalBufferSize возвращает оптимальный размер буфера для копирования
// в зависимости от размера файла. Это помогает ускорить копирование
// и не расходовать лишнюю память.
func optimalBufferSize(fileSize int64) int {
	switch {
	case fileSize < 128*1024: // до 128 КБ
		return 4 * 1024
	case fileSize < 1*1024*1024: // до 1 МБ
		return 64 * 1024
	case fileSize < 100*1024*1024: // до 100 МБ
		return 256 * 1024
	case fileSize < 1*1024*1024*1024: // до 1 ГБ
		return 512 * 1024
	default:
		return 1024 * 1024 // 1 МБ
	}
}

// Copy копирует данные из файла fromPath в файл toPath с поддержкой смещения (offset)
// и лимита (limit). В процессе копирования отображается прогресс-бар.
func Copy(fromPath, toPath string, offset, limit int64) error {
	// Открываем исходный файл для чтения
	file, err := os.Open(fromPath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Получаем информацию о файле (размер, тип)
	info, err := file.Stat()
	if err != nil {
		return err
	}
	// Проверяем, что это обычный файл, а не, например, директория или спецфайл
	if !info.Mode().IsRegular() {
		return ErrUnsupportedFile
	}
	// Получаем размер файла
	fileSize := info.Size()
	// Проверяем, что смещение не превышает размер файла
	if offset > fileSize {
		return ErrOffsetExceedsFileSize
	}
	// Устанавливаем указатель чтения на нужную позицию
	_, err = file.Seek(offset, io.SeekStart)
	if err != nil {
		return err
	}
	// Определяем, сколько байт нужно скопировать
	remain := fileSize - offset
	toCopy := remain
	if limit > 0 && limit < remain {
		toCopy = limit
	}

	// Открываем файл назначения для записи
	dst, err := os.Create(toPath)
	if err != nil {
		return err
	}
	defer dst.Close()

	// Выбираем оптимальный размер буфера для копирования
	bufSize := optimalBufferSize(toCopy)
	buf := make([]byte, bufSize)
	var copied int64
	start := time.Now() // Запоминаем время начала для расчёта скорости и ETA

	// Основной цикл копирования
	for copied < toCopy {
		readSize := bufSize
		left := toCopy - copied
		// Если осталось скопировать меньше, чем размер буфера, уменьшаем размер чтения
		if left < int64(readSize) {
			readSize = int(left)
		}
		// Читаем данные из исходного файла
		n, readErr := file.Read(buf[:readSize])
		if n > 0 {
			// Пишем прочитанные данные в файл назначения
			written, writeErr := dst.Write(buf[:n])
			if writeErr != nil {
				return writeErr
			}
			// Проверяем, что записано столько же байт, сколько прочитано
			if written != n {
				return io.ErrShortWrite
			}
			copied += int64(written)
			// Отображаем прогресс-бар
			printProgress(copied, toCopy, start)
		}
		// Если достигнут конец файла, выходим из цикла
		if readErr == io.EOF {
			break
		}
		// Если возникла другая ошибка при чтении, возвращаем её
		if readErr != nil {
			return readErr
		}
	}
	// После завершения копирования гарантируем отображение 100% прогресса
	printProgress(toCopy, toCopy, start)
	fmt.Println()
	return nil
}

// printProgress отображает прогресс-бар в консоли, а также скорость копирования и ETA.
// Использует ANSI-цвета для красивого отображения.
func printProgress(done, total int64, start time.Time) {
	if total == 0 {
		fmt.Print("\r\033[32m[████████████████████████████████████████] 100%\033[0m")
		return
	}
	barWidth := 30
	pct := float64(done) / float64(total)
	filled := int(pct * float64(barWidth))
	if filled > barWidth {
		filled = barWidth
	}
	// Цвета: зелёный — выполнено, синий — индикатор, серый — не выполнено
	green := "\033[32m"
	blue := "\033[34m"
	gray := "\033[90m"
	reset := "\033[0m"

	bar := green + strings.Repeat("█", filled)
	if filled < barWidth {
		bar += blue + ">" + gray + strings.Repeat("░", barWidth-filled-1)
	}
	bar += reset
	bar = "[" + bar + "]"

	percent := int(pct * 100)
	if percent > 100 {
		percent = 100
	}

	// Вычисляем скорость копирования и ETA
	elapsed := time.Since(start).Seconds()
	speed := float64(done) / 1024 / elapsed // KB/sec
	var eta string
	if speed > 0 {
		etaSec := float64(total-done) / 1024 / speed
		eta = fmt.Sprintf(" ETA: %2.0fs", etaSec)
	} else {
		eta = ""
	}

	// Выводим прогресс-бар, процент, скорость и ETA
	fmt.Printf("\r%s %3d%% \033[36m%6.1f KB/s%s\033[0m", bar, percent, speed, eta)
}
