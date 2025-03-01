// auth/consumer.go
package auth

import (
	"errors"
	"net/http"
)

func GetConsumer(header http.Header) (string, error) {
	consumer := header.Get("x-consumer-username")
	if consumer == "" {
		consumer = header.Get("x-application-id")
		if consumer == "" {
			return "", errors.New("no authentication found")
		}
	}
	return consumer, nil
}
