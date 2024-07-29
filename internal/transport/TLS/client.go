package TLS

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

func Client() http.Client {
	dir, err := os.Getwd()
	if err != nil {
		fmt.Println("Ошибка получения рабочего каталога:", err)
		return http.Client{}
	}

	// Загрузка клиентского сертификата и ключа
	cert, err := tls.LoadX509KeyPair(dir+"/certificate/client.crt", dir+"/certificate/client.key")
	if err != nil {
		log.Fatalf("Ошибка загрузки клиентского сертификата и ключа: %v", err)
	}

	// Загрузка корневого сертификата CA
	caCert, err := ioutil.ReadFile(dir + "/certificate/ca.crt")
	if err != nil {
		log.Fatalf("Ошибка чтения корневого сертификата: %v", err)
	}

	// Создание нового пула доверенных корневых сертификатов
	caCertPool := x509.NewCertPool()
	if !caCertPool.AppendCertsFromPEM(caCert) {
		log.Fatalf("Не удалось добавить корневой сертификат в пул")
	}

	// Создание конфигурации TLS
	tlsConfig := &tls.Config{
		Certificates:       []tls.Certificate{cert}, // Если используется клиентская аутентификация
		RootCAs:            caCertPool,
		InsecureSkipVerify: true, // Не отключать проверку сертификата
	}

	// Создание HTTP-клиента с кастомным транспортом
	transport := &http.Transport{
		TLSClientConfig: tlsConfig,
	}

	client := &http.Client{
		Transport: transport,
	}
	return *client
}
