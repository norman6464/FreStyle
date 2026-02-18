package com.example.FreStyle.service;

import java.time.Instant;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.UUID;
import java.util.stream.Collectors;

import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Repository;

import com.example.FreStyle.dto.ChatMessageDto;
import com.example.FreStyle.repository.ChatMessageDynamoRepository;

import jakarta.annotation.PostConstruct;
import software.amazon.awssdk.auth.credentials.AwsBasicCredentials;
import software.amazon.awssdk.auth.credentials.StaticCredentialsProvider;
import software.amazon.awssdk.regions.Region;
import software.amazon.awssdk.services.dynamodb.DynamoDbClient;
import software.amazon.awssdk.services.dynamodb.model.AttributeValue;
import software.amazon.awssdk.services.dynamodb.model.DeleteItemRequest;
import software.amazon.awssdk.services.dynamodb.model.PutItemRequest;
import software.amazon.awssdk.services.dynamodb.model.QueryRequest;
import software.amazon.awssdk.services.dynamodb.model.QueryResponse;

@Repository
public class ChatMessageDynamoService implements ChatMessageDynamoRepository {

    private static final String ATTR_ROOM_ID = "room_id";
    private static final String ATTR_CREATED_AT = "created_at";
    private static final String ATTR_MESSAGE_ID = "message_id";
    private static final String ATTR_SENDER_ID = "sender_id";
    private static final String ATTR_CONTENT = "content";

    private DynamoDbClient dynamoDbClient;
    private String tableName;

    @Value("${aws.access-key:}")
    private String accessKey;

    @Value("${aws.secret-key:}")
    private String secretKey;

    @Value("${aws.region:}")
    private String region;

    @Value("${aws.dynamodb.table-name.chat:fre_style_chat}")
    private String configTableName;

    // テスト用コンストラクタ
    public ChatMessageDynamoService(DynamoDbClient dynamoDbClient, String tableName) {
        this.dynamoDbClient = dynamoDbClient;
        this.tableName = tableName;
    }

    public ChatMessageDynamoService() {
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
    public ChatMessageDto save(Integer roomId, Integer senderId, String content) {
        String messageId = UUID.randomUUID().toString();
        long now = Instant.now().toEpochMilli();

        Map<String, AttributeValue> item = new HashMap<>();
        item.put(ATTR_ROOM_ID, AttributeValue.builder().n(roomId.toString()).build());
        item.put(ATTR_CREATED_AT, AttributeValue.builder().n(String.valueOf(now)).build());
        item.put(ATTR_MESSAGE_ID, AttributeValue.builder().s(messageId).build());
        item.put(ATTR_SENDER_ID, AttributeValue.builder().n(senderId.toString()).build());
        item.put(ATTR_CONTENT, AttributeValue.builder().s(content).build());

        PutItemRequest putRequest = PutItemRequest.builder()
                .tableName(tableName)
                .item(item)
                .build();

        dynamoDbClient.putItem(putRequest);

        // senderNameはDynamoDBでは保持しない（サービス層で解決）
        return new ChatMessageDto(messageId, roomId, senderId, null, content, now);
    }

    @Override
    public List<ChatMessageDto> findByRoomId(Integer roomId) {
        QueryRequest queryRequest = QueryRequest.builder()
                .tableName(tableName)
                .keyConditionExpression(ATTR_ROOM_ID + " = :room_id")
                .expressionAttributeValues(Map.of(
                        ":room_id", AttributeValue.builder().n(roomId.toString()).build()
                ))
                .scanIndexForward(true)
                .build();

        QueryResponse response = dynamoDbClient.query(queryRequest);

        return response.items().stream()
                .map(this::toDto)
                .toList();
    }

    @Override
    public void deleteByRoomIdAndCreatedAt(Integer roomId, Long createdAt) {
        Map<String, AttributeValue> key = new HashMap<>();
        key.put(ATTR_ROOM_ID, AttributeValue.builder().n(roomId.toString()).build());
        key.put(ATTR_CREATED_AT, AttributeValue.builder().n(createdAt.toString()).build());

        DeleteItemRequest deleteRequest = DeleteItemRequest.builder()
                .tableName(tableName)
                .key(key)
                .build();

        dynamoDbClient.deleteItem(deleteRequest);
    }

    @Override
    public ChatMessageDto findLatestByRoomId(Integer roomId) {
        QueryRequest queryRequest = QueryRequest.builder()
                .tableName(tableName)
                .keyConditionExpression(ATTR_ROOM_ID + " = :room_id")
                .expressionAttributeValues(Map.of(
                        ":room_id", AttributeValue.builder().n(roomId.toString()).build()
                ))
                .scanIndexForward(false)
                .limit(1)
                .build();

        QueryResponse response = dynamoDbClient.query(queryRequest);

        if (response.items().isEmpty()) {
            return null;
        }

        return toDto(response.items().get(0));
    }

    @Override
    public Map<Integer, ChatMessageDto> findLatestByRoomIds(List<Integer> roomIds) {
        Map<Integer, ChatMessageDto> result = new HashMap<>();
        for (Integer roomId : roomIds) {
            ChatMessageDto latest = findLatestByRoomId(roomId);
            if (latest != null) {
                result.put(roomId, latest);
            }
        }
        return result;
    }

    private ChatMessageDto toDto(Map<String, AttributeValue> item) {
        return new ChatMessageDto(
                item.get(ATTR_MESSAGE_ID).s(),
                Integer.parseInt(item.get(ATTR_ROOM_ID).n()),
                Integer.parseInt(item.get(ATTR_SENDER_ID).n()),
                null, // senderNameはDynamoDBでは保持しない
                getStringAttr(item, ATTR_CONTENT, ""),
                Long.parseLong(item.get(ATTR_CREATED_AT).n())
        );
    }

    private String getStringAttr(Map<String, AttributeValue> item, String key, String defaultValue) {
        return item.containsKey(key) ? item.get(key).s() : defaultValue;
    }
}
