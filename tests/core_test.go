package _tests

import (
	"context"
	"encoding/json"
	"errors"
	"firehose/pkg/api"
	"firehose/pkg/core"
	"firehose/pkg/utils"
	"io"
	"testing"
	"time"

	"github.com/bluesky-social/indigo/api/atproto"
	"github.com/bluesky-social/indigo/api/bsky"
	"github.com/bluesky-social/indigo/lex/util"
	"github.com/cenkalti/backoff/v5"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type CoreTestSuite struct {
	suite.Suite
	originalBackoffOpts backoff.RetryOption
	originalMaxRetries  backoff.RetryOption
}

func TestCoreTestSuite(t *testing.T) {
	suite.Run(t, &CoreTestSuite{})
}

type MockCBORMarshaler struct {
	mock.Mock
	LexiconTypeID string `cborgen:"const=app.bsky.feed.like"`
}

func (m *MockCBORMarshaler) MarshalCBOR(w io.Writer) error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockCBORMarshaler) MarshalJSON() ([]byte, error) {
	args := m.Called()
	return args.Get(0).([]byte), args.Error(1)
}

type MockDownloadClient struct {
	mock.Mock
}

func (m *MockDownloadClient) FetchPostIdentifier(ctx context.Context, client api.APIClient, repo, path string) (string, error) {
	args := m.Called(ctx, client, repo, path)
	return args.Get(0).(string), args.Error(1)
}

func (m *MockDownloadClient) FetchPostDetails(ctx context.Context, client api.APIClient, atUri string) (*core.PostDetails, error) {
	args := m.Called(ctx, client, atUri)
	return args.Get(0).(*core.PostDetails), args.Error(1)
}

func (m *MockDownloadClient) DownloadBlobs(ctx context.Context, APIClient api.APIClient, FSClient utils.FileSystem, media *utils.Media, postDetails *core.PostDetails, directory string) error {
	args := m.Called(ctx, APIClient, FSClient, media, postDetails, directory)
	return args.Error(0)
}

