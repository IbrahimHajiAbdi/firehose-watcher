package _tests

import (
	"context"
	"encoding/json"
	"errors"
	"firehose/pkg/utils"
	"fmt"
	"os"
	"testing"

	"github.com/bluesky-social/indigo/api/atproto"
	"github.com/bluesky-social/indigo/api/bsky"
	"github.com/bluesky-social/indigo/xrpc"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type UtilsTestSuite struct {
	suite.Suite
	mockDirectory      string
	mockRkey           string
	mockHandle         string
	mockText           string
	mockMedia          string
	mockI              int
	mockMaxByte        int
	quoteImageFeedPost *bsky.FeedPost_Embed
	imageFeedPost      *bsky.FeedPost_Embed
	videoFeedPost      *bsky.FeedPost_Embed
}

func TestUtilsTestSuite(t *testing.T) {
	suite.Run(t, &UtilsTestSuite{})
}

type MockHandleResolver struct {
	mock.Mock
}

type MockFile struct {
	mock.Mock
}

type MockFileSystem struct {
	mock.Mock
}

func (m *MockFileSystem) OpenFile(name string, flag int, perm os.FileMode) (utils.File, error) {
	args := m.Called(name, flag, perm)
	return args.Get(0).(utils.File), args.Error(1)
}

func (m *MockFile) Write(data []byte) (int, error) {
	args := m.Called(data)
	return args.Get(0).(int), args.Error(1)
}

func (m *MockFile) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockHandleResolver) IdentityResolveHandle(ctx context.Context, c *xrpc.Client, handle string) (*atproto.IdentityResolveHandle_Output, error) {
	args := m.Called(ctx, c, handle)
	return args.Get(0).(*atproto.IdentityResolveHandle_Output), args.Error(1)
}

func (suite *UtilsTestSuite) SetupSuite() {
	suite.mockDirectory = "mock_dir"
	suite.mockRkey = "mock_rkey"
	suite.mockHandle = "mock_handle"
	suite.mockText = "mock_text"
	suite.mockMedia = "mock_type"
	suite.mockI = 2
	suite.mockMaxByte = 233
	var videoFeedPost *bsky.FeedPost_Embed
	var imageFeedPost *bsky.FeedPost_Embed
	var quoteImageFeedPost *bsky.FeedPost_Embed
	err := json.Unmarshal([]byte(`{"$type":"app.bsky.embed.video","aspectRatio":{"height":1024,"width":576},"video":{"$type":"blob","ref":{"$link":"bafkreiawmtb3mxmfcwuf4w4cmun6gjgrj3ktv5zpsq2gwtos2nkjaenqqe"},"mimeType":"video/mp4","size":2934003}}`), &videoFeedPost)
	if err != nil {
		fmt.Println(err)
		return
	}
	err = json.Unmarshal([]byte(`{"$type":"app.bsky.embed.images","images":[{"alt":"","aspectRatio":{"height":1936,"width":1936},"image":{"$type":"blob","ref":{"$link":"bafkreie6rdowktrct6f4ehi5ti5vjpx7krfflekgrwjkyjgrav7tziegpe"},"mimeType":"image/jpeg","size":985982}}]}`), &imageFeedPost)
	if err != nil {
		fmt.Println(err)
		return
	}
	err = json.Unmarshal([]byte(`{"$type":"app.bsky.embed.recordWithMedia","media":{"$type":"app.bsky.embed.images","images":[{"alt":"A man standing in the desert early morning wearing noise canceling headphones, aviator sunglasses and a blue Nike tech fleece ","aspectRatio":{"height":1998,"width":1126},"image":{"$type":"blob","ref":{"$link":"bafkreig3gejydod7xpuwd2bwtkkl2v3537raudjzhy2exqzqhzchuu6zlu"},"mimeType":"image/jpeg","size":825697}}]},"record":{"$type":"app.bsky.embed.record","record":{"cid":"bafyreif3nr2bekrhyxtvg44udrsshvib53hcpoiwkiuznoixu2ose47pxm","uri":"at://did:plc:7uqcpvrwbm3a6cu2edenskvd/app.bsky.feed.post/3lgedpgkcfc2x"}}}`), &quoteImageFeedPost)
	if err != nil {
		fmt.Println(err)
		return
	}
	suite.imageFeedPost = imageFeedPost
	suite.quoteImageFeedPost = quoteImageFeedPost
	suite.videoFeedPost = videoFeedPost
}

func (suite *UtilsTestSuite) TestFilepath() {
	res := utils.MakeFilepath(
		suite.mockDirectory,
		suite.mockRkey,
		suite.mockHandle,
		suite.mockText,
		suite.mockMedia,
		suite.mockI,
		suite.mockMaxByte,
	)
	expected := fmt.Sprintf(
		"%s/%s_%s_%s_%d.%s",
		suite.mockDirectory,
		suite.mockRkey,
		suite.mockHandle,
		suite.mockText,
		suite.mockI,
		suite.mockMedia,
	)
	suite.Assert().Equal(expected, res)
}

func (suite *UtilsTestSuite) TestFilenameLengthLimit_Exceed() {
	char := "😊" // 4 bytes
	var mockFilename string
	var expected string
	// 4 * 140 = 560 bytes
	for i := 0; i < 140; i++ {
		mockFilename += char
		if i < 63 {
			expected += char
		}
	}
	res := utils.FilenameLengthLimit(mockFilename, 255)

	suite.Assert().LessOrEqual(len([]byte(res)), 255)
	suite.Assert().Equal(expected, res)
}

func (suite *UtilsTestSuite) TestFilenameLengthLimit_UnderLimit() {
	mockFilename := "😊" // 4 bytes
	res := utils.FilenameLengthLimit(mockFilename, 255)

	suite.Assert().LessOrEqual(len(res), 255)
	suite.Assert().Equal(mockFilename, res)
}

