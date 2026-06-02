package persistence

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
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// aiChatMessageRepository は [repository.AiChatMessageRepository] の DynamoDB 実装。
type aiChatMessageRepository struct {
	svc   *dynamodb.Client
	table string
}

// NewAiChatMessageRepository は DynamoDB クライアントを組み立てて返す。
func NewAiChatMessageRepository(ctx context.Context, region, table string) (repository.AiChatMessageRepository, error) {
	cfg, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("dynamodb: load aws config: %w", err)
	}
	return &aiChatMessageRepository{
		svc:   dynamodb.NewFromConfig(cfg),
		table: table,
	}, nil
}

// dynamoItem は DynamoDB に保存するアイテム形式。BlobData は永続化せず、
// attachments 列が無い既存メッセージのため omitempty で扱う。
type dynamoItem struct {
	SessionID   string         `dynamodbav:"sessionId"`
	MessageID   string         `dynamodbav:"messageId"`
	Role        string         `dynamodbav:"role"`
	Content     string         `dynamodbav:"content"`
	Attachments []dynamoAttach `dynamodbav:"attachments,omitempty"`
	CreatedAt   string         `dynamodbav:"createdAt"`
}

type dynamoAttach struct {
	Key         string `dynamodbav:"key"`
	Filename    string `dynamodbav:"filename"`
	ContentType string `dynamodbav:"contentType"`
	Format      string `dynamodbav:"format"`
	Kind        string `dynamodbav:"kind"`
	SizeBytes   int64  `dynamodbav:"sizeBytes"`
}

func toDynamoAttachments(in []domain.Attachment) []dynamoAttach {
	if len(in) == 0 {
		return nil
	}
	out := make([]dynamoAttach, len(in))
	for i, a := range in {
		out[i] = dynamoAttach{
			Key:         a.Key,
			Filename:    a.Filename,
			ContentType: a.ContentType,
			Format:      a.Format,
			Kind:        a.Kind,
			SizeBytes:   a.SizeBytes,
		}
	}
	return out
}

func fromDynamoAttachments(in []dynamoAttach) []domain.Attachment {
	if len(in) == 0 {
		return nil
	}
	out := make([]domain.Attachment, len(in))
	for i, a := range in {
		out[i] = domain.Attachment{
			Key:         a.Key,
			Filename:    a.Filename,
			ContentType: a.ContentType,
			Format:      a.Format,
			Kind:        a.Kind,
			SizeBytes:   a.SizeBytes,
		}
	}
	return out
}

func (r *aiChatMessageRepository) Save(ctx context.Context, msg *domain.AiChatMessage) error {
	item := dynamoItem{
		SessionID:   strconv.FormatUint(msg.SessionID, 10),
		MessageID:   msg.MessageID,
		Role:        msg.Role,
		Content:     msg.Content,
		Attachments: toDynamoAttachments(msg.Attachments),
		CreatedAt:   msg.CreatedAt.UTC().Format(time.RFC3339),
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
			SessionID:   sessionID,
			MessageID:   item.MessageID,
			Role:        item.Role,
			Content:     item.Content,
			Attachments: fromDynamoAttachments(item.Attachments),
			CreatedAt:   t,
		})
	}
	return msgs, nil
}
