package users

import (
	"bytes"
	"encoding/base64"
	"encoding/csv"
	"fmt"
	"os"
)

var UserEPMap map[string]string

func init() {
	UserEPMap = make(map[string]string)
}

func getUsers(db string) ([][]string, error) {
	raw, err := os.ReadFile(db)
	if err != nil {
		return nil, err
	}

	r := csv.NewReader(bytes.NewReader(raw))
	users, err := r.ReadAll()
	if err != nil {
		return nil, err
	}

	return users, nil
}

func AddUser(db, user, passwd string) error {
	users, err := getUsers(db)
	if err != nil {
		return err
	}

	for _, uinfo := range users {
		if len(uinfo) < 2 {
			continue
		}

		if user == uinfo[0] {
			return fmt.Errorf("user %s existed", user)
		}
	}

	f, err := os.OpenFile(db, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return err
	}
	defer f.Close()

	passwdEnc := base64.StdEncoding.EncodeToString([]byte(passwd))
	line := fmt.Sprintf("%s,%s\n", user, passwdEnc)

	if _, err = f.WriteString(line); err != nil {
		return err
	}

	return nil
}

func ValidateUser(db, user, passwd string) error {
	users, err := getUsers(db)
	if err != nil {
		return err
	}

	for _, uinfo := range users {
		if len(uinfo) < 2 {
			continue
		}

		if uinfo[0] == user {
			if uinfo[1] == base64.StdEncoding.EncodeToString([]byte(passwd)) {
				return nil
			}

			return fmt.Errorf("wrong passwd for user %s", user)
		}
	}

	return fmt.Errorf("user %s not exist", user)
}
