package handler

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"

	"example.com/Go/internal/transport/TLS"
)

func UploadFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	username := r.FormValue("username")
	if username == "" {
		http.Error(w, "Username is required", http.StatusBadRequest)
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

	// Get the file and the username
	file, handler, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Failed to retrieve file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Создание буфера для хранения данных формы
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Создание части формы с файлом
	part, err := writer.CreateFormFile("file", filepath.Base(handler.Filename))
	if err != nil {
		http.Error(w, "Failed to parse multipart form", http.StatusBadRequest)
		return
	}

	_, err = io.Copy(part, file)
	if err != nil {
		http.Error(w, "Failed to parse multipart form", http.StatusBadRequest)
		return
	}

	// Добавление поля username в форму
	err = writer.WriteField("username", username)
	if err != nil {
		http.Error(w, "Failed to parse multipart form", http.StatusBadRequest)
		return
	}

	// Закрытие writer для написания завершающих границ
	err = writer.Close()
	if err != nil {
		http.Error(w, "Failed to parse multipart form", http.StatusBadRequest)
		return
	}

	// Создание HTTP-запроса
	request, err := http.NewRequest("POST", "https://127.0.0.1:8443/upload-file", body)
	if err != nil {
		http.Error(w, "Error request", http.StatusUnauthorized)
		return
	}
	request.Header.Set("Content-Type", writer.FormDataContentType())
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
