package handler

import (
	"io"
	"net/http"
	"strings"

	"example.com/Go/internal/transport/TLS"
)

func DeleteFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	username := r.URL.Query().Get("username")
	filename := r.URL.Query().Get("filename")
	if username == "" || filename == "" {
		http.Error(w, "Отсутствуют параметры username или filename", http.StatusBadRequest)
		return
	}

	// Извлечение токена из заголовка Authorization
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Authorization header is required", http.StatusUnauthorized)
		return
	}

	// Проверка формата заголовка и извлечение токена
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
		return
	}
	token := parts[1]

	// Создание HTTP-запроса
	request, err := http.NewRequest("DELETE", "https://127.0.0.1:8443/delete-file?filename="+filename+"&username="+username, nil)
	if err != nil {
		http.Error(w, "Error request", http.StatusUnauthorized)
		return
	}

	request.Header.Set("Authorization", "Bearer "+token)

	var client http.Client

	client = TLS.Client()

	response, err := client.Do(request)
	if err != nil {
		http.Error(w, "Error client", http.StatusUnauthorized)
		return
	}
	defer response.Body.Close()

	// Чтение тела ответа
	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		http.Error(w, "Error reading response body", http.StatusInternalServerError)
		return
	}

	// Отправка содержимого ответа клиенту
	_, err = w.Write(responseBody)
	if err != nil {
		http.Error(w, "Error writing response to client", http.StatusInternalServerError)
		return
	}
}
