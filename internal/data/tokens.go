package data

import (
	"Green/internal/validator"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base32"
	"time"
)


const(
	ScopeActivation = "activation"
)

type Token struct {
	Plaintext string
	Hash []byte
	UserID int64
	Expiry time.Time
	scope  string
}


// the time.Duration is a struct with nanosecond as int64

// for security related works use crypto/rand as standard practice, dont use math/random it can be predictable
// use math/random for normal uses which we need faster random no. generation(not security related)
func generateToken(userID int64, ttl time.Duration, scope string) (*Token, error) {
	token := &Token{
		UserID: userID,
		Expiry: time.Now().Add(ttl), // .add function accept duration struct and return Time struct
		scope: scope,
	}

	// initializing a byte slice of size 16byte
	randomBytes := make([]byte,16)

	// read the random bytes cryptographically secure random
	//number generator (CSPRNG) into the randombyte slice
	_,err := rand.Read(randomBytes)
	if err != nil {
		return nil,err
	}


	// random bytes → base32 encode → human-readable string → store as token
	//base32.StdEncoding Base32 is an encoding scheme that converts binary data into readable text using only 32 characters:
	//.WithPadding(base32.NoPadding) By default base32 adds = padding at the end: NoPadding removes those = signs — cleaner for use in URLs or headers.
	// Base32 has exactly 32 possible characters:2⁵ = 32  →  you need exactly 5 bits
	//base32 reads in groups of 5 bits, not 8: so 16bytes = 128bits, 128/5 = 26characters
	//token.Plaintext contain 26byte content, the 5 bit grouping comes from the no.of characters it can represent its best to user 5 to convert binary to string
	token.Plaintext = base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(randomBytes)
	
	
      // the sum256 fn always return a 32 byte array, we convert plantext to byte and pass to sum256
	hash := sha256.Sum256([]byte(token.Plaintext))

	// convert the array to slice
	// database driver expects a slice []byte, not an array [32]byte
	token.Hash = hash[:]
	return token, nil
}


// validating the input token
func ValidateTokenPlaintext(v *validator.Validator, tokenPlaintext string) {
	// check the string is empty or not
	v.Check(tokenPlaintext != "", "token", "must be provide ")
	//check the string has 26 byte long or length is 26 
	v.Check(len(tokenPlaintext) == 26, "token", "must be 26 byte long")
}

// Define tokenModel for distributing db instance
type TokenModel struct {
	DB *sql.DB
}

func(m TokenModel) New(userID int64, ttl time.Duration, scope string) (*Token, error) {
	token,err := generateToken(userID, ttl, scope)
	if err!=nil {
		return nil,err
	}
	err = m.Insert(token)
	return token, nil
}


// insert the newly generated token and user details to the Token db
func(m TokenModel) Insert(t *Token) error {
	query := `INSERT INTO tokens (hash, user_id, expiry, scope)
			VALUES ($1, $2, $3, $4)`

	args := []any{t.Hash, t.UserID, t.Expiry,t.scope}

	ctx,cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_,err := m.DB.ExecContext(ctx, query, args...)

	return err
}

// pass as pointer if struct is small or no modification done by that fn
// if struct is large or need to modify the data inside then use pointer


// delete the token or other things of a user from table
func(m TokenModel) DeleteAllForUser(userID int64, scope string) error {
	query := `
			DELETE FROM tokens
			WHERE scope = $1 AND user_id = $2`
	
	// Creating a context with timeout of 3 second
	ctx,cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []any{scope,userID}
	_,err := m.DB.ExecContext(ctx, query,args... )
	return err
}