package handler

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"example.com/Go/internal/database"
	"example.com/Go/internal/transport/TLS"
	_ "github.com/lib/pq"
)

type Person struct {
	Id       int    `json:"id"`
	Login    string `json:"login"`
	Password string `json:"password"`
}

// generateToken генерирует случайный токен заданной длины.
func generateToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func getPeople(w http.ResponseWriter, r *http.Request) {
	var person Person
	err := json.NewDecoder(r.Body).Decode(&person)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	db, err := database.Initialize()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	result, err := db.DB.Query("select id from Person where login = $1", person.Login)
	if err != nil {
		panic(err)
	}
	defer result.Close()

	if result.Next() {
		var id int
		if err := result.Scan(&id); err != nil {
			http.Error(w, "Error scanning result", http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(id)
	} else {
		fmt.Fprintf(w, "Person not found")
	}
}

func postSign(w http.ResponseWriter, r *http.Request) {
	var person Person
	err := json.NewDecoder(r.Body).Decode(&person)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	hasher := sha256.New()
	hasher.Write([]byte(person.Password))
	hashInBytes := hasher.Sum(nil)
	hashPsw := hex.EncodeToString(hashInBytes)

	db, err := database.Initialize()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	result, err := db.DB.Query("select password from Person where login = $1", person.Login)
	if err != nil {
		panic(err)
	}
	defer result.Close()

	var passwordHash string
	flag := 0
	// Перемещаем курсор к первой строке результата
	if result.Next() {
		// Сканируем значение поля "password" в переменную
		err := result.Scan(&passwordHash)
		if err != nil {
			panic(err)
		}
	} else {
		flag = 1
	}

	if hashPsw != passwordHash {
		flag = 1
	}

	token, err := generateToken(32)
	if err != nil {
		log.Fatalf("Error generating token: %v", err)
	}

	db, ok := database.Initialize()
	if ok != nil {
		log.Fatal(ok)
	}
	defer db.Close()

	_, ok = db.DB.Exec("update Person set token = $1 where login = $2", token, person.Login)
	if ok != nil {
		fmt.Fprintf(w, "Error create person: "+fmt.Sprint(ok))
		return
	}

	response := make(map[string]interface{})

	switch flag {
	case 0:
		response["message"] = token
		response["code"] = 0
	case 1:
		response["message"] = "Wrong login or password"
		response["code"] = 1
	default:
		response["message"] = "Unknown"
		response["code"] = -1
	}

	jsonResponse, err := json.Marshal(response)
	if err != nil {
		http.Error(w, "Ошибка при сериализации JSON", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResponse)
}

func updatePeople(w http.ResponseWriter, r *http.Request) {
	var person Person
	err := json.NewDecoder(r.Body).Decode(&person)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if person.Login != "" && person.Password != "" {
		hasher := sha256.New()
		hasher.Write([]byte(person.Password))
		hashInBytes := hasher.Sum(nil)
		hashPsw := hex.EncodeToString(hashInBytes)

		db, err := database.Initialize()
		if err != nil {
			log.Fatal(err)
		}
		defer db.Close()

		result, err := db.DB.Exec("update Person set login = $1, password = $2 where id = $3", person.Login, hashPsw, person.Id)
		if err != nil {
			fmt.Fprintf(w, "Error update person: "+fmt.Sprint(err))
			return
		}
		num, err := result.RowsAffected()
		if err != nil {
			fmt.Fprintf(w, "Error update person")
			return
		}

		fmt.Fprintf(w, fmt.Sprint(num)+" line updated")
	} else if person.Login != "" {
		db, err := database.Initialize()
		if err != nil {
			log.Fatal(err)
		}
		defer db.Close()

		result, err := db.DB.Exec("update Person set login = $1 where id = $2", person.Login, person.Id)
		if err != nil {
			fmt.Fprintf(w, "Error update person: "+fmt.Sprint(err))
			return
		}
		num, err := result.RowsAffected()
		if err != nil {
			fmt.Fprintf(w, "Error update person")
			return
		}

		fmt.Fprintf(w, fmt.Sprint(num)+" line updated")
	} else if person.Password != "" {
		hasher := sha256.New()
		hasher.Write([]byte(person.Password))
		hashInBytes := hasher.Sum(nil)
		hashPsw := hex.EncodeToString(hashInBytes)

		db, err := database.Initialize()
		if err != nil {
			log.Fatal(err)
		}
		defer db.Close()

		result, err := db.DB.Exec("update Person set password = $1 where id = $2", hashPsw, person.Id)
		if err != nil {
			fmt.Fprintf(w, "Error update person: "+fmt.Sprint(err))
			return
		}
		num, err := result.RowsAffected()
		if err != nil {
			fmt.Fprintf(w, "Error update person")
			return
		}

		fmt.Fprintf(w, fmt.Sprint(num)+" line updated")
	} else {
		fmt.Fprintf(w, "Error input")
	}
}

func deletePeople(w http.ResponseWriter, r *http.Request) {
	var person Person
	err := json.NewDecoder(r.Body).Decode(&person)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	db, err := database.Initialize()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	result, err := db.DB.Exec("delete from Person where login = $1", person.Login)
	if err != nil {
		fmt.Fprintf(w, "Error delete person: "+fmt.Sprint(err))
		return
	}

	num, err := result.RowsAffected()
	if err != nil {
		fmt.Fprintf(w, "Error deleted person")
	}

	fmt.Fprintf(w, fmt.Sprint(num)+" line deleted")
}

func Sign(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		getPeople(w, r)
	case http.MethodPost:
		postSign(w, r)
	case http.MethodPatch:
		updatePeople(w, r)
	case http.MethodDelete:
		deletePeople(w, r)
	default:
		http.Error(w, "Invalid http method", http.StatusMethodNotAllowed)
	}
}

func Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Структура для хранения данных
	var data struct {
		Login                  string `json:"login"`
		Password               string `json:"password"`
		FullName               string `json:"fullName"`
		Address                string `json:"address"`
		PassportSeriesNumber   string `json:"passportSeriesNumber"`
		PassportIssueDate      string `json:"passportIssueDate"`
		PassportIssueCode      string `json:"passportIssueCode"`
		PassportIssueAuthority string `json:"passportIssueAuthority"`
		ConsentText            string `json:"consentText"`
	}

	// Декодируем JSON из тела запроса в структуру data
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		fmt.Println(err.Error())
		http.Error(w, "Could not decode JSON", http.StatusBadRequest)
		//return
	}

	token, err := generateToken(32)
	if err != nil {
		log.Fatalf("Error generating token: %v", err)
	}

	hasher := sha256.New()
	hasher.Write([]byte(data.Password))
	hashInBytes := hasher.Sum(nil)
	hashPsw := hex.EncodeToString(hashInBytes)

	db, err := database.Initialize()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	result, err := db.DB.Exec(`insert into Person (token, login, password, full_name, address, 
		passport_series_number, passport_issue_date, passport_issue_code,
		passport_issue_authority, consent_text) values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		token, data.Login, hashPsw, data.FullName, data.Address, data.PassportSeriesNumber,
		data.PassportIssueDate, data.PassportIssueCode, data.PassportIssueAuthority,
		data.ConsentText)
	if err != nil {
		fmt.Println(err)
		fmt.Fprintf(w, "Error create person: "+fmt.Sprint(err))
		return
	}

	_, err = result.RowsAffected()
	if err != nil {
		fmt.Fprintf(w, "Error adding person")
		return
	}

	// Create the data to send
	userData := map[string]string{
		"login": data.Login,
	}

	// Marshal the data into JSON format
	jsonData, err := json.Marshal(userData)
	if err != nil {
		http.Error(w, "Error marshaling JSON", http.StatusInternalServerError)
		return
	}

	request, err := http.NewRequest("POST", "https://127.0.0.1:8443/create-user", bytes.NewBuffer(jsonData))
	if err != nil {
		http.Error(w, "Error request", http.StatusUnauthorized)
		return
	}

	request.Header.Set("Content-Type", "application/json")
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

	resp := map[string]interface{}{
		"message": respStr,
	}

	jsonResponse, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, "Ошибка при сериализации JSON", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResponse)
}

func getField(field string) string {
	var fieldRet string

	switch field {
	case "lastName":
		fieldRet = "last_name"
	case "firstName":
		fieldRet = "first_name"
	case "middleName":
		fieldRet = "middle_name"
	case "birthDate":
		fieldRet = "birth_date"
	case "gender":
		fieldRet = "gender"
	case "email":
		fieldRet = "email"
	case "additionalEmail":
		fieldRet = "additional_email"
	case "phone":
		fieldRet = "phone"
	case "mobile":
		fieldRet = "mobile"
	case "inn":
		fieldRet = "inn"
	case "snils":
		fieldRet = "snils"
	case "companyName":
		fieldRet = "company_name"
	case "shortCompanyName":
		fieldRet = "short_company_name"
	case "ogrn":
		fieldRet = "ogrn"
	case "inn2":
		fieldRet = "inn2"
	case "kpp":
		fieldRet = "kpp"
	default:
		fieldRet = "Error"
	}

	return fieldRet
}

func SaveData(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		fmt.Fprintf(w, "Incorrect request")
		return
	}

	var data struct {
		Field string `json:"field"`
		Value string `json:"value"`
		Token string `json:"token"`
	}

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Невозможно прочитать JSON", http.StatusBadRequest)
		return
	}

	db, err := database.Initialize()
	if err != nil {
		log.Fatal(err)
		return
	}
	defer db.Close()

	field := getField(data.Field)

	_, err = db.DB.Exec("update Person set "+field+" = $1 where token = $2", data.Value, data.Token)
	if err != nil {
		fmt.Fprintf(w, "Error create person: "+fmt.Sprint(err))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Данные успешно получены"})
}

func GetData(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		fmt.Fprintf(w, "Incorrect request")
		return
	}

	var data struct {
		Token string `json:"token"`
	}

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Невозможно прочитать JSON", http.StatusBadRequest)
		return
	}

	db, err := database.Initialize()
	if err != nil {
		log.Fatal(err)
		return
	}
	defer db.Close()

	rows, err := db.DB.Query(`select login, last_name, first_name, middle_name, 
										birth_date, gender, email, additional_email, 
										phone, mobile, inn, snils, company_name, 
										short_company_name, ogrn, inn2, kpp 
							from Person where token = $1`, data.Token)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	var p struct {
		Login            sql.NullString
		LastName         sql.NullString
		FirstName        sql.NullString
		MiddleName       sql.NullString
		BirthDate        sql.NullString
		Gender           sql.NullString
		Email            sql.NullString
		AdditionalEmail  sql.NullString
		Phone            sql.NullString
		Mobile           sql.NullString
		INN              sql.NullString
		SNILS            sql.NullString
		CompanyName      sql.NullString
		ShortCompanyName sql.NullString
		OGRN             sql.NullString
		INN2             sql.NullString
		KPP              sql.NullString
	}

	for rows.Next() {
		_ = rows.Scan(
			&p.Login, &p.LastName, &p.FirstName, &p.MiddleName,
			&p.BirthDate, &p.Gender, &p.Email, &p.AdditionalEmail,
			&p.Phone, &p.Mobile, &p.INN, &p.SNILS, &p.CompanyName,
			&p.ShortCompanyName, &p.OGRN, &p.INN2, &p.KPP,
		)
	}

	var PersonData struct {
		Login            string
		LastName         string
		FirstName        string
		MiddleName       string
		BirthDate        string
		Gender           string
		Email            string
		AdditionalEmail  string
		Phone            string
		Mobile           string
		INN              string
		SNILS            string
		CompanyName      string
		ShortCompanyName string
		OGRN             string
		INN2             string
		KPP              string
	}

	if p.Login.Valid {
		PersonData.Login = p.Login.String
	} else {
		PersonData.Login = ""
	}

	if p.LastName.Valid {
		PersonData.LastName = p.LastName.String
	} else {
		PersonData.LastName = ""
	}

	if p.FirstName.Valid {
		PersonData.FirstName = p.FirstName.String
	} else {
		PersonData.FirstName = ""
	}

	if p.MiddleName.Valid {
		PersonData.MiddleName = p.MiddleName.String
	} else {
		PersonData.MiddleName = ""
	}

	if p.BirthDate.Valid {
		PersonData.BirthDate = p.BirthDate.String
	} else {
		PersonData.BirthDate = ""
	}

	if p.Gender.Valid {
		PersonData.Gender = p.Gender.String
	} else {
		PersonData.Gender = ""
	}

	if p.Email.Valid {
		PersonData.Email = p.Email.String
	} else {
		PersonData.Email = ""
	}

	if p.AdditionalEmail.Valid {
		PersonData.AdditionalEmail = p.AdditionalEmail.String
	} else {
		PersonData.AdditionalEmail = ""
	}

	if p.Phone.Valid {
		PersonData.Phone = p.Phone.String
	} else {
		PersonData.Phone = ""
	}

	if p.Mobile.Valid {
		PersonData.Mobile = p.Mobile.String
	} else {
		PersonData.Mobile = ""
	}

	if p.INN.Valid {
		PersonData.INN = p.INN.String
	} else {
		PersonData.INN = ""
	}

	if p.SNILS.Valid {
		PersonData.SNILS = p.SNILS.String
	} else {
		PersonData.SNILS = ""
	}

	if p.CompanyName.Valid {
		PersonData.CompanyName = p.CompanyName.String
	} else {
		PersonData.CompanyName = ""
	}

	if p.ShortCompanyName.Valid {
		PersonData.ShortCompanyName = p.ShortCompanyName.String
	} else {
		PersonData.ShortCompanyName = ""
	}

	if p.OGRN.Valid {
		PersonData.OGRN = p.OGRN.String
	} else {
		PersonData.OGRN = ""
	}

	if p.INN2.Valid {
		PersonData.INN2 = p.INN2.String
	} else {
		PersonData.INN2 = ""
	}

	if p.KPP.Valid {
		PersonData.KPP = p.KPP.String
	} else {
		PersonData.KPP = ""
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(PersonData)
}

type Application struct {
	Token     string `json:"token"`
	Login     string `json:"login"`
	Number    string `json:"number"`
	Project   string `json:"project"`
	Amount    string `json:"amount"`
	Selection string `json:"selection"`
	Status    string `json:"status"`
}

func GetApplications(w http.ResponseWriter, r *http.Request) {
	var Token struct {
		Token string `json:"token"`
	}
	// Декодирование JSON из тела запроса
	if err := json.NewDecoder(r.Body).Decode(&Token); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	db, ok := database.Initialize()
	if ok != nil {
		log.Fatal(ok)
	}
	defer db.Close()

	result, err := db.DB.Query("select login from Person where token = $1", Token.Token)
	if err != nil {
		panic(err)
	}
	defer result.Close()

	var login string
	if result.Next() {
		err := result.Scan(&login)
		if err != nil {
			panic(err)
		}
	} else {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Response{Success: false})
	}

	rows, err := db.DB.Query("select number, project, amount, selection, status from Appl where login = $1", login)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var appls []Application

	for rows.Next() {
		var appl Application
		if err := rows.Scan(&appl.Number, &appl.Project, &appl.Amount, &appl.Selection, &appl.Status); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		//appl.Token = Token.Token
		appls = append(appls, appl)
	}

	// Проверяем ошибки после завершения цикла
	if err = rows.Err(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Устанавливаем заголовок Content-Type
	w.Header().Set("Content-Type", "application/json")

	// Кодировка данных в формат JSON
	err = json.NewEncoder(w).Encode(appls)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

type Response struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func AddApplication(w http.ResponseWriter, r *http.Request) {
	var appl Application

	// Декодирование JSON из тела запроса
	if err := json.NewDecoder(r.Body).Decode(&appl); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	db, ok := database.Initialize()
	if ok != nil {
		log.Fatal(ok)
	}
	defer db.Close()

	result, err := db.DB.Query("select login from Person where token = $1", appl.Token)
	if err != nil {
		panic(err)
	}
	defer result.Close()

	var login string
	if result.Next() {
		err := result.Scan(&login)
		if err != nil {
			panic(err)
		}
	} else {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Response{Success: false, Message: "Заявка не добавлена"})
	}

	_, ok = db.DB.Exec(`insert into Appl (login, number, project,
										  amount, selection, status) values ($1, $2, $3, $4, $5, $6)`,
		login, appl.Number, appl.Project, appl.Amount, appl.Selection, "В ожидании")
	if ok != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Response{Success: false, Message: "Заявка не добавлена"})
		return
	}

	// Пример успешного ответа
	response := Response{
		Success: true,
		Message: "Заявка успешно добавлена",
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func GetLogin(w http.ResponseWriter, r *http.Request) {
	var Token struct {
		Token string `json:"token"`
	}
	// Декодирование JSON из тела запроса
	if err := json.NewDecoder(r.Body).Decode(&Token); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	db, ok := database.Initialize()
	if ok != nil {
		log.Fatal(ok)
	}
	defer db.Close()

	result, err := db.DB.Query("select login from Person where token = $1", Token.Token)
	if err != nil {
		panic(err)
	}
	defer result.Close()

	var login string
	if result.Next() {
		err := result.Scan(&login)
		if err != nil {
			panic(err)
		}
	}

	// Устанавливаем заголовок Content-Type
	w.Header().Set("Content-Type", "application/json")

	// Кодировка данных в формат JSON
	err = json.NewEncoder(w).Encode(login)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
