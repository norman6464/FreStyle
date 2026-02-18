package com.example.FreStyle.service;

import java.time.Instant;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.UUID;

import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Repository;

import com.example.FreStyle.dto.AiChatMessageResponseDto;
import com.example.FreStyle.repository.AiChatMessageDynamoRepository;

import jakarta.annotation.PostConstruct;
import software.amazon.awssdk.auth.credentials.AwsBasicCredentials;
import software.amazon.awssdk.auth.credentials.StaticCredentialsProvider;
import software.amazon.awssdk.regions.Region;
import software.amazon.awssdk.services.dynamodb.DynamoDbClient;
import software.amazon.awssdk.services.dynamodb.model.AttributeValue;
import software.amazon.awssdk.services.dynamodb.model.BatchWriteItemRequest;
import software.amazon.awssdk.services.dynamodb.model.DeleteRequest;
import software.amazon.awssdk.services.dynamodb.model.PutItemRequest;
import software.amazon.awssdk.services.dynamodb.model.QueryRequest;
import software.amazon.awssdk.services.dynamodb.model.QueryResponse;
import software.amazon.awssdk.services.dynamodb.model.Select;
import software.amazon.awssdk.services.dynamodb.model.WriteRequest;

@Repository
public class AiChatMessageDynamoService implements AiChatMessageDynamoRepository {

    private static final String ATTR_SESSION_ID = "session_id";
    private static final String ATTR_CREATED_AT = "created_at";
    private static final String ATTR_MESSAGE_ID = "message_id";
    private static final String ATTR_USER_ID = "user_id";
    private static final String ATTR_ROLE = "role";
    private static final String ATTR_CONTENT = "content";

    private static final String GSI_USER_ID = "user_id-created_at-index";

    private DynamoDbClient dynamoDbClient;
    private String tableName;

    @Value("${aws.access-key:}")
    private String accessKey;

    @Value("${aws.secret-key:}")
    private String secretKey;

    @Value("${aws.region:}")
    private String region;

    @Value("${aws.dynamodb.table-name.ai-chat:fre_style_ai_chat}")
    private String configTableName;

    // テスト用コンストラクタ
    public AiChatMessageDynamoService(DynamoDbClient dynamoDbClient, String tableName) {
        this.dynamoDbClient = dynamoDbClient;
        this.tableName = tableName;
    }

    public AiChatMessageDynamoService() {
    }

    @PostConstruct
    public void init() {
        if (dynamoDbClient == null && !accessKey.isEmpty()) {
            dynamoDbClient = DynamoDbClient.builder()
                    .region(Region.of(region))
                    .credentialsProvider(
                            StaticCredentialsProvider.create(
                                    AwsBasicCredentials.create(accessKey, secretKey)
                            )
                    )
                    .build();
        }
        if (tableName == null) {
            tableName = configTableName;
        }
    }

    @Override
    public AiChatMessageResponseDto save(Integer sessionId, Integer userId, String role, String content) {
        String messageId = UUID.randomUUID().toString();
        long now = Instant.now().toEpochMilli();

        Map<String, AttributeValue> item = new HashMap<>();
        item.put(ATTR_SESSION_ID, AttributeValue.builder().n(sessionId.toString()).build());
        item.put(ATTR_CREATED_AT, AttributeValue.builder().n(String.valueOf(now)).build());
        item.put(ATTR_MESSAGE_ID, AttributeValue.builder().s(messageId).build());
        item.put(ATTR_USER_ID, AttributeValue.builder().n(userId.toString()).build());
        item.put(ATTR_ROLE, AttributeValue.builder().s(role).build());
        item.put(ATTR_CONTENT, AttributeValue.builder().s(content).build());

        PutItemRequest putRequest = PutItemRequest.builder()
                .tableName(tableName)
                .item(item)
                .build();

        dynamoDbClient.putItem(putRequest);

        return new AiChatMessageResponseDto(messageId, sessionId, userId, role, content, now);
    }

