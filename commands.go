// Copyright 2016 Mender Software AS
//
//    Licensed under the Apache License, Version 2.0 (the "License");
//    you may not use this file except in compliance with the License.
//    You may obtain a copy of the License at
//
//        http://www.apache.org/licenses/LICENSE-2.0
//
//    Unless required by applicable law or agreed to in writing, software
//    distributed under the License is distributed on an "AS IS" BASIS,
//    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//    See the License for the specific language governing permissions and
//    limitations under the License.
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/mendersoftware/go-lib-micro/config"
	"github.com/mendersoftware/go-lib-micro/log"
	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh/terminal"

	"github.com/mendersoftware/useradm/model"
	"github.com/mendersoftware/useradm/store/mongo"
	"github.com/mendersoftware/useradm/user"
)

// safeReadPassword reads a user password from a terminal in a safe way (without
// echoing the characters input by the user)
func safeReadPassword() (string, error) {
	stdinfd := int(os.Stdin.Fd())

	if !terminal.IsTerminal(stdinfd) {
		return "", errors.New("stdin is not a terminal")
	}

	fmt.Fprintf(os.Stderr, "Enter password: ")
	raw, err := terminal.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		return "", errors.Wrap(err, "failed to read password")
	}
	fmt.Fprintf(os.Stderr, "\n")

	return string(raw), nil
}

func commandCreateUser(c config.Reader, username string, password string) error {
	l := log.NewEmpty()

	l.Debugf("create user '%s'", username)

	if password == "" {
		var err error
		if password, err = safeReadPassword(); err != nil {
			return err
		}
	}

	u := model.User{
		Email:    username,
		Password: password,
	}

	if err := u.ValidateNew(); err != nil {
		return errors.Wrap(err, "user validation failed")
	}

	db, err := mongo.GetDataStoreMongo(c.GetString(SettingDb))
	if err != nil {
		return errors.Wrap(err, "database connection failed")
	}

	ua := useradm.NewUserAdm(nil, db, useradm.Config{})

	if err := ua.CreateUser(context.Background(), &u); err != nil {
		return errors.Wrap(err, "creating user failed")
	}

	return nil
}
