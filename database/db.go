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
	Id                int       `json:"id"`
	Email             string    `json:"email"`
	Role              string    `json:"role"`
	Dni               string    `json:"dni"`
	Name              string    `json:"name"`
	LastnameMain      string    `json:"lastnameMain"`
	LastnameSecondary string    `json:"lastnameSecondary"`
	Address           string    `json:"address"`
	CreatedAt         time.Time `json:"createdAt"`
}

type Store interface {
	GetUser(email string) (*User, error)
	CreateUser(user *User) error
	DeleteUser(id int) error
	//GetUsers() ([]*User, error)
	//UpdateUser(user *User) error
}

func NewPgStore(db *pgxpool.Pool) Store {
	return &storePostgres{db} // storePostgres type implements storePostgres interface
}

// The actual storePostgres containing the Postgres database pool (state)
type storePostgres struct {
	db *pgxpool.Pool
}

func (s *storePostgres) GetUser(email string) (*User, error) {
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

func (s *storePostgres) CreateUser(user *User) error {
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

func (s *storePostgres) DeleteUser(id int) error {
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

func (s *storePostgres) UpdateUser(id int) error {
	_, err := s.db.Exec(context.Background(), "update users set email = $1, role = $2, dni = $3, name = $4, lastname_main = $5, lastname_secondary = $6, address = $7 where id = $8", id)
	if err != nil {
		log.Println("Error captured from database layer in UpdateUser")
		log.Println(err)
		return errors.New("internal server error")
	}
	return nil
}
