package data

import (
	"Green/internal/validator"
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"time"

	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

// user struct to accept data from the handler to the internal
type User struct {
	ID 		int64		`json:"id"`
	Created_at  time.Time	`json:"created_at"`
	Name 		string	`json:"name"`
	Email 	string	`json:"email"`
	Password    password	`json:"-"`
	Activated   bool		`json:"activated"`
	Version     int		`json:"-"`

}

//  If you used a plain string, a missing password and an empty password would look identical — and you might accidentally hash an empty string and store it.
// if we use normal string for plaintext instead of pointer then 
// You can't distinguish between these two situations:
// The user didn't provide a plaintext password at all
// The user provided a plaintext password that happens to be an empty string
//In both cases, plaintext would be "".
//but if we use *string then password not provided -> value = nil, Password is empty string -> points to  ""
// , Password is "Secret123" points to "secret123"
// So you can check if plaintext == nil to know definitively that no password was supplied — not just that it was empty.
type password struct {
	plainText *string
	hash []byte

}


func (p *password) Set(plainTextPassword string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(plainTextPassword), 12)
	if err!=nil {
		return err
	}
	p.plainText = &plainTextPassword
	p.hash = hash
	return nil
}

func (p *password) Matches(plainTextPassword string) (bool, error) {
	err := bcrypt.CompareHashAndPassword(p.hash,[]byte(plainTextPassword))
	if err!=nil {
		// checking any error happening in match password
		switch {
		case errors.Is(err,bcrypt.ErrMismatchedHashAndPassword):
			return false,nil // return false and no error is password is not matches
		default:
			return false,err // return error if any error happended in comparehash fn
		}
	}
	return true, nil
}



// validation check for user

func ValidateEmail(v *validator.Validator, email string) {
	v.Check(email!="", "email", "must be provided")
	v.Check(validator.IsValidEmail(email), "email", "Must be a valid email address")
}

func ValidatePassword(v *validator.Validator, password string) {
	v.Check(password != "", "password", "must be provided") // its provided a empty string
	v.Check(len(password)>=8, "password", "must be atleast 8 byte long")
	v.Check(len(password)<= 72, "password", "must not be more than 72 bytes long")
}

func ValidateUser(v *validator.Validator, user *User) {
	v.Check(user.Name != "", "name", "must be provided")
	v.Check(len(user.Name)<=500, "name", 	"must not be more than 500 bytes long")
	ValidateEmail(v, user.Email)

	// if password plaintext pointer is not nil it means something is provided by user
	if user.Password.plainText !=nil {
		ValidatePassword(v, *user.Password.plainText)
	}

	if user.Password.hash == nil {
		panic("missing password hash for user")
	}
}


var (
	ErrDuplicateEmail = errors.New("duplicate email")
)

type UserModel struct {
	db *sql.DB
}

func (m UserModel) Insert(user *User) error {
	query := `
			INSERT INTO users (name, email, password_hash, activated)
			VALUES ($1, $2, $3, $4)
			RETURNING id, created_at, version`
	args := []any{user.Name, user.Email, user.Password.hash, user.Activated}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.db.QueryRowContext(ctx, query,args...).Scan(&user.ID, &user.Created_at, &user.Version)
	if err!=nil {
    		var pqErr *pq.Error
		switch {
		case errors.As(err, &pqErr) && pqErr.Code == "23505":
			return ErrDuplicateEmail
		default:
			return err
		}
	}
	return nil
}

func(m UserModel) GetByEmail(email string) (*User, error) {
	query := `
		SELECT id, created_at, name, email, password_hash, activated, version
		FROM users
		WHERE email = $1`
	
	var user User

	ctx, cancel  :=  context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Created_at,
		&user.Name,
		&user.Email,
		&user.Password.hash, 
		&user.Activated,
		&user.Version,
	)
	if err!=nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	return &user, nil
}

func (m *UserModel) Update(user *User) error {
	query := `
			UPDATE users
			SET name = $1, email = $2, password_hash = $3, activated = $4, version = version + 1
			WHERE id = $5 AND version = $6
			RETURNING version`

	args := []any{
		user.Name,
		user.Email,
		user.Password.hash,
		user.Activated,
		user.ID,
		user.Version,
	}

	ctx, cancel := context.WithTimeout(context.Background(),3*time.Second)
	defer cancel()

	err := m.db.QueryRowContext(ctx, query,args...).Scan(&user.Version)
	if err!=nil {
		switch {
			case err.Error() == `pq: duplicate key value violates unique constraint "users_email_key"`:
				return ErrDuplicateEmail  // citext (case insensitive text) as the field in db so it will manage capital letters
			case errors.Is(err, sql.ErrNoRows):
				return ErrEditConflict
			default:
				return err
		}
	}
	return nil
}

func (m UserModel) GetForToken(tokenScope, tokenPlaintext string) (*User, error) {
	tokenHash := sha256.Sum256([]byte(tokenPlaintext))

	query := `
			SELECT users.id, users.created_at, users.name, users.email, users.password_hash, users.activated, users.version
			FROM users
			INNER JOIN tokens
			ON users.id = tokens.user_id
			WHERE tokens.hash = $1
			AND tokens.scope = $2
			AND tokens.expiry > $3`

	args := []any{tokenHash[:], tokenScope, time.Now()}

	ctx, cancel := context.WithTimeout(context.Background(), 3 *time.Second)
	defer cancel()

	var user User

	// if a row is return or we want to scan something from db user queryrow or queryrowcontext other wise use exec or execcontext
	err := m.db.QueryRowContext(ctx,query,args...).Scan(&user.ID,
		 &user.Created_at, 
		 &user.Name, 
		 &user.Email, 
		 &user.Password.hash,
		 &user.Activated,
		 &user.Version,				
		)	

	if err!=nil {
		switch {
		case errors.Is(err,sql.ErrNoRows):
			return nil,ErrRecordNotFound
		default:
			return nil,err
		}
	}

	return &user, nil
}
