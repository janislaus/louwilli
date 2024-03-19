package configuration

type Config struct {
	Database struct {
		DatabaseServer   string `envconfig:"DB_SERVER" default:"localhost" required:"true"`
		DatabaseName     string `envconfig:"DB_NAME" default:"gamestats" required:"true"`
		DatabaseUser     string `envconfig:"DB_USER" default:"sa" required:"true"`
		DatabasePassword string `envconfig:"DB_PASSWORD" default:"superuser" required:"true"`
		DatabasePort     int    `envconfig:"DB_PORT" default:"27017" required:"true"`
	}
	Kafka struct {
		Server          string `envconfig:"KAFKA_SERVER" default:"localhost" required:"true"`
		Port            string `envconfig:"KAFKA_PORT" default:"9093" required:"true"`
		LouieEventTopic string `envconfig:"LOUIE_EVENT_TOPIC" default:"LOUIE_EVENT" required:"true"`
	}
}
