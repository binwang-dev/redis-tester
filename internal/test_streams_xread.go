package internal

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"

	testerutils "github.com/codecrafters-io/tester-utils"
	"github.com/codecrafters-io/tester-utils/logger"
	testerutils_random "github.com/codecrafters-io/tester-utils/random"
	"github.com/go-redis/redis"
)

type XREADTest struct {
	streams          []string
	block            *time.Duration
	expectedResponse []redis.XStream
	expectedError    string
}

func (t *XREADTest) Run(client *redis.Client, logger *logger.Logger) error {
	var resp []redis.XStream
	var err error

	if t.block == nil {
		logger.Infof("$ redis-cli xread streams %s", strings.Join(t.streams, " "))

		resp, err = client.XRead(&redis.XReadArgs{
			Streams: t.streams,
		}).Result()
	} else {
		logger.Infof("$ redis-cli xread block %v streams %s", t.block.Milliseconds(), strings.Join(t.streams, " "))

		resp, err = client.XRead(&redis.XReadArgs{
			Streams: t.streams,
			Block:   *t.block,
		}).Result()
	}

	if t.expectedError != "" {
		if err.Error() != t.expectedError {
			logFriendlyError(logger, err)
			return err
		} else {
			logFriendlyError(logger, err)
			logger.Successf("Received error: \"%v\"", err.Error())
			return nil
		}
	}

	if err != nil {
		logFriendlyError(logger, err)
		return err
	}

	expectedRespJson, err := json.MarshalIndent(t.expectedResponse, "", "  ")

	if err != nil {
		logFriendlyError(logger, err)
		return err
	}

	respJson, err := json.MarshalIndent(resp, "", "  ")

	if err != nil {
		logFriendlyError(logger, err)
		return err
	}

	if !reflect.DeepEqual(resp, t.expectedResponse) {
		logger.Infof("Received response: \"%v\"", string(respJson))
		return fmt.Errorf("Expected %#v, got %#v", string(expectedRespJson), string(respJson))
	} else {
		logger.Successf("Received response: \"%v\"", string(respJson))
	}

	return nil
}

func testStreamsXread(stageHarness *testerutils.StageHarness) error {
	b := NewRedisBinary(stageHarness)
	if err := b.Run(); err != nil {
		return err
	}

	logger := stageHarness.Logger
	client := NewRedisClient("localhost:6379")

	randomKey := testerutils_random.RandomWord()
	randomInt := testerutils_random.RandomInt(1, 100)

	testXadd(client, logger, XADDTest{
		streamKey:        randomKey,
		id:               "0-1",
		values:           map[string]interface{}{"temperature": randomInt},
		expectedResponse: "0-1",
	})

	expectedResp := []redis.XStream{
		{
			Stream: randomKey,
			Messages: []redis.XMessage{
				{
					ID:     "0-1",
					Values: map[string]interface{}{"temperature": randomInt},
				},
			},
		},
	}

	(&XREADTest{
		streams:          []string{randomKey, "0-0"},
		expectedResponse: expectedResp,
	}).Run(client, logger)

	return nil
}
