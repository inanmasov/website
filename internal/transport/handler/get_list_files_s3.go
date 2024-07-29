package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"example.com/Go/internal/transport/TLS"
)

func GetListFiles(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	username := r.URL.Query().Get("username")
	if username == "" {
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
	request, err := http.NewRequest("GET", "https://127.0.0.1:8443/list-files?username="+username, nil)
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

	// Чтение ответа от сервера
	respBody, err := io.ReadAll(response.Body)
	if err != nil {
		http.Error(w, "error", http.StatusUnauthorized)
		return
	}
	// Преобразование массива байтов в строку
	respStr := string(respBody)

	// Отправка строки в формате JSON
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(map[string]string{"response": respStr})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
