package utils

import "github.com/vektah/gqlparser/v2/gqlerror"

// Функция для создания ошибки с кодом
func NewGraphQLError(message string, code string) *gqlerror.Error {
	return &gqlerror.Error{
		Message: message,
		Extensions: map[string]interface{}{
			"code": code, // сюда можно положить HTTP-код
		},
	}
}
