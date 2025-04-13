package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/rekognition"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// RekognitionService Rekognitionサービスを管理する構造体
type RekognitionService struct {
	client   *rekognition.Client
	s3Client *s3.Client
}

// NewRekognitionService RekognitionServiceの新しいインスタンスを作成
func NewRekognitionService() (*RekognitionService, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, fmt.Errorf("SDKの設定を読み込めません: %v", err)
	}

	return &RekognitionService{
		client:   rekognition.NewFromConfig(cfg),
		s3Client: s3.NewFromConfig(cfg),
	}, nil
}

// CompareFaces 2つの画像の顔を比較し、一致するかどうかを判定
func (r *RekognitionService) CompareFaces(sourceImage string, targetImage string) (bool, error) {
	input := &rekognition.CompareFacesInput{
		SourceImage: &rekognition.Image{
			S3Object: &rekognition.S3Object{
				Bucket: aws.String(os.Getenv("S3_BUCKET")),
				Name:   aws.String(sourceImage),
			},
		},
		TargetImage: &rekognition.Image{
			S3Object: &rekognition.S3Object{
				Bucket: aws.String(os.Getenv("S3_BUCKET")),
				Name:   aws.String(targetImage),
			},
		},
		SimilarityThreshold: aws.Float32(80.0), // 類似度の閾値を80%に設定
	}

	result, err := r.client.CompareFaces(context.TODO(), input)
	if err != nil {
		return false, fmt.Errorf("顔の比較に失敗しました: %v", err)
	}

	return len(result.FaceMatches) > 0, nil
}

func main() {
	service, err := NewRekognitionService()
	if err != nil {
		log.Fatalf("Rekognitionサービスの作成に失敗しました: %v", err)
	}

	// コマンドライン引数から画像パスを取得
	if len(os.Args) != 3 {
		log.Fatal("使用方法: go run main.go <source_image> <target_image>")
	}

	sourceImage := os.Args[1]
	targetImage := os.Args[2]

	matched, err := service.CompareFaces(sourceImage, targetImage)
	if err != nil {
		log.Fatalf("顔の比較中にエラーが発生しました: %v", err)
	}

	if matched {
		fmt.Println("顔が一致しました！")
	} else {
		fmt.Println("顔が一致しませんでした。")
	}
}
