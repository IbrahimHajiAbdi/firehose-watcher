package _tests

import (
	"context"
	"errors"
	"firehose/pkg/api"
	"testing"
	"time"

	"github.com/bluesky-social/indigo/api/atproto"
	"github.com/bluesky-social/indigo/api/bsky"
	"github.com/bluesky-social/indigo/xrpc"
	"github.com/cenkalti/backoff/v5"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type APITestSuite struct {
	suite.Suite
	originalBackoffOpts backoff.RetryOption
	originalMaxRetries  backoff.RetryOption
}

func TestAPITestSuite(t *testing.T) {
	suite.Run(t, &APITestSuite{})
}

type MockAPIClient struct {
	mock.Mock
}

func (m *MockAPIClient) SyncGetBlob(ctx context.Context, client *xrpc.Client, cid, repo string) ([]byte, error) {
	args := m.Called(ctx, client, cid, repo)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockAPIClient) RepoGetRecord(ctx context.Context, client *xrpc.Client, token, collection, repo, rkey string) (*atproto.RepoGetRecord_Output, error) {
	args := m.Called(ctx, client, token, collection, repo, rkey)
	return args.Get(0).(*atproto.RepoGetRecord_Output), args.Error(1)
}

func (m *MockAPIClient) FeedGetPosts(ctx context.Context, client *xrpc.Client, uris []string) (*bsky.FeedGetPosts_Output, error) {
	args := m.Called(ctx, client, uris)
	return args.Get(0).(*bsky.FeedGetPosts_Output), args.Error(1)
}

func (suite *APITestSuite) SetupSuite() {
	suite.originalBackoffOpts = api.BackoffOpts
	suite.originalMaxRetries = api.MaxRetries

	api.BackoffOpts = backoff.WithBackOff(
		&backoff.ExponentialBackOff{
			InitialInterval:     1 * time.Millisecond,
			RandomizationFactor: 0.0,
			Multiplier:          1.0,
			MaxInterval:         1 * time.Millisecond,
		})
	api.MaxRetries = backoff.WithMaxTries(5)
}

func (suite *APITestSuite) TearDownSuite() {
	api.BackoffOpts = suite.originalBackoffOpts
	api.MaxRetries = suite.originalMaxRetries
}

func (suite *APITestSuite) TestGetBlob_Success() {
	mockClient := new(MockAPIClient)
	mockClient.On("SyncGetBlob", mock.Anything, mock.Anything, "cid1", "repo1").Return([]byte("blob data"), nil)

	res, err := api.GetBlob(mockClient, "repo1", "cid1")

	suite.Assert().NoError(err)
	suite.Assert().NotNil(res)
	suite.Assert().Equal([]byte("blob data"), *res)
}

func (suite *APITestSuite) TestGetBlob_Failure() {
	mockClient := new(MockAPIClient)
	mockClient.On("SyncGetBlob", mock.Anything, mock.Anything, "cid1", "repo1").Return(([]byte)(nil), errors.New("mock error"))

	res, err := api.GetBlob(mockClient, "repo1", "cid1")

	suite.Assert().Error(err)
	suite.Assert().Nil(res)
}

func (suite *APITestSuite) TestGetPosts_Success() {
	mockClient := new(MockAPIClient)
	expectedOutput := &bsky.FeedGetPosts_Output{}
	mockClient.On("FeedGetPosts", mock.Anything, mock.Anything, mock.Anything).Return(expectedOutput, nil)

	res, err := api.GetPost(mockClient, "mock_uri")

	suite.Assert().Nil(err)
	suite.Assert().Equal(expectedOutput, res)
	mockClient.AssertExpectations(suite.T())
}

func (suite *APITestSuite) TestGetPosts_Failure() {
	mockClient := new(MockAPIClient)
	mockClient.On("FeedGetPosts", mock.Anything, mock.Anything, mock.Anything).Return((*bsky.FeedGetPosts_Output)(nil), errors.New("error"))

	res, err := api.GetPost(mockClient, "mock_uri")

	suite.Assert().Nil(res)
	suite.Assert().Error(err)
	mockClient.AssertExpectations(suite.T())
}

func (suite *APITestSuite) TestGetRecord_Success() {
	mockClient := new(MockAPIClient)
	expectedOutput := &atproto.RepoGetRecord_Output{}
	mockClient.On("RepoGetRecord", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(expectedOutput, nil)

	res, err := api.GetRecord(mockClient, "", "", "")

	suite.Assert().Nil(err)
	suite.Assert().Equal(expectedOutput, res)
	mockClient.AssertExpectations(suite.T())
}

func (suite *APITestSuite) TestGetRecord_Failure() {
	mockClient := new(MockAPIClient)
	mockClient.On("RepoGetRecord", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return((*atproto.RepoGetRecord_Output)(nil), errors.New("error"))

	res, err := api.GetRecord(mockClient, "", "", "")

	suite.Assert().Nil(res)
	suite.Assert().Error(err)
	mockClient.AssertExpectations(suite.T())
}
