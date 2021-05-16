package hw10programoptimization

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

const bufferSize = 512

type UserEmail struct {
	Email string
}

type DomainStat map[string]int

func GetDomainStat(r io.Reader, domain string) (DomainStat, error) {
	result := DomainStat{}
	domainSuffix := "." + domain

	scanner := bufio.NewScanner(r)
	buffer := make([]byte, bufferSize)
	scanner.Buffer(buffer, 4*bufferSize)
	userEmail := &UserEmail{}
	for scanner.Scan() {
		*userEmail = UserEmail{}
		if err := json.Unmarshal(scanner.Bytes(), userEmail); err != nil {
			return nil, err
		}

		email := userEmail.Email
		if strings.HasSuffix(email, domainSuffix) {
			if idx := strings.IndexRune(email, '@'); idx >= 0 {
				result[strings.ToLower(email[idx+1:])]++
			} else {
				return nil, fmt.Errorf("user email: %s format is not valid (doesn't contain \"@\" symbol)", userEmail.Email)
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error while reading input: %w", err)
	}
	return result, nil
}
