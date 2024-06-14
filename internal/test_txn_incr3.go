package internal

import (
	"github.com/codecrafters-io/redis-tester/internal/redis_executable"
	"github.com/codecrafters-io/redis-tester/internal/resp_assertions"

	"github.com/codecrafters-io/redis-tester/internal/instrumented_resp_connection"
	"github.com/codecrafters-io/redis-tester/internal/test_cases"
	"github.com/codecrafters-io/tester-utils/random"
	"github.com/codecrafters-io/tester-utils/test_case_harness"
)

func testTxIncr3(stageHarness *test_case_harness.TestCaseHarness) error {
	b := redis_executable.NewRedisExecutable(stageHarness)
	if err := b.Run(); err != nil {
		return err
	}

	logger := stageHarness.Logger

	client, err := instrumented_resp_connection.NewFromAddr(stageHarness, "localhost:6379", "client")
	if err != nil {
		logFriendlyError(logger, err)
		return err
	}
	defer client.Close()

	uniqueKeys := random.RandomWords(2)
	randomKey, randomValue := uniqueKeys[0], uniqueKeys[1]

	multiCommandTestCase := test_cases.MultiCommandTestCase{
		Commands: [][]string{
			{"SET", randomKey, randomValue},
			{"INCR", randomKey},
		},
		Assertions: []resp_assertions.RESPAssertion{
			resp_assertions.NewStringAssertion("OK"),
			resp_assertions.NewErrorAssertion("ERR value is not an integer or out of range"),
		},
	}

	return multiCommandTestCase.RunAll(client, logger)
}
