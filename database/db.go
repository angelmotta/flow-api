package database

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
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
	GetUser(email string) (*User, error)
	CreateUser(user *User) error
	DeleteUser(id int) error
	//GetUsers() ([]*User, error)
	//UpdateUser(user *User) error
}

func NewStore(db *pgxpool.Pool) Store {
	return &store{db} // store type implements store interface
}

// The actual store containing the Postgres database pool (state)
type store struct {
	db *pgxpool.Pool
}

func (s *store) GetUser(email string) (*User, error) {
	var user User
	err := s.db.QueryRow(context.Background(), "select id, email, role, dni, name, lastname_main, lastname_secondary, address, created_at from users where email = $1", email).Scan(&user.Id, &user.Email, &user.Role, &user.Dni, &user.Name, &user.LastnameMain, &user.LastnameSecondary, &user.Address, &user.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			log.Println("db layer: User not found")
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (s *store) CreateUser(user *User) error {
	_, err := s.db.Exec(context.Background(), "insert into users (email, role, dni, name, lastname_main, lastname_secondary, address) values ($1, $2, $3, $4, $5, $6, $7)", user.Email, user.Role, user.Dni, user.Name, user.LastnameMain, user.LastnameSecondary, user.Address)
	var pgErr *pgconn.PgError
	if err != nil {
		log.Println("Error captured from database layer in CreateUser")
		log.Println(err)
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return errors.New("user already exists")
			}
		}
		return errors.New("internal server error")
	}
	return nil
}

func (s *store) DeleteUser(id int) error {
	commandTag, err := s.db.Exec(context.Background(), "delete from users where id = $1", id)
	if err != nil {
		log.Println("Error captured from database layer in DeleteUser")
		log.Println(err)
		return errors.New("internal server error")
	}
	if commandTag.RowsAffected() != 1 {
		return errors.New("user not found")
	}
	return nil
}
