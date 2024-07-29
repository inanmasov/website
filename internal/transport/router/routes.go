package router

import (
	"net/http"

	handler "example.com/Go/internal/transport/handler"
)

func RegisterRoutes() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "front/start.html", http.StatusFound)
	})
	http.HandleFunc("/sign", handler.Sign)
	http.HandleFunc("/register", handler.Register)
	http.HandleFunc("/save_data", handler.SaveData)
	http.HandleFunc("/get_data", handler.GetData)
	http.HandleFunc("/get_applications", handler.GetApplications)
	http.HandleFunc("/add_application", handler.AddApplication)
	http.HandleFunc("/get_login", handler.GetLogin)
	http.HandleFunc("/get_list_files", handler.GetListFiles)
	http.HandleFunc("/get_file", handler.GetFile)
	http.HandleFunc("/delete_file", handler.DeleteFile)
	http.HandleFunc("/upload_file", handler.UploadFile)

	fs := http.FileServer(http.Dir("front"))
	http.Handle("/front/", http.StripPrefix("/front/", fs))

}
