package com.example.FreStyle.service;

import com.example.FreStyle.dto.NoteDto;
import com.example.FreStyle.repository.NoteRepository;
import jakarta.annotation.PostConstruct;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Repository;
import software.amazon.awssdk.auth.credentials.AwsBasicCredentials;
import software.amazon.awssdk.auth.credentials.StaticCredentialsProvider;
import software.amazon.awssdk.regions.Region;
import software.amazon.awssdk.services.dynamodb.DynamoDbClient;
import software.amazon.awssdk.services.dynamodb.model.*;

import java.time.Instant;
import java.util.*;

@Repository
public class NoteService implements NoteRepository {

    private static final String ATTR_USER_ID = "user_id";
    private static final String ATTR_NOTE_ID = "note_id";
    private static final String ATTR_TITLE = "title";
    private static final String ATTR_CONTENT = "content";
    private static final String ATTR_IS_PINNED = "is_pinned";
    private static final String ATTR_CREATED_AT = "created_at";
    private static final String ATTR_UPDATED_AT = "updated_at";

    private DynamoDbClient dynamoDbClient;
    private String tableName;

    @Value("${aws.access-key:}")
    private String accessKey;

    @Value("${aws.secret-key:}")
    private String secretKey;

    @Value("${aws.region:}")
    private String region;

    @Value("${aws.dynamodb.table-name.notes:fre_style_notes}")
    private String configTableName;

    public NoteService(DynamoDbClient dynamoDbClient, String tableName) {
        this.dynamoDbClient = dynamoDbClient;
        this.tableName = tableName;
    }

    public NoteService() {
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
    public List<NoteDto> findByUserId(Integer userId) {
        QueryRequest queryRequest = QueryRequest.builder()
                .tableName(tableName)
                .keyConditionExpression(ATTR_USER_ID + " = :user_id")
                .expressionAttributeValues(Map.of(
                        ":user_id", AttributeValue.builder().n(userId.toString()).build()
                ))
                .scanIndexForward(false)
                .build();

        QueryResponse response = dynamoDbClient.query(queryRequest);

        return response.items().stream()
                .map(this::toDto)
                .toList();
    }

    @Override
    public NoteDto save(Integer userId, String title) {
        String noteId = UUID.randomUUID().toString();
        long now = Instant.now().toEpochMilli();

        Map<String, AttributeValue> item = new HashMap<>();
        item.put(ATTR_USER_ID, AttributeValue.builder().n(userId.toString()).build());
        item.put(ATTR_NOTE_ID, AttributeValue.builder().s(noteId).build());
        item.put(ATTR_TITLE, AttributeValue.builder().s(title).build());
        item.put(ATTR_CONTENT, AttributeValue.builder().s("").build());
        item.put(ATTR_IS_PINNED, AttributeValue.builder().bool(false).build());
        item.put(ATTR_CREATED_AT, AttributeValue.builder().n(String.valueOf(now)).build());
        item.put(ATTR_UPDATED_AT, AttributeValue.builder().n(String.valueOf(now)).build());

        PutItemRequest putRequest = PutItemRequest.builder()
                .tableName(tableName)
                .item(item)
                .build();

        dynamoDbClient.putItem(putRequest);

        NoteDto dto = new NoteDto();
        dto.setNoteId(noteId);
        dto.setUserId(userId);
        dto.setTitle(title);
        dto.setContent("");
        dto.setIsPinned(false);
        dto.setCreatedAt(now);
        dto.setUpdatedAt(now);
        return dto;
    }

    @Override
    public void update(Integer userId, String noteId, String title, String content, Boolean isPinned) {
        long now = Instant.now().toEpochMilli();

        UpdateItemRequest updateRequest = UpdateItemRequest.builder()
                .tableName(tableName)
                .key(buildKey(userId, noteId))
                .updateExpression("SET " + ATTR_TITLE + " = :title, " + ATTR_CONTENT + " = :content, " + ATTR_IS_PINNED + " = :is_pinned, " + ATTR_UPDATED_AT + " = :updated_at")
                .expressionAttributeValues(Map.of(
                        ":title", AttributeValue.builder().s(title).build(),
                        ":content", AttributeValue.builder().s(content).build(),
                        ":is_pinned", AttributeValue.builder().bool(isPinned).build(),
                        ":updated_at", AttributeValue.builder().n(String.valueOf(now)).build()
                ))
                .build();

        dynamoDbClient.updateItem(updateRequest);
    }

    @Override
    public void delete(Integer userId, String noteId) {
        DeleteItemRequest deleteRequest = DeleteItemRequest.builder()
                .tableName(tableName)
                .key(buildKey(userId, noteId))
                .build();

        dynamoDbClient.deleteItem(deleteRequest);
    }

    private Map<String, AttributeValue> buildKey(Integer userId, String noteId) {
        Map<String, AttributeValue> key = new HashMap<>();
        key.put(ATTR_USER_ID, AttributeValue.builder().n(userId.toString()).build());
        key.put(ATTR_NOTE_ID, AttributeValue.builder().s(noteId).build());
        return key;
    }

    private NoteDto toDto(Map<String, AttributeValue> item) {
        NoteDto dto = new NoteDto();
        dto.setUserId(Integer.parseInt(item.get(ATTR_USER_ID).n()));
        dto.setNoteId(item.get(ATTR_NOTE_ID).s());
        dto.setTitle(getStringAttr(item, ATTR_TITLE, ""));
        dto.setContent(getStringAttr(item, ATTR_CONTENT, ""));
        dto.setIsPinned(item.containsKey(ATTR_IS_PINNED) ? item.get(ATTR_IS_PINNED).bool() : false);
        dto.setCreatedAt(getLongAttr(item, ATTR_CREATED_AT, 0L));
        dto.setUpdatedAt(getLongAttr(item, ATTR_UPDATED_AT, 0L));
        return dto;
    }

    private String getStringAttr(Map<String, AttributeValue> item, String key, String defaultValue) {
        return item.containsKey(key) ? item.get(key).s() : defaultValue;
    }

    private long getLongAttr(Map<String, AttributeValue> item, String key, long defaultValue) {
        return item.containsKey(key) ? Long.parseLong(item.get(key).n()) : defaultValue;
    }
}
