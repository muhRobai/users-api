package restapi

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"

	"github.com/jackc/pgx/v4"
)

func (c *InitAPI) initDb() {
	// dbHost := "127.0.0.1"
	// dbPass := "root"
	// dbName := "postgres"
	// dbPort := "5432"
	// dbUser := "postgres"

	// port, err := strconv.Atoi(dbPort)
	// if err != nil {
	// 	log.Println(err)
	// 	return
	// }

	// dbConfig := &pgx.ConnConfig{
	// 	Port:     uint16(port),
	// 	Host:     dbHost,
	// 	User:     dbUser,
	// 	Password: dbPass,
	// 	Database: dbName,
	// }

	// connection := pgx.ConnPoolConfig{
	// 	ConnConfig:     *dbConfig,
	// 	MaxConnections: 5,
	// }
	var err error
	var url = "postgres://tdwaspai:ThGq-16XbXx44fGrr_3ABrpOX29WyhDa@satao.db.elephantsql.com:5432/tdwaspai"
	c.Db, err = pgx.Connect(context.Background(), url)
	if err != nil {
		log.Println(err)
		return
	}
}

func (c *InitAPI) HandleListUser(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var p GetUsers
	err := json.NewDecoder(r.Body).Decode(&p)
	if err != nil {
		http.Error(w, "failed-to-convert-json", http.StatusBadRequest)
		return
	}

	resp, err := c.ListUser(ctx, &p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, "failed-conver-json", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func (c *InitAPI) HandleCreateUser(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var p User
	err := json.NewDecoder(r.Body).Decode(&p)
	if err != nil {
		http.Error(w, "failed-to-convert-json", http.StatusBadRequest)
		return
	}

	roleid := r.Header.Get("ROLE-ID")
	resp, err := c.CreateUser(ctx, &p, roleid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, "failed-conver-json", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func (c *InitAPI) HandleUploadPhoto(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer file.Close()

	id := r.FormValue("userId")
	resp, err := c.InsertProfilePhoto(ctx, &FileItem{
		File:     file,
		UserId:   id,
		Filename: header.Filename,
		FileSize: header.Size,
		FileType: header.Header["Content-Type"][0],
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, "failed-conver-json", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func (c *InitAPI) HandleGetProfilePhoto(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["userid"]

	img, fileType, err := c.GetProfilePhoto(r.Context(), &GetFile{
		UserId: id,
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", fileType) // <-- set the content-type header
	io.Copy(w, img)
}

func StartHttp() http.Handler {
	api := createAPI()
	api.initDb()

	r := mux.NewRouter()
	r.HandleFunc("/api/user/list", api.HandleListUser).Methods("GET")
	r.HandleFunc("/api/user/create", api.HandleCreateUser).Methods("POST")
	r.HandleFunc("/api/user/photo", api.HandleUploadPhoto).Methods("POST")
	r.HandleFunc("/api/user/photo/{userid}", api.HandleGetProfilePhoto).Methods("GET")
	return r
}
