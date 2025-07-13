package main

import (
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// TestRunCmd проверяет корректную работу функции RunCmd с передачей переменных окружения и аргументов.
func TestRunCmd(t *testing.T) {
	// Создаём временную директорию для тестового скрипта
	dir := t.TempDir()
	scriptPath := filepath.Join(dir, "printenv.sh")

	// Тестовый shell-скрипт, который выводит значения переменных окружения и аргументы
	script := `#!/bin/sh
				echo "FOO=$FOO"
				echo "BAR=$BAR"
				echo "EMPTY=$EMPTY"
				echo "UNSET=$UNSET"
				echo "ARGS=$@"
			`

	// Пропускаем тест на Windows, так как shell-скрипты там не поддерживаются
	if runtime.GOOS == "windows" {
		t.Skip("Skipping on Windows")
	}

	// Записываем скрипт в файл и делаем его исполняемым
	if err := os.WriteFile(scriptPath, []byte(script), 0755); err != nil {
		t.Fatalf("Не удалось записать скрипт: %v", err)
	}

	// Формируем окружение для теста
	env := Environment{
		"FOO":   EnvValue{"foo", false},
		"BAR":   EnvValue{"bar", false},
		"EMPTY": EnvValue{"", false},
		"UNSET": EnvValue{"", true},
	}

	// Команда для запуска: /bin/sh printenv.sh arg1 arg2
	cmd := []string{"/bin/sh", scriptPath, "arg1", "arg2"}

	// Перехватываем вывод stdout и stderr, чтобы проверить результат выполнения
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stdout = w
	os.Stderr = w

	// Запускаем команду с нужным окружением
	code := RunCmd(cmd, env)

	// Закрываем pipe и проверяем ошибку закрытия (важно для линтера errcheck)
	if err := w.Close(); err != nil {
		t.Fatalf("ошибка при закрытии pipe: %v", err)
	}
	os.Stdout = oldStdout
	os.Stderr = oldStderr

	// Читаем весь вывод команды
	outBytes, _ := io.ReadAll(r)
	output := string(outBytes)

	// Проверяем код возврата
	if code != 0 {
		t.Errorf("Ожидался код возврата 0, получено %d", code)
	}
	if !strings.Contains(output, "FOO=foo") {
		t.Errorf("В выводе не найдено FOO: %q", output)
	}
	if !strings.Contains(output, "BAR=bar") {
		t.Errorf("В выводе не найдено BAR: %q", output)
	}
	if !strings.Contains(output, "EMPTY=") {
		t.Errorf("В выводе не найдено EMPTY: %q", output)
	}
	if !strings.Contains(output, "UNSET=") {
		t.Errorf("В выводе не найдено UNSET: %q", output)
	}
	if !strings.Contains(output, "ARGS=arg1 arg2") {
		t.Errorf("В выводе не найдено ARGS: %q", output)
	}
}
