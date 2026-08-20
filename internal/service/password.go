package service

import "golang.org/x/crypto/bcrypt"

// hashPassword хэширует пароль перед сохранением в БД.
func hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// checkPassword сравнивает введённый пароль с сохранённым хэшем.
// Возвращает true, если пароль верный.
func checkPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
