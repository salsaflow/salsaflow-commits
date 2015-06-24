// +build heroku

package main

func LoadConfig() (config *Config, err error) {
	defer recoverEnvironmentPanic(&err)

	return &Config{
		Addr:          ":" + mustGetenv("PORT"),
		MongoURL:      mustGetenv("MONGOHQ_URL"),
		AccessToken:   mustGetenv("ACCESS_TOKEN"),
		IsDevelopment: false,
	}, nil
}
