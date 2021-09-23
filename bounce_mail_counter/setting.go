package bouncemailcounter

import (
	"log"

	bouncemailcounter "github.com/Tungnt24/bounce-mail-counter"
	"github.com/kelseyhightower/envconfig"
)

func Load() bouncemailcounter.Config {
	var cfg bouncemailcounter.Config
	err := envconfig.Process("mail_counter", &cfg)
	if err != nil {
		log.Fatal(err.Error())
	}
	return cfg
}
