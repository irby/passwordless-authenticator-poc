package mocks

import "github.com/teamhanko/hanko/backend/config"

func GenerateMockConfig() config.Config {
	config := config.Config{
		Passcode: config.Passcode{
			Email: config.Email{
				FromAddress: "hello@example.com",
				FromName:    "Hello",
			},
		},
		Service: config.Service{
			Name: "Passwordless Authenticator",
		},
	}
	return config
}
