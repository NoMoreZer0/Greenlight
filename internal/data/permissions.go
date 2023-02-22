package data

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/lib/pq"
)

type Permissions []string

func (p Permissions) Include(code string) bool {
	for i := range p {
		if code == p[i] {
			return true
		}
	}
	return false
}

type PermissionModel struct {
	DB *pgxpool.Pool
}

func (m PermissionModel) GetAllForUser(userID int64) (Permissions, error) {
	conn, err := m.DB.Acquire(context.Background())
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	fmt.Println(userID)
	rows, err := conn.Query(ctx, "SELECT permissions.code FROM permissions INNER JOIN users_permissions ON users_permissions.permission_id = permissions.id INNER JOIN users ON users_permissions.user_id = users.id WHERE users.id = $1", userID)
	if err != nil {
		return nil, err
	}

	fmt.Println("ok1")

	defer rows.Close()

	var permissions Permissions

	for rows.Next() {
		var permission string
		err = rows.Scan(&permission)
		if err != nil {
			return nil, err
		}
		permissions = append(permissions, permission)
	}

	fmt.Println("ok2")

	if err = rows.Err(); err != nil {
		return nil, err
	}

	fmt.Println("ok3")

	return permissions, nil
}

func (m PermissionModel) AddForUser(userID int64, codes ...string) error {
	conn, err := m.DB.Acquire(context.Background())
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err = conn.Query(ctx, "INSERT INTO users_permissions SELECT $1, permissions.id FROM permissions WHERE permissions.code = ANY($2)", userID, pq.Array(codes))

	return err
}