func (suite *CoreTestSuite) SetupSuite() {
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

func (suite *CoreTestSuite) TearDownSuite() {
	api.BackoffOpts = suite.originalBackoffOpts
	api.MaxRetries = suite.originalMaxRetries
}

func (suite *CoreTestSuite) TestFetchPostIdentifier_Success() {
	mockClient := new(MockAPIClient)
	mockMarshaler := &MockCBORMarshaler{
		LexiconTypeID: "app.bsky.feed.like",
	}

	mockJSON := []byte(`{"$type":"app.bsky.feed.like","createdAt":"2025-01-26T14:35:51.135Z","subject":{"cid":"bafyreid34vpni5jvmiisfvtpjp6s54k2me5wtomfb6qcr63lqledr7tcxy","uri":"at://did:plc:vdnlidrx2n2nitqimqymzutr/app.bsky.feed.post/3lgmu7ro53226"}}`)
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

	res, err := core.FetchPostIdentifier(context.Background(), mockClient, "repo", "collection/rkey")

	suite.Assert().NoError(err)
	suite.Assert().Equal("at://did:plc:vdnlidrx2n2nitqimqymzutr/app.bsky.feed.post/3lgmu7ro53226", res)

	mockClient.AssertExpectations(suite.T())
	mockMarshaler.AssertExpectations(suite.T())
}

func (suite *CoreTestSuite) TestFetchPostIdentifier_Failure_Record() {
	mockClient := new(MockAPIClient)

	mockClient.On(
		"RepoGetRecord",
		mock.Anything,
		mock.Anything,
		"",
		"collection",
		"repo",
		"rkey",
	).Return((*atproto.RepoGetRecord_Output)(nil), errors.New(""))

	res, err := core.FetchPostIdentifier(context.Background(), mockClient, "repo", "collection/rkey")

	suite.Assert().Error(err)
	suite.Assert().Equal("", res)

	mockClient.AssertExpectations(suite.T())
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

	res, err := core.FetchPostIdentifier(context.Background(), mockClient, "repo", "collection/rkey")

	suite.Assert().Error(err)
	suite.Assert().Equal("", res)

	mockClient.AssertExpectations(suite.T())
	// mockMarshaler.AssertExpectations(suite.T())
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

	mockClient.On(
		"FeedGetPosts",
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(&mockRecord, nil)

	postDetails, err := core.FetchPostDetails(context.Background(), mockClient, "at://example/repo/rkey")

	suite.Assert().NoError(err)
	suite.Assert().NotNil(postDetails)
	suite.Assert().Equal("exampleHandle", postDetails.Handle)
	suite.Assert().Equal("Test post text", postDetails.Text)
	suite.Assert().Equal("exampleDid", postDetails.Repo)
	suite.Assert().Equal("rkey", postDetails.Rkey)

	mockClient.AssertExpectations(suite.T())
	mockMarshaler.AssertExpectations(suite.T())
}

func (suite *CoreTestSuite) TestFetchPostDetails_Failure_No_Posts() {
	mockClient := new(MockAPIClient)

	mockRecord := bsky.FeedGetPosts_Output{
		Posts: []*bsky.FeedDefs_PostView{},
	}

	mockClient.On(
		"FeedGetPosts",
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(&mockRecord, nil)

	postDetails, err := core.FetchPostDetails(context.Background(), mockClient, "at://example/repo/rkey")

	suite.Assert().Error(err)
	suite.Assert().Nil(postDetails)

	mockClient.AssertExpectations(suite.T())
}

func (suite *CoreTestSuite) TestDownloadBlobs_Success_Images() {
	mockClient := &MockAPIClient{}
	mockFS := &MockFileSystem{}
	mockFile := &MockFile{}
	mockMedia := utils.Media{
		ImageCid:  []string{"example_cid_1", "exmaple_cid_2"},
		MediaType: "jpeg",
	}
	mockPostDetails := &core.PostDetails{
		Handle: "example_handle",
		Text:   "example_text",
		Repo:   "did:plc:example",
		Rkey:   "example_rkey",
		Media: &utils.Media{
			MediaType: "jpeg",
		},
	}

	mockClient.On("SyncGetBlob", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]byte("blob data"), nil)
	mockFile.On("Write", mock.Anything).Return(len([]byte("test data")), nil)
	mockFile.On("Close").Return(nil)
	mockFS.On("OpenFile", mock.Anything, mock.Anything, mock.Anything).Return(mockFile, nil)

	err := core.DownloadBlobs(context.Background(), mockClient, mockFS, &mockMedia, mockPostDetails, "example_dir")

	suite.Assert().Nil(err)

	mockClient.AssertExpectations(suite.T())
	mockFS.AssertExpectations(suite.T())
	mockFile.AssertExpectations(suite.T())
}

func (suite *CoreTestSuite) TestDownloadBlobs_Success_Video() {
	mockClient := &MockAPIClient{}
	mockFS := &MockFileSystem{}
	mockFile := &MockFile{}
	mockMedia := utils.Media{
		VideoCid:  "example_cid",
		MediaType: "mp4",
	}
	mockPostDetails := &core.PostDetails{
		Handle: "example_handle",
		Text:   "example_text",
		Repo:   "did:plc:example",
		Rkey:   "example_rkey",
		Media: &utils.Media{
			MediaType: "mp4",
		},
	}

	mockClient.On("SyncGetBlob", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]byte("blob data"), nil)
	mockFile.On("Write", mock.Anything).Return(len([]byte("test data")), nil)
	mockFile.On("Close").Return(nil)
	mockFS.On("OpenFile", mock.Anything, mock.Anything, mock.Anything).Return(mockFile, nil)

	err := core.DownloadBlobs(context.Background(), mockClient, mockFS, &mockMedia, mockPostDetails, "example_dir")

	suite.Assert().Nil(err)

	mockClient.AssertExpectations(suite.T())
	mockFS.AssertExpectations(suite.T())
	mockFile.AssertExpectations(suite.T())
}

func (suite *CoreTestSuite) TestDownloadBlobs_Failure_API() {
	mockClient := &MockAPIClient{}
	mockFS := &MockFileSystem{}
	mockMedia := utils.Media{
		VideoCid:  "example_cid",
		MediaType: "mp4",
	}
	mockPostDetails := &core.PostDetails{
		Handle: "example_handle",
		Text:   "example_text",
		Repo:   "did:plc:example",
		Rkey:   "example_rkey",
		Media: &utils.Media{
			MediaType: "mp4",
		},
	}

	mockClient.On("SyncGetBlob", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(([]byte)(nil), errors.New(""))

	err := core.DownloadBlobs(context.Background(), mockClient, mockFS, &mockMedia, mockPostDetails, "example_dir")

	suite.Assert().Error(err)

	mockClient.AssertExpectations(suite.T())
}

func (suite *CoreTestSuite) TestDownloadBlobs_Failure_Write_File() {
	mockClient := &MockAPIClient{}
	mockFS := &MockFileSystem{}
	mockFile := &MockFile{}
	mockMedia := utils.Media{
		VideoCid:  "example_cid",
		MediaType: "mp4",
	}
	mockPostDetails := &core.PostDetails{
		Handle: "example_handle",
		Text:   "example_text",
		Repo:   "did:plc:example",
		Rkey:   "example_rkey",
		Media: &utils.Media{
			MediaType: "mp4",
		},
	}

	mockClient.On("SyncGetBlob", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]byte("blob data"), nil)
	mockFS.On("OpenFile", mock.Anything, mock.Anything, mock.Anything).Return(mockFile, nil)
	mockFile.On("Write", mock.Anything).Return(0, errors.New(""))
	mockFile.On("Close").Return(nil)

	err := core.DownloadBlobs(context.Background(), mockClient, mockFS, &mockMedia, mockPostDetails, "example_dir")

	suite.Assert().Error(err)

	mockClient.AssertExpectations(suite.T())
	mockFS.AssertExpectations(suite.T())
	mockFile.AssertExpectations(suite.T())
}

func (suite *CoreTestSuite) TestDownloadBlobs_Failure_Open_File() {
	mockClient := &MockAPIClient{}
	mockFS := &MockFileSystem{}
	mockFile := &MockFile{}
	mockMedia := utils.Media{
		VideoCid:  "example_cid",
		MediaType: "mp4",
	}
	mockPostDetails := &core.PostDetails{
		Handle: "example_handle",
		Text:   "example_text",
		Repo:   "did:plc:example",
		Rkey:   "example_rkey",
		Media: &utils.Media{
			MediaType: "mp4",
		},
	}

	mockClient.On("SyncGetBlob", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]byte("blob data"), nil)
	mockFS.On("OpenFile", mock.Anything, mock.Anything, mock.Anything).Return(mockFile, errors.New(""))

	err := core.DownloadBlobs(context.Background(), mockClient, mockFS, &mockMedia, mockPostDetails, "example_dir")

	suite.Assert().Error(err)

	mockClient.AssertExpectations(suite.T())
	mockFS.AssertExpectations(suite.T())
	mockFile.AssertExpectations(suite.T())
}

