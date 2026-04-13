package data

import (
	"Green/internal/validator"
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
)


type User struct {
	ID 		int		`json:"id"`
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
