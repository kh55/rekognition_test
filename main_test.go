package main

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/rekognition"
	"github.com/aws/aws-sdk-go-v2/service/rekognition/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// MockRekognitionClient Rekognitionのモッククライアント
type MockRekognitionClient struct {
	CompareFacesFunc func(ctx context.Context, params *rekognition.CompareFacesInput, optFns ...func(*rekognition.Options)) (*rekognition.CompareFacesOutput, error)
}

func (m *MockRekognitionClient) CompareFaces(ctx context.Context, params *rekognition.CompareFacesInput, optFns ...func(*rekognition.Options)) (*rekognition.CompareFacesOutput, error) {
	return m.CompareFacesFunc(ctx, params, optFns...)
}

// MockS3Client S3のモッククライアント
type MockS3Client struct{}

func (m *MockS3Client) GetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
	return &s3.GetObjectOutput{}, nil
}

// TestCompareFaces 顔比較のテスト
func TestCompareFaces(t *testing.T) {
	tests := []struct {
		name          string
		sourceImage   string
		targetImage   string
		expectedMatch bool
		expectedError bool
		mockResponse  *rekognition.CompareFacesOutput
		mockError     error
	}{
		{
			name:          "顔が一致する場合",
			sourceImage:   "test/source.jpg",
			targetImage:   "test/target.jpg",
			expectedMatch: true,
			expectedError: false,
			mockResponse: &rekognition.CompareFacesOutput{
				FaceMatches: []types.CompareFacesMatch{
					{
						Similarity: 90.0,
					},
				},
			},
			mockError: nil,
		},
		{
			name:          "顔が一致しない場合",
			sourceImage:   "test/source.jpg",
			targetImage:   "test/target.jpg",
			expectedMatch: false,
			expectedError: false,
			mockResponse: &rekognition.CompareFacesOutput{
				FaceMatches: []types.CompareFacesMatch{},
			},
			mockError: nil,
		},
		{
			name:          "エラーが発生する場合",
			sourceImage:   "test/source.jpg",
			targetImage:   "test/target.jpg",
			expectedMatch: false,
			expectedError: true,
			mockResponse:  nil,
			mockError:     fmt.Errorf("APIエラー"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// モッククライアントの設定
			mockRekognition := &MockRekognitionClient{
				CompareFacesFunc: func(ctx context.Context, params *rekognition.CompareFacesInput, optFns ...func(*rekognition.Options)) (*rekognition.CompareFacesOutput, error) {
					return tt.mockResponse, tt.mockError
				},
			}

			service := &RekognitionService{
				client:   mockRekognition,
				s3Client: &MockS3Client{},
			}

			// テスト実行
			matched, err := service.CompareFaces(tt.sourceImage, tt.targetImage)

			// エラーチェック
			if tt.expectedError {
				if err == nil {
					t.Errorf("エラーが期待されましたが、発生しませんでした")
				}
				return
			}

			if err != nil {
				t.Errorf("予期しないエラーが発生しました: %v", err)
				return
			}

			// 結果のチェック
			if matched != tt.expectedMatch {
				t.Errorf("期待される結果: %v, 実際の結果: %v", tt.expectedMatch, matched)
			}
		})
	}
}

// TestNewRekognitionService RekognitionServiceの初期化テスト
func TestNewRekognitionService(t *testing.T) {
	service, err := NewRekognitionService()
	if err != nil {
		t.Errorf("RekognitionServiceの初期化に失敗しました: %v", err)
	}

	if service == nil {
		t.Error("RekognitionServiceがnilです")
	}

	if service.client == nil {
		t.Error("Rekognitionクライアントがnilです")
	}

	if service.s3Client == nil {
		t.Error("S3クライアントがnilです")
	}
}
