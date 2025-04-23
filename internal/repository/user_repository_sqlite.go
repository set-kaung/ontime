package repository

import (
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/set-kaung/senior_project_1/internal/domain"
)

type SQLiteUserRepository struct {
	db *sql.DB
}

func NewSQLiteUserRepository(db *sql.DB) *SQLiteUserRepository {
	return &SQLiteUserRepository{db: db}
}

func (s *SQLiteUserRepository) GetUserByID(id int) (*domain.User, error) {
	stmt := "SELECT * FROM users WHERE id = ?"
	row := s.db.QueryRow(stmt, id)
	u := &domain.User{}
	err := row.Scan(&u.ID, &u.Username, &u.PasswordHash, &u.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoRecord
		} else {
			return nil, err
		}
	}
	return u, nil
}

func (s *SQLiteUserRepository) GetUserByUsername(username string) (*domain.User, error) {
	stmt := "SELECT * FROM users WHERE username = ?"
	row := s.db.QueryRow(stmt, username)
	u := &domain.User{}
	err := row.Scan(&u.ID, &u.Username, &u.PasswordHash, &u.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoRecord
		} else {
			return nil, err
		}
	}
	return u, nil
}

func (s *SQLiteUserRepository) GetUserByEmail(email string) (*domain.User, error) {
	stmt := "SELECT * FROM users WHERE email = ?"
	row := s.db.QueryRow(stmt, email)
	u := &domain.User{}
	err := row.Scan(&u.ID, &u.PasswordHash, &u.Email.Address)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoRecord
		} else {
			return nil, err
		}
	}
	return u, nil
}

func (s *SQLiteUserRepository) InsertUser(user *domain.User) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	stmt := `INSERT INTO users (password,email) VALUES(?,?)`
	result, err := s.db.Exec(stmt, user.PasswordHash, user.Email.Address)
	if err != nil {
		tx.Rollback()
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return ErrDuplicateEmail
		}
		return err
	}
	userID, err := result.LastInsertId()
	if err != nil {
		tx.Rollback()
		return err
	}

	profileInsertStmt := `INSERT INTO profiles (user_id,username,tokens,date_joined) VALUES(?,?,?,CURRENT_DATE)`
	_, err = tx.Exec(profileInsertStmt, userID, user.Username, 0, time.Now())
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (s *SQLiteUserRepository) GetUserProfile(id int) (*domain.User, error) {
	stmt := `
					SELECT u.email, p.username, p.tokens, p.date_joined
					FROM users u
					JOIN profiles p ON u.id = p.user_id
					WHERE u.id = ?`
	row := s.db.QueryRow(stmt, id)
	u := &domain.User{}
	err := row.Scan(&u.Email.Address, &u.Username, &u.Tokens, &u.Joined)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoRecord
		} else {
			return nil, err
		}
	}
	return u, nil
}

func (s *SQLiteUserRepository) Exists(id int) (bool, error) {
	var exists bool
	stmt := "SELECT EXISTS(SELECT true FROM users where id = ?)"
	err := s.db.QueryRow(stmt, id).Scan(&exists)
	return exists, err
}
