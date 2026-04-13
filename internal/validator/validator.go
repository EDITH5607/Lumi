package validator

import (
	"net/mail"
)





type Validator struct {
	Errors map[string]string
}

func New() *Validator {
	return &Validator{
		Errors: make(map[string]string),
	}
}

func (v *Validator)Valid() bool {
	return len(v.Errors) == 0
}

func (v *Validator)AddError(key, message string) {
	if _,exists :=v.Errors[key]; !exists {
		v.Errors[key] = message
	}
} 

func (v *Validator)Check(ok bool, key, message string) {
	if !ok {
		v.AddError(key, message)
	}
}

func PermittedValue[T comparable] (value T, permittedValues ...T) bool {
	for i := range permittedValues {
		if value == permittedValues[i] {
			return  true
		}
	}
	return false
}

// func Matches(value string, rx *regexp.Regexp) bool {
// 	return rx.MatchString(value)
// }

func IsValidEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	if err!=nil {
		return false
	}
	return true
}

func	Unique[T comparable](values []T) bool {
	uniqueValue := make(map[T]bool)

	for _,value := range values {
		uniqueValue[value] = true
	} 
	return len(values) == len(uniqueValue)	
}