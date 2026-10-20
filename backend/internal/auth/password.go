package auth

import (
	"sync"

	"golang.org/x/crypto/bcrypt"
)

// Cost 10 equals Spring's BCryptPasswordEncoder default; existing "$2a$10$..."
// hashes created by the Spring backend verify unchanged.
const bcryptCost = 10

func HashPassword(password string) (string, error) {
	h, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	return string(h), err
}

func CheckPassword(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

var dummyHash = sync.OnceValue(func() []byte {
	h, _ := bcrypt.GenerateFromPassword([]byte("revu-timing-equaliser"), bcryptCost)
	return h
})

// BurnPasswordCheck performs a bcrypt comparison against a dummy hash so that
// logins for unknown e-mails take as long as logins with a wrong password.
func BurnPasswordCheck(password string) {
	_ = bcrypt.CompareHashAndPassword(dummyHash(), []byte(password))
}
