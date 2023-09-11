package database

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
)

type User struct {
	Id                int
	Email             string
	Role              string
	Dni               string
	Name              string
	LastnameMain      string
	LastnameSecondary string
	Address           string
	CreatedAt         time.Time
}

type Store interface {
	GetUser(id int) (*User, error)
	//GetUsers() ([]*User, error)
	//CreateUser(user *User) error
	//UpdateUser(user *User) error
	//DeleteUser(id int) error
}

func NewStore(db *pgxpool.Pool) Store {
	return &store{db}
}

// The actual store containing the database pool (state)
type store struct {
	db *pgxpool.Pool
}

func (d *store) GetUser(id int) (*User, error) {
	var user User
	err := d.db.QueryRow(context.Background(), "select id, email, role, dni, name, lastname_main, lastname_secondary, address, created_at from users where id = $1", id).Scan(&user.Id, &user.Email, &user.Role, &user.Dni, &user.Name, &user.LastnameMain, &user.LastnameSecondary, &user.Address, &user.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &user, nil
}
