package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadDir(t *testing.T) {
	// Создаём временную директорию для тестовых файлов
	dir := t.TempDir()

	// Описываем набор тестовых случаев:
	// name     — имя файла (переменной окружения)
	// content  — содержимое файла
	// expected — ожидаемое значение EnvValue после чтения
	tests := []struct {
		name     string
		content  []byte
		expected EnvValue
	}{
		// Обычное значение с пробелами и табуляцией в конце
		{"FOO", []byte("foo value   \t"), EnvValue{"foo value", false}},
		// Значение с null-байтом, который должен быть заменён на \n
		{"BAR", []byte("bar\x00baz   "), EnvValue{"bar\nbaz", false}},
		// Пустой файл — должен помечаться для удаления (NeedRemove = true)
		{"EMPTY", []byte(""), EnvValue{"", true}},
		// Многострочный файл — берём только первую строку
		{"MULTILINE", []byte("first line\nsecond line"), EnvValue{"first line", false}},
	}

	// Создаём тестовые файлы в директории
	for _, tc := range tests {
		err := os.WriteFile(filepath.Join(dir, tc.name), tc.content, 0644)
		if err != nil {
			t.Fatalf("Не удалось создать тестовый файл: %v", err)
		}
	}

	// Читаем директорию и получаем переменные окружения
	env, err := ReadDir(dir)
	if err != nil {
		t.Fatalf("ReadDir завершился с ошибкой: %v", err)
	}

	// Проверяем, что все переменные считаны правильно
	for _, tc := range tests {
		val, ok := env[tc.name]
		if !ok {
			t.Errorf("Ожидалась переменная %q, но она не найдена", tc.name)
			continue
		}
		if val != tc.expected {
			t.Errorf("Для %q: получено %+v, ожидалось %+v", tc.name, val, tc.expected)
		}
	}

	// Проверяем обработку некорректного имени файла (содержит '=')
	badname := "BAD=NAME"
	err = os.WriteFile(filepath.Join(dir, badname), []byte("bad"), 0644)
	if err != nil {
		t.Fatalf("Не удалось создать файл с плохим именем: %v", err)
	}
	_, err = ReadDir(dir)
	if err == nil {
		t.Errorf("Ожидалась ошибка для имени файла с '=', но ошибки не было")
	}
}
