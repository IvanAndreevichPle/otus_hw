package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Environment — отображение переменных окружения: имя переменной -> значение и флаг NeedRemove.
type Environment map[string]EnvValue

// EnvValue позволяет различать пустые файлы и файлы с пустой первой строкой.
// Value      — значение переменной (первая строка файла, обработанная по правилам).
// NeedRemove — если true, переменная должна быть удалена из окружения.
type EnvValue struct {
	Value      string
	NeedRemove bool
}

// ReadDir читает директорию и возвращает map переменных окружения.
// Имя файла — имя переменной, первая строка файла — значение.
// Пустой файл помечается как NeedRemove, null-байты заменяются на \n.
func ReadDir(dir string) (Environment, error) {
	env := make(Environment)

	// Получаем список файлов в директории
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("не удалось прочитать директорию %s: %w", dir, err)
	}

	for _, entry := range entries {
		// Пропускаем вложенные директории
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		// Проверяем, что имя файла не содержит символ '=' (это запрещено для переменных окружения)
		if strings.Contains(name, "=") {
			return nil, fmt.Errorf("имя файла переменной окружения содержит '=': %s", name)
		}

		fullPath := filepath.Join(dir, name)
		// Открываем файл для чтения
		f, err := os.Open(fullPath)
		if err != nil {
			return nil, fmt.Errorf("не удалось открыть файл %s: %w", fullPath, err)
		}

		scanner := bufio.NewScanner(f)
		var val string
		var needRemove bool

		// Читаем первую строку файла
		if scanner.Scan() {
			line := scanner.Bytes()
			// Заменяем все null-байты на символ новой строки
			line = bytes.ReplaceAll(line, []byte{0x00}, []byte{'\n'})
			// Обрезаем пробелы и табуляцию справа
			val = strings.TrimRight(string(line), " \t")
			needRemove = false
		} else {
			// Если файл пустой, переменная должна быть удалена (NeedRemove = true)
			if err := scanner.Err(); err != nil {
				f.Close()
				return nil, fmt.Errorf("ошибка при чтении файла %s: %w", fullPath, err)
			}
			val = ""
			needRemove = true
		}

		// Закрываем файл и обрабатываем ошибку закрытия
		if err := f.Close(); err != nil {
			return nil, fmt.Errorf("не удалось закрыть файл %s: %w", fullPath, err)
		}

		// Добавляем переменную в итоговую карту окружения
		env[name] = EnvValue{Value: val, NeedRemove: needRemove}
	}

	// Возвращаем итоговую карту переменных окружения
	return env, nil
}
