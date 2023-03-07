package utils

import (
	"fmt"
	"os"
)

func GetMongoURI() string {
	url := fmt.Sprintf("mongodb://%s:%s@%s:%s/test?authSource=admin", os.Getenv("MONGO_USER_NAME"),
		os.Getenv("MONGO_PASSWORD"), os.Getenv("MONGO_IP_ADDRESS"), os.Getenv("MONGO_PORT"))
	return url
}
