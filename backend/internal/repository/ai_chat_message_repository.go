package repository

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	dbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/norman6464/FreStyle/backend/internal/domain"
)

// AiChatMessageRepository は DynamoDB 上の AI チャットメッセージへのアクセスを提供する。
// PK: sessionId (String)、SK: messageId (UUID String)
type AiChatMessageRepository interface {
	Save(ctx context.Context, msg *domain.AiChatMessage) error
	ListBySessionID(ctx context.Context, sessionID uint64) ([]domain.AiChatMessage, error)
}

type aiChatMessageRepository struct {
	svc   *dynamodb.Client
	table string
}

// NewAiChatMessageRepository は DynamoDB クライアントを組み立てて repository を返す。
func NewAiChatMessageRepository(ctx context.Context, region, table string) (AiChatMessageRepository, error) {
	cfg, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("dynamodb: load aws config: %w", err)
	}
	return &aiChatMessageRepository{
		svc:   dynamodb.NewFromConfig(cfg),
		table: table,
	}, nil
}

// dynamoItem は DynamoDB に保存するアイテムの形式。
type dynamoItem struct {
	SessionID string `dynamodbav:"sessionId"`
	MessageID string `dynamodbav:"messageId"`
	Role      string `dynamodbav:"role"`
	Content   string `dynamodbav:"content"`
	CreatedAt string `dynamodbav:"createdAt"`
}

func (r *aiChatMessageRepository) Save(ctx context.Context, msg *domain.AiChatMessage) error {
	item := dynamoItem{
		SessionID: strconv.FormatUint(msg.SessionID, 10),
		MessageID: msg.MessageID,
		Role:      msg.Role,
		Content:   msg.Content,
		CreatedAt: msg.CreatedAt.UTC().Format(time.RFC3339),
	}
	av, err := attributevalue.MarshalMap(item)
	if err != nil {
		return fmt.Errorf("dynamodb: marshal message: %w", err)
	}
	_, err = r.svc.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(r.table),
		Item:      av,
	})
	return err
}

func (r *aiChatMessageRepository) ListBySessionID(ctx context.Context, sessionID uint64) ([]domain.AiChatMessage, error) {
	resp, err := r.svc.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(r.table),
		KeyConditionExpression: aws.String("sessionId = :sid"),
		ExpressionAttributeValues: map[string]dbtypes.AttributeValue{
			":sid": &dbtypes.AttributeValueMemberS{Value: strconv.FormatUint(sessionID, 10)},
		},
		ScanIndexForward: aws.Bool(true),
	})
	if err != nil {
		return nil, fmt.Errorf("dynamodb: query messages: %w", err)
	}

	msgs := make([]domain.AiChatMessage, 0, len(resp.Items))
	for _, raw := range resp.Items {
		var item dynamoItem
		if err := attributevalue.UnmarshalMap(raw, &item); err != nil {
			continue
		}
		t, _ := time.Parse(time.RFC3339, item.CreatedAt)
		msgs = append(msgs, domain.AiChatMessage{
			SessionID: sessionID,
			MessageID: item.MessageID,
			Role:      item.Role,
			Content:   item.Content,
			CreatedAt: t,
		})
	}
	return msgs, nil
}
