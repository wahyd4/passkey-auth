package database

import (
	"database/sql"
	"time"

	_ "github.com/glebarez/go-sqlite"
)

type DB struct {
	conn *sql.DB
}

type User struct {
	ID          int       `json:"id"`
	Email       string    `json:"email"`
	DisplayName string    `json:"display_name"`
	Approved    bool      `json:"approved"`
	CreatedAt   time.Time `json:"created_at"`
}

type Credential struct {
	ID              []byte    `json:"id"`
	UserID          int       `json:"user_id"`
	PublicKey       []byte    `json:"public_key"`
	AttestationType string    `json:"attestation_type"`
	AAGUID          []byte    `json:"aaguid"`
	SignCount       uint32    `json:"sign_count"`
	CloneWarning    bool      `json:"clone_warning"`
	CreatedAt       time.Time `json:"created_at"`
}

func New(dbPath string) (*DB, error) {
	conn, err := sql.Open("sqlite", dbPath+"?_fk=1")
	if err != nil {
		return nil, err
	}

	db := &DB{conn: conn}
	if err := db.migrate(); err != nil {
		return nil, err
	}

	return db, nil
}

func (db *DB) Close() error {
	return db.conn.Close()
}

func (db *DB) migrate() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			email TEXT UNIQUE NOT NULL,
			display_name TEXT NOT NULL,
			approved BOOLEAN DEFAULT FALSE,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS credentials (
			id BLOB PRIMARY KEY,
			user_id INTEGER NOT NULL,
			public_key BLOB NOT NULL,
			attestation_type TEXT NOT NULL,
			aaguid BLOB,
			sign_count INTEGER DEFAULT 0,
			clone_warning BOOLEAN DEFAULT FALSE,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)`,
		`CREATE INDEX IF NOT EXISTS idx_users_email ON users(email)`,
		`CREATE INDEX IF NOT EXISTS idx_credentials_user_id ON credentials(user_id)`,
	}

	for _, query := range queries {
		if _, err := db.conn.Exec(query); err != nil {
			return err
		}
	}

	return nil
}

func (db *DB) CreateUser(email, displayName string) (*User, error) {
	result, err := db.conn.Exec(
		"INSERT INTO users (email, display_name) VALUES (?, ?)",
		email, displayName,
	)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return db.GetUser(int(id))
}

func (db *DB) CreateUserWithApproval(email, displayName string, approved bool) (*User, error) {
	result, err := db.conn.Exec(
		"INSERT INTO users (email, display_name, approved) VALUES (?, ?, ?)",
		email, displayName, approved,
	)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return db.GetUser(int(id))
}

func (db *DB) GetUser(id int) (*User, error) {
	var user User
	err := db.conn.QueryRow(
		"SELECT id, email, display_name, approved, created_at FROM users WHERE id = ?",
		id,
	).Scan(&user.ID, &user.Email, &user.DisplayName, &user.Approved, &user.CreatedAt)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (db *DB) GetUserByEmail(email string) (*User, error) {
	var user User
	err := db.conn.QueryRow(
		"SELECT id, email, display_name, approved, created_at FROM users WHERE email = ?",
		email,
	).Scan(&user.ID, &user.Email, &user.DisplayName, &user.Approved, &user.CreatedAt)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (db *DB) ListUsers() ([]*User, error) {
	rows, err := db.conn.Query(
		"SELECT id, email, display_name, approved, created_at FROM users ORDER BY created_at DESC",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*User
	for rows.Next() {
		var user User
		if err := rows.Scan(&user.ID, &user.Email, &user.DisplayName, &user.Approved, &user.CreatedAt); err != nil {
			return nil, err
		}
		users = append(users, &user)
	}

	return users, nil
}

func (db *DB) ApproveUser(id int) error {
	_, err := db.conn.Exec("UPDATE users SET approved = TRUE WHERE id = ?", id)
	return err
}

func (db *DB) DeleteUser(id int) error {
	_, err := db.conn.Exec("DELETE FROM users WHERE id = ?", id)
	return err
}

func (db *DB) SaveCredential(cred *Credential) error {
	_, err := db.conn.Exec(
		`INSERT INTO credentials (id, user_id, public_key, attestation_type, aaguid, sign_count, clone_warning)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		cred.ID, cred.UserID, cred.PublicKey, cred.AttestationType,
		cred.AAGUID, cred.SignCount, cred.CloneWarning,
	)
	return err
}

func (db *DB) GetCredentialsByUserID(userID int) ([]*Credential, error) {
	rows, err := db.conn.Query(
		`SELECT id, user_id, public_key, attestation_type, aaguid, sign_count, clone_warning, created_at
		 FROM credentials WHERE user_id = ?`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var credentials []*Credential
	for rows.Next() {
		var cred Credential
		if err := rows.Scan(
			&cred.ID, &cred.UserID, &cred.PublicKey, &cred.AttestationType,
			&cred.AAGUID, &cred.SignCount, &cred.CloneWarning, &cred.CreatedAt,
		); err != nil {
			return nil, err
		}
		credentials = append(credentials, &cred)
	}

	return credentials, nil
}

func (db *DB) UpdateCredentialSignCount(credID []byte, signCount uint32) error {
	_, err := db.conn.Exec(
		"UPDATE credentials SET sign_count = ? WHERE id = ?",
		signCount, credID,
	)
	return err
}
