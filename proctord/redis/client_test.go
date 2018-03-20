package redis

import (
	"sort"
	"testing"

	"github.com/garyburd/redigo/redis"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type RedisClientTestSuite struct {
	suite.Suite
	testRedisClient Client
	testRedisConn   redis.Conn
}

func (s *RedisClientTestSuite) SetupTest() {
	s.testRedisClient = NewClient()
	connPool, err := newPool()
	assert.NoError(s.T(), err)

	s.testRedisConn = connPool.Get()
}

func (s *RedisClientTestSuite) TestSET() {
	t := s.T()

	key, value := "anyKey", []byte("anyValue")
	err := s.testRedisClient.SET(key, value)
	assert.NoError(t, err)

	savedValue, err := redis.String(s.testRedisConn.Do("GET", key))
	assert.NoError(t, err)

	assert.Equal(t, "anyValue", string(savedValue))
}

func (s *RedisClientTestSuite) TestGET() {
	t := s.T()

	key, value := "anyKey", []byte("anyValue")
	err := s.testRedisClient.SET(key, value)
	assert.NoError(t, err)

	binaryValue, err := s.testRedisClient.GET(key)
	assert.NoError(t, err)

	assert.Equal(t, value, binaryValue)
}

func (s *RedisClientTestSuite) TestKEYS() {
	t := s.T()

	key, value := "job1-suffix", []byte("anyValue1")
	err := s.testRedisClient.SET(key, value)
	assert.NoError(t, err)
	key, value = "job2-suffix", []byte("anyValue2")
	err = s.testRedisClient.SET(key, value)
	assert.NoError(t, err)
	key, value = "job3", []byte("anyValue3")
	err = s.testRedisClient.SET(key, value)
	assert.NoError(t, err)

	jobNameKeyRegex := "*-suffix"
	keys, err := s.testRedisClient.KEYS(jobNameKeyRegex)
	assert.NoError(t, err)

	sort.Strings(keys)

	assert.EqualValues(t, []string{"job1-suffix", "job2-suffix"}, keys)
}

func (s *RedisClientTestSuite) TestMGET() {
	t := s.T()

	key, value := "job1-suffix", []byte("anyValue1")
	err := s.testRedisClient.SET(key, value)
	assert.NoError(t, err)
	key, value = "job2-suffix", []byte("anyValue2")
	err = s.testRedisClient.SET(key, value)
	assert.NoError(t, err)
	key, value = "job3", []byte("anyValue3")
	err = s.testRedisClient.SET(key, value)
	assert.NoError(t, err)

	jobKeys := make([]interface{}, 2)
	jobKeys[0] = "job1-suffix"
	jobKeys[1] = "job2-suffix"

	values, err := s.testRedisClient.MGET(jobKeys...)
	assert.NoError(t, err)

	sort.Slice(values, func(i, j int) bool {
		return string(values[i]) < string(values[j])
	})

	assert.EqualValues(t, [][]byte{[]byte("anyValue1"), []byte("anyValue2")}, values)
}

func (s *RedisClientTestSuite) TearDownSuite() {
	s.testRedisConn.Close()
}

func TestRedisClientTestSuite(t *testing.T) {
	suite.Run(t, new(RedisClientTestSuite))
}