func (suite *UtilsTestSuite) TestExtractMedia_Image() {
	res := utils.ExtractMedia(suite.imageFeedPost)
	expected := utils.Media{
		ImageCid:  []string{"bafkreie6rdowktrct6f4ehi5ti5vjpx7krfflekgrwjkyjgrav7tziegpe"},
		MediaType: "jpeg",
	}

	suite.Assert().Equal(expected, *res)
}

func (suite *UtilsTestSuite) TestExtractMedia_Video() {
	res := utils.ExtractMedia(suite.videoFeedPost)
	expected := utils.Media{
		VideoCid:  "bafkreiawmtb3mxmfcwuf4w4cmun6gjgrj3ktv5zpsq2gwtos2nkjaenqqe",
		MediaType: "mp4",
	}

	suite.Assert().Equal(expected, *res)
}

func (suite *UtilsTestSuite) TestExtractMedia_QuoteImage() {
	res := utils.ExtractMedia(suite.quoteImageFeedPost)
	expected := utils.Media{
		ImageCid:  []string{"bafkreig3gejydod7xpuwd2bwtkkl2v3537raudjzhy2exqzqhzchuu6zlu"},
		MediaType: "jpeg",
	}

	suite.Assert().Equal(expected, *res)
}

func (suite *UtilsTestSuite) TestResolveHandle_Success() {
	mockHandleResolver := &MockHandleResolver{}
	expected := &atproto.IdentityResolveHandle_Output{}
	mockHandleResolver.On("IdentityResolveHandle", mock.Anything, mock.Anything, mock.Anything).Return(&atproto.IdentityResolveHandle_Output{}, nil)
	res, err := utils.ResolveHandle(mockHandleResolver, suite.mockHandle)

	suite.Assert().Nil(err)
	suite.Assert().Equal(expected, res)
	mockHandleResolver.AssertExpectations(suite.T())
}

func (suite *UtilsTestSuite) TestResolveHandle_Failure() {
	mockHandleResolver := &MockHandleResolver{}
	mockHandleResolver.On("IdentityResolveHandle", mock.Anything, mock.Anything, mock.Anything).Return((*atproto.IdentityResolveHandle_Output)(nil), errors.New(""))
	res, err := utils.ResolveHandle(mockHandleResolver, suite.mockHandle)

	suite.Assert().Nil(res)
	suite.Assert().Error(err)
	mockHandleResolver.AssertExpectations(suite.T())
}

func (suite *UtilsTestSuite) TestFindExpression_Success() {
	mockRegex := `[^/]*$`
	mockStr := "hello/world"

	res := utils.FindExpression(mockRegex, mockStr)

	suite.Assert().Equal("world", res)
}

func (suite *UtilsTestSuite) TestFindExpression_Failure() {
	mockRegex := `[0-9]`
	mockStr := "sadasdasd"

	res := utils.FindExpression(mockRegex, mockStr)

	suite.Assert().Equal("", res)
}

func (suite *UtilsTestSuite) TestWriteFile_Success() {
	mockFS := &MockFileSystem{}
	mockFile := &MockFile{}
	mockFile.On("Write", mock.Anything).Return(len([]byte("test data")), nil)
	mockFile.On("Close").Return(nil)
	mockFS.On("OpenFile", "testfile.txt", os.O_CREATE|os.O_WRONLY, os.FileMode(0644)).Return(mockFile, nil)

	data := []byte("test data")
	err := utils.WriteFile(mockFS, "testfile.txt", &data)

	suite.Assert().NoError(err)
	mockFile.AssertCalled(suite.T(), "Write", data)
	mockFile.AssertCalled(suite.T(), "Close")
	mockFS.AssertCalled(suite.T(), "OpenFile", "testfile.txt", os.O_CREATE|os.O_WRONLY, os.FileMode(0644))

	mockFile.AssertExpectations(suite.T())
	mockFS.AssertExpectations(suite.T())
}

func (suite *UtilsTestSuite) TestWriteFile_Failure_Write() {
	mockFS := &MockFileSystem{}
	mockFile := &MockFile{}
	mockFile.On("Write", mock.Anything).Return(0, errors.New(""))
	mockFile.On("Close").Return(nil)
	mockFS.On("OpenFile", "testfile.txt", os.O_CREATE|os.O_WRONLY, os.FileMode(0644)).Return(mockFile, nil)

	data := []byte("test data")
	err := utils.WriteFile(mockFS, "testfile.txt", &data)

	suite.Assert().Error(err)
	mockFile.AssertCalled(suite.T(), "Write", data)
	mockFile.AssertCalled(suite.T(), "Close")
	mockFS.AssertCalled(suite.T(), "OpenFile", "testfile.txt", os.O_CREATE|os.O_WRONLY, os.FileMode(0644))

	mockFile.AssertExpectations(suite.T())
	mockFS.AssertExpectations(suite.T())
}

func (suite *UtilsTestSuite) TestWriteFile_Failure_Open() {
	mockFS := &MockFileSystem{}
	mockFile := &MockFile{}

	mockFS.On("OpenFile", "testfile.txt", os.O_CREATE|os.O_WRONLY, os.FileMode(0644)).Return(mockFile, errors.New(""))

	data := []byte("test data")
	err := utils.WriteFile(mockFS, "testfile.txt", &data)

	suite.Assert().Error(err)
	mockFS.AssertCalled(suite.T(), "OpenFile", "testfile.txt", os.O_CREATE|os.O_WRONLY, os.FileMode(0644))
	mockFS.AssertExpectations(suite.T())
}