    @Override
    public List<AiChatMessageResponseDto> findBySessionId(Integer sessionId) {
        QueryRequest queryRequest = QueryRequest.builder()
                .tableName(tableName)
                .keyConditionExpression(ATTR_SESSION_ID + " = :session_id")
                .expressionAttributeValues(Map.of(
                        ":session_id", AttributeValue.builder().n(sessionId.toString()).build()
                ))
                .scanIndexForward(true)
                .build();

        QueryResponse response = dynamoDbClient.query(queryRequest);

        return response.items().stream()
                .map(this::toDto)
                .toList();
    }

    @Override
    public List<AiChatMessageResponseDto> findByUserId(Integer userId) {
        QueryRequest queryRequest = QueryRequest.builder()
                .tableName(tableName)
                .indexName(GSI_USER_ID)
                .keyConditionExpression(ATTR_USER_ID + " = :user_id")
                .expressionAttributeValues(Map.of(
                        ":user_id", AttributeValue.builder().n(userId.toString()).build()
                ))
                .scanIndexForward(true)
                .build();

        QueryResponse response = dynamoDbClient.query(queryRequest);

        return response.items().stream()
                .map(this::toDto)
                .toList();
    }

    @Override
    public Long countBySessionId(Integer sessionId) {
        QueryRequest queryRequest = QueryRequest.builder()
                .tableName(tableName)
                .keyConditionExpression(ATTR_SESSION_ID + " = :session_id")
                .expressionAttributeValues(Map.of(
                        ":session_id", AttributeValue.builder().n(sessionId.toString()).build()
                ))
                .select(Select.COUNT)
                .build();

        QueryResponse response = dynamoDbClient.query(queryRequest);
        return (long) response.count();
    }

    @Override
    public void deleteBySessionId(Integer sessionId) {
        // まずセッションの全メッセージを取得
        QueryRequest queryRequest = QueryRequest.builder()
                .tableName(tableName)
                .keyConditionExpression(ATTR_SESSION_ID + " = :session_id")
                .expressionAttributeValues(Map.of(
                        ":session_id", AttributeValue.builder().n(sessionId.toString()).build()
                ))
                .projectionExpression(ATTR_SESSION_ID + ", " + ATTR_CREATED_AT)
                .build();

        QueryResponse response = dynamoDbClient.query(queryRequest);
        List<Map<String, AttributeValue>> items = response.items();

        if (items.isEmpty()) {
            return;
        }

        // BatchWriteItem (最大25件ずつ)
        List<WriteRequest> deleteRequests = new ArrayList<>();
        for (Map<String, AttributeValue> item : items) {
            Map<String, AttributeValue> key = new HashMap<>();
            key.put(ATTR_SESSION_ID, item.get(ATTR_SESSION_ID));
            key.put(ATTR_CREATED_AT, item.get(ATTR_CREATED_AT));

            deleteRequests.add(WriteRequest.builder()
                    .deleteRequest(DeleteRequest.builder().key(key).build())
                    .build());
        }

        // 25件ずつバッチ処理
        for (int i = 0; i < deleteRequests.size(); i += 25) {
            List<WriteRequest> batch = deleteRequests.subList(i, Math.min(i + 25, deleteRequests.size()));
            BatchWriteItemRequest batchRequest = BatchWriteItemRequest.builder()
                    .requestItems(Map.of(tableName, batch))
                    .build();
            dynamoDbClient.batchWriteItem(batchRequest);
        }
    }

    private AiChatMessageResponseDto toDto(Map<String, AttributeValue> item) {
        return new AiChatMessageResponseDto(
                item.get(ATTR_MESSAGE_ID).s(),
                Integer.parseInt(item.get(ATTR_SESSION_ID).n()),
                Integer.parseInt(item.get(ATTR_USER_ID).n()),
                item.get(ATTR_ROLE).s(),
                getStringAttr(item, ATTR_CONTENT, ""),
                Long.parseLong(item.get(ATTR_CREATED_AT).n())
        );
    }

    private String getStringAttr(Map<String, AttributeValue> item, String key, String defaultValue) {
        return item.containsKey(key) ? item.get(key).s() : defaultValue;
    }
}
