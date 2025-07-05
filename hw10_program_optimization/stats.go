package hw10programoptimization

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strings"
)

type User struct {
	ID       int
	Name     string
	Username string
	Email    string
	Phone    string
	Password string
	Address  string
}

type DomainStat map[string]int

func GetDomainStat(r io.Reader, domain string) (DomainStat, error) {
	result := make(DomainStat)
	scanner := bufio.NewScanner(r)
	domain = strings.ToLower(domain)

	for scanner.Scan() {
		line := scanner.Bytes()
		var email string
		if err := extractEmail(line, &email); err != nil {
			continue // пропускаем строки с ошибками
		}
		email = strings.ToLower(email)
		at := strings.LastIndex(email, "@")
		if at == -1 {
			continue
		}
		dom := email[at+1:]
		if strings.HasSuffix(dom, "."+domain) {
			dot := strings.LastIndex(dom, ".")
			if dot == -1 {
				continue
			}
			domName := dom[:dot]
			result[domName+"."+domain]++
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan error: %w", err)
	}
	return result, nil
}

// Быстрый парсер email из json-строки
func extractEmail(line []byte, email *string) error {
	key := []byte(`"Email":"`)
	idx := bytes.Index(line, key)
	if idx == -1 {
		return fmt.Errorf("no email")
	}
	start := idx + len(key)
	end := bytes.IndexByte(line[start:], '"')
	if end == -1 {
		return fmt.Errorf("bad email")
	}
	*email = string(line[start : start+end])
	return nil
}
