// +build !heroku

package main

func LoadConfig() (config *Config, err error) {
	defer recoverEnvironmentPanic(&err)

	return &Config{
		Addr:          mustGetenv("ADDR"),
		MongoURL:      mustGetenv("MONGODB_URL"),
		AccessToken:   mustGetenv("ACCESS_TOKEN"),
		IsDevelopment: true,
	}, nil
}
