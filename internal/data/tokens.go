package data

import (
	"crypto/rand"
	"crypto/sha256"
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

	token.Plaintext = base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(randomBytes)
	
	
	hash := sha256.Sum256([]byte(token.Plaintext))

	token.Hash = hash[:]
	return token, nil
}