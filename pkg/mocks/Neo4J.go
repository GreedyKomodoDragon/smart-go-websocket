package mocks

import (
	"context"
	"fmt"
	"io"

	. "github.com/onsi/gomega"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func StartNeo4jContainer(ctx context.Context, username, password string) (testcontainers.Container, error) {
	request := testcontainers.ContainerRequest{
		Image:        "neo4j",
		ExposedPorts: []string{"7687/tcp"},
		Env:          map[string]string{"NEO4J_AUTH": fmt.Sprintf("%s/%s", username, password)},
		WaitingFor:   wait.ForLog("Bolt enabled"),
	}
	return testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: request,
		Started:          true,
	})
}

func Close(closer io.Closer, resourceName string) {
	Expect(closer.Close()).
		To(BeNil(), "%s should close", resourceName)
}