func (suite *CoreTestSuite) TestDownloadPost_Success() {
	mockFile := &MockFile{}
	mockAPIClient := &MockAPIClient{}
	mockFS := &MockFileSystem{}
	mockClient := &MockDownloadClient{}
	mockPostDetails := &core.PostDetails{
		Handle: "example_handle",
		Text:   "example_text",
		Repo:   "did:plc:example",
		Rkey:   "example_rkey",
		Media: &utils.Media{
			MediaType: "mp4",
		},
	}
	mockAtUri := "example_aturi"

	mockFile.On("Write", mock.Anything).Return(len([]byte("test data")), nil)
	mockFile.On("Close").Return(nil)
	mockFS.On("OpenFile", mock.Anything, mock.Anything, mock.Anything).Return(mockFile, nil)

	mockClient.On("FetchPostIdentifier", mock.Anything, mockAPIClient, mock.Anything, mock.Anything).Return(mockAtUri, nil)
	mockClient.On("FetchPostDetails", mock.Anything, mockAPIClient, mockAtUri).Return(mockPostDetails, nil)
	mockClient.On("DownloadBlobs", mock.Anything, mockAPIClient, mockFS, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	core.DownloadPost(context.Background(), mockClient, mockAPIClient, mockFS, "repo_string", "repo_path", "dir")
	mockFile.AssertExpectations(suite.T())
	mockAPIClient.AssertExpectations(suite.T())
	mockFS.AssertExpectations(suite.T())
	mockClient.AssertExpectations(suite.T())
}
