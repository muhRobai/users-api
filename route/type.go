package restapi

import (
	"io"
	"time"

	"github.com/jackc/pgx/v4"
)

type GetUsers struct {
	Limit int32   `json:"limit"`
	List  []*User `json:"list"`
}

type User struct {
	Id        string    `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Status    int       `json:"status"`
	RoleId    string    `json:"roleId"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt string    `json:"updatedAt"`
}

type InitAPI struct {
	Db *pgx.Conn
}

type UserId struct {
	Id string `json:"id"`
}

type UserName struct {
	Name string `json:"name"`
}

type FileItem struct {
	File     io.Reader
	Filename string
	FileSize int64
	FileType string
	UserId   string
}

type GetFile struct {
	UserId string `json:"userId"`
}
