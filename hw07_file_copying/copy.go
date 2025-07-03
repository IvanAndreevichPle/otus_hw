package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

var (
	ErrUnsupportedFile       = errors.New("unsupported file")
	ErrOffsetExceedsFileSize = errors.New("offset exceeds file size")
)

// Функция для определения оптимального размера буфера в зависимости от размера файла
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

func Copy(fromPath, toPath string, offset, limit int64) error {
	file, err := os.Open(fromPath)
	if err != nil {
		return err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return err
	}
	// Поддержка файла
	if !info.Mode().IsRegular() {
		return ErrUnsupportedFile
	}
	// Получаем размер файла
	fileSize := info.Size()
	if offset > fileSize {
		return ErrOffsetExceedsFileSize
	}
	// Проверка установки указателя в нужню позицию
	_, err = file.Seek(offset, io.SeekStart)
	if err != nil {
		return err
	}
	// Осталось записать
	remain := fileSize - offset
	toCopy := remain
	if limit > 0 && limit < remain {
		toCopy = limit
	}

	dst, err := os.Create(toPath)
	if err != nil {
		return err
	}
	defer dst.Close()

	bufSize := optimalBufferSize(toCopy)
	buf := make([]byte, bufSize)
	var copied int64
	start := time.Now()

	for copied < toCopy {
		readSize := bufSize
		left := toCopy - copied
		if left < int64(readSize) {
			readSize = int(left)
		}
		n, readErr := file.Read(buf[:readSize])
		if n > 0 {
			written, writeErr := dst.Write(buf[:n])
			if writeErr != nil {
				return writeErr
			}
			if written != n {
				return io.ErrShortWrite
			}
			copied += int64(written)
			printProgress(copied, toCopy, start)
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return readErr
		}
	}
	printProgress(toCopy, toCopy, start)
	fmt.Println()
	return nil
}

// Функция отображения прогресс-бара
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

	// Скорость и время
	elapsed := time.Since(start).Seconds()
	speed := float64(done) / 1024 / elapsed // KB/sec
	var eta string
	if speed > 0 {
		etaSec := float64(total-done) / 1024 / speed
		eta = fmt.Sprintf(" ETA: %2.0fs", etaSec)
	} else {
		eta = ""
	}

	fmt.Printf("\r%s %3d%% \033[36m%6.1f KB/s%s\033[0m", bar, percent, speed, eta)
}
