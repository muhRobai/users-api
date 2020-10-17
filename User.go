package restapi

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"strconv"
	"time"
)

func (c *InitAPI) ListUser(ctx context.Context, req *GetUsers) (*GetUsers, error) {
	limit := 10

	if req.Limit != 0 {
		limit = int(req.Limit)
	}

	rows, err := c.Db.Query(`
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
	err = c.Db.QueryRow(`INSERT INTO users (username, email, status, role_id, updated_at) VALUES ($1, $2, $3, $4, $5) RETURNING id`,
		req.Username, req.Email, status, "uuid-ngarang", time.Now().Format("RFC3339")).Scan(&id)
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
	err := c.Db.QueryRow(`SELECT roles FROM roles WHERE id = $1`, id).Scan(&roles)
	if err != nil {
		log.Println(err)
		return "", err
	}

	return roles, nil
}
