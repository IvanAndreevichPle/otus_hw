package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// RunCmd запускает команду с аргументами (cmd) с переменными окружения из env.
// Возвращает код завершения процесса.
func RunCmd(cmd []string, env Environment) (returnCode int) {
	// Проверяем, что команда не пуста
	if len(cmd) == 0 {
		return 111
	}
	command := exec.Command(cmd[0], cmd[1:]...)

	// Формируем карту переменных окружения на основе текущего окружения процесса
	envMap := make(map[string]string)
	for _, e := range os.Environ() {
		if i := strings.IndexByte(e, '='); i >= 0 {
			envMap[e[:i]] = e[i+1:]
		}
	}

	// Модифицируем окружение в соответствии с env:
	// - если NeedRemove, удаляем переменную
	// - иначе устанавливаем новое значение
	for k, v := range env {
		if v.NeedRemove {
			delete(envMap, k)
		} else {
			envMap[k] = v.Value
		}
	}

	// Преобразуем карту обратно в срез строк вида "ключ=значение"
	var envList []string
	for k, v := range envMap {
		envList = append(envList, k+"="+v)
	}
	command.Env = envList

	// Перенаправляем стандартные потоки ввода/вывода/ошибок на текущий процесс
	command.Stdin = os.Stdin
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr

	// Запускаем команду и обрабатываем возможные ошибки
	if err := command.Run(); err != nil {
		// Если это ошибка завершения процесса, возвращаем соответствующий код выхода
		if exitErr, ok := err.(*exec.ExitError); ok {
			// Для UNIX-систем: получаем код выхода через ExitStatus()
			if status, ok := exitErr.Sys().(interface{ ExitStatus() int }); ok {
				return status.ExitStatus()
			}
			// Для других случаев: используем ExitCode()
			return exitErr.ExitCode()
		}
		// Для других ошибок выводим сообщение и возвращаем 111
		fmt.Fprintf(os.Stderr, "RunCmd error: %v\n", err)
		return 111
	}
	// Если команда завершилась успешно, возвращаем 0
	return 0
}
