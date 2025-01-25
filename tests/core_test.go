package _tests

import (
	"encoding/json"
	"errors"
	"firehose/pkg/core"
	"io"

	"github.com/bluesky-social/indigo/api/atproto"
	"github.com/bluesky-social/indigo/api/bsky"
	"github.com/bluesky-social/indigo/lex/util"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type CoreTestSuite struct {
	suite.Suite
}

type MockCBORMarshaler struct {
	mock.Mock
}

func (m *MockCBORMarshaler) MarshalCBOR(w io.Writer) error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockCBORMarshaler) MarshalJSON() ([]byte, error) {
	args := m.Called()
	return args.Get(0).([]byte), args.Error(1)
}
func (suite *CoreTestSuite) TestFetchPostIdentifier_Success() {
	mockClient := new(MockAPIClient)

	mockMarshaler := new(MockCBORMarshaler)
	mockJSON := []byte(`{"Subject":{"Uri":"mock-uri"}}`)

	mockMarshaler.On("MarshalJSON").Return(mockJSON, nil)

	mockDecoder := &util.LexiconTypeDecoder{
		Val: mockMarshaler,
	}

	mockOutput := &atproto.RepoGetRecord_Output{
		Value: mockDecoder,
	}

	mockClient.On(
		"RepoGetRecord",
		mock.Anything,
		mock.Anything,
		"",
		"collection",
		"repo",
		"rkey",
	).Return(mockOutput, nil)

	res, err := core.FetchPostIdentifier(mockClient, "repo", "collection/rkey")

	suite.Assert().NoError(err)
	suite.Assert().Equal("mock-uri", res)

	mockClient.AssertExpectations(suite.T())
	mockMarshaler.AssertExpectations(suite.T())
}

func (suite *CoreTestSuite) TestFetchPostIdentifier_Failure_Record() {
	mockClient := new(MockAPIClient)

	mockMarshaler := new(MockCBORMarshaler)
	mockJSON := []byte(`{"Subject":{"Uri":"mock-uri"}}`)

	mockMarshaler.On("MarshalJSON").Return(mockJSON, nil)

	mockClient.On(
		"RepoGetRecord",
		mock.Anything,
		mock.Anything,
		"",
		"collection",
		"repo",
		"rkey",
	).Return(nil, errors.New(""))

	res, err := core.FetchPostIdentifier(mockClient, "repo", "collection/rkey")

	suite.Assert().Error(err)
	suite.Assert().Nil(res)

	mockClient.AssertExpectations(suite.T())
	mockMarshaler.AssertExpectations(suite.T())
}

func (suite *CoreTestSuite) TestFetchPostIdentifier_Failure_MarshalJSON() {
	mockClient := new(MockAPIClient)

	mockMarshaler := new(MockCBORMarshaler)

	mockMarshaler.On("MarshalJSON").Return(([]byte)(nil), errors.New(""))

	mockDecoder := &util.LexiconTypeDecoder{
		Val: mockMarshaler,
	}

	mockOutput := &atproto.RepoGetRecord_Output{
		Value: mockDecoder,
	}

	mockClient.On(
		"RepoGetRecord",
		mock.Anything,
		mock.Anything,
		"",
		"collection",
		"repo",
		"rkey",
	).Return(mockOutput, nil)

	res, err := core.FetchPostIdentifier(mockClient, "repo", "collection/rkey")

	suite.Assert().Error(err)
	suite.Assert().Nil(res)

	mockClient.AssertExpectations(suite.T())
	mockMarshaler.AssertExpectations(suite.T())
}

func (suite *CoreTestSuite) TestFetchPostDetails_Success() {
	mockClient := new(MockAPIClient)
	mockMarshaler := new(MockCBORMarshaler)

	mockPost := bsky.FeedPost{Text: "Test post text"}
	mockAuthor := bsky.ActorDefs_ProfileViewBasic{Handle: "exampleHandle", Did: "exampleDid"}
	mockJSON, _ := json.Marshal(mockPost)
	mockMarshaler.On("MarshalJSON").Return(mockJSON, nil)

	mockPostView := bsky.FeedDefs_PostView{
		Uri:    "at://example/repo/rkey",
		Author: &mockAuthor,
		Record: &util.LexiconTypeDecoder{Val: mockMarshaler},
	}

	mockRecord := bsky.FeedGetPosts_Output{
		Posts: []*bsky.FeedDefs_PostView{&mockPostView},
	}

	mockClient.On("GetPost", mock.Anything, "at://example/repo/rkey").Return(mockRecord, nil)

	postDetails, err := core.FetchPostDetails(mockClient, "at://example/repo/rkey")

	suite.Assert().NoError(err)
	suite.Assert().NotNil(postDetails)
	suite.Assert().Equal("exampleHandle", postDetails.Handle)
	suite.Assert().Equal("Test post text", postDetails.Text)
	suite.Assert().Equal("exampleDid", postDetails.Repo)
	suite.Assert().Equal("rkey", postDetails.Rkey)

	mockClient.AssertExpectations(suite.T())
	mockMarshaler.AssertExpectations(suite.T())
}
