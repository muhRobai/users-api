package restapi

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
)

func (c *InitAPI) ListUser(ctx context.Context, req *GetUsers) (*GetUsers, error) {
	limit := 10

	if req.Limit != 0 {
		limit = int(req.Limit)
	}

	rows, err := c.Db.Query(ctx, `
		SELECT id, 
			username, 
			email,
			status, 
			role_id,
			created_at,
			updated_at
		FROM users LIMIT $1
	`, limit)

	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer rows.Close()
	var items []*User
	for rows.Next() {
		var item User
		var updateTime sql.NullString
		var status string
		err = rows.Scan(&item.Id,
			&item.Username,
			&item.Email,
			&status,
			&item.RoleId,
			&item.CreatedAt,
			&updateTime,
		)

		if err != nil {
			log.Println(err)
			return nil, err
		}

		item.UpdatedAt = updateTime.String

		items = append(items, &item)
	}

	if len(items) == 0 {
		return nil, errors.New("user-not-found")
	}

	return &GetUsers{
		List: items,
	}, nil
}

// CreateUser for creating user
func (c *InitAPI) CreateUser(ctx context.Context, req *User, rolesId string) (*UserId, error) {
	var id string
	roles, err := c.GetRoles(rolesId)
	if err != nil {
		log.Println(err)
		if err.Error() == "no rows in result set" {
			return nil, errors.New("ERROR-NO-ADMIN-FOUND")
		}
		return nil, err
	}

	if roles != "ADMIN" {
		return nil, errors.New("invalid-roles")
	}

	status := strconv.Itoa(req.Status)
	err = c.Db.QueryRow(ctx, `INSERT INTO users (username, email, status, role_id) VALUES ($1, $2, $3, $4) RETURNING id`,
		req.Username, req.Email, status, "uuid-ngarang").Scan(&id)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return &UserId{
		Id: id,
	}, nil
}

func (c *InitAPI) GetRoles(id string) (string, error) {
	var roles string
	err := c.Db.QueryRow(context.Background(), `SELECT roles FROM roles WHERE id = $1`, id).Scan(&roles)
	if err != nil {
		log.Println(err)
		return "", err
	}

	return roles, nil
}

func (c *InitAPI) GetCustomerById(id string) bool {
	var userId string
	err := c.Db.QueryRow(context.Background(), `SELECT username FROM users WHERE id = $1`, id).Scan(&userId)
	if err != nil {
		return false
	}

	return userId != ""
}

func (c *InitAPI) InsertProfilePhoto(ctx context.Context, req *FileItem) (*UserId, error) {
	if !c.GetCustomerById(req.UserId) {
		return nil, errors.New("user-not-found")
	}

	var profileId string
	err := c.Db.QueryRow(ctx, `INSERT INTO profile_photo (user_id, filename, file_type, size) VALUES ($1, $2, $3, $4) RETURNING id`,
		req.UserId, req.Filename, req.FileType, req.FileSize,
	).Scan(&profileId)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	filename := fmt.Sprintf("assert/%s", req.Filename)

	file, err := os.Create(filename)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer file.Close()
	_, err = io.Copy(file, req.File)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return &UserId{
		Id: profileId,
	}, nil
}

func (c *InitAPI) GetProfilePhotoById(id string) (string, string, error) {
	var filename, fileType string
	err := c.Db.QueryRow(context.Background(), `SELECT filename, file_type FROM profile_photo WHERE user_id = $1`, id).Scan(&filename, &fileType)

	if err != nil {
		log.Println(err)
		return "", "", err
	}

	return filename, fileType, nil
}

func (c *InitAPI) GetProfilePhoto(ctx context.Context, req *GetFile) (io.Reader, string, error) {
	filename, fileType, err := c.GetProfilePhotoById(req.UserId)
	if err != nil {
		return nil, "", nil
	}
	url := fmt.Sprintf("assert/%s", filename)
	file, err := os.Open(url)
	if err != nil {
		log.Println(err)
		return nil, "", err
	}

	return file, fileType, nil
}
