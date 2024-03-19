package repository

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"strconv"
	"testing"

	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	_ "github.com/lib/pq"
)

type RepositoryTestSuite struct {
	suite.Suite
	mongoDb          *mongo.Database
	dockerPool       *dockertest.Pool
	dockerDbResource *dockertest.Resource
}

func (s *RepositoryTestSuite) SetupSuite() {

	user := "root"
	password := "password"

	pool, err := dockertest.NewPool("")
	require.NoError(s.T(), err, "could not connect to Docker")

	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Name:       "mongo-looping-louie-test",
		Repository: "mongo",
		Tag:        "4.4-rc-focal",
		Env: []string{
			fmt.Sprintf("MONGO_INITDB_ROOT_USERNAME=%s", user),
			fmt.Sprintf("MONGO_INITDB_ROOT_PASSWORD=%s", password),
		},
	}, func(config *docker.HostConfig) {
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{
			Name: "no",
		}
	})
	require.NoError(s.T(), err, "could not start container")

	port, err := strconv.ParseInt(resource.GetPort("27017/tcp"), 10, 32)

	clientOptions := options.Client().ApplyURI(fmt.Sprintf("mongodb://%s:%s@%s:%d", user, password, "localhost", int(port)))
	client, err := mongo.Connect(context.Background(), clientOptions)

	if err := pool.Retry(func() error {
		var err error

		s.dockerPool = pool
		s.dockerDbResource = resource
		s.mongoDb = client.Database("test-db")

		if err != nil {
			log.Fatalf("Could not connect to database: %s", err)
			return err
		}

		return client.Ping(context.Background(), nil)
	}); err != nil {
		log.Fatalf("Could not connect to database: %s", err)
	}

	assert.NoError(s.T(), err)
}

func (s *RepositoryTestSuite) TearDownTest() {
	err := s.mongoDb.Drop(context.Background())

	assert.NoError(s.T(), err)
}

func (s *RepositoryTestSuite) TearDownSuite() {
	err := s.mongoDb.Client().Disconnect(context.Background())
	if err = s.dockerPool.Purge(s.dockerDbResource); err != nil {
		log.Fatalf("Could not purge resource: %s", err)
	}

	assert.NoError(s.T(), err)
}

func TestRepositorySuite(t *testing.T) {
	suite.Run(t, new(RepositoryTestSuite))
}
