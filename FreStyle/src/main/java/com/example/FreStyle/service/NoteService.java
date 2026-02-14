package com.example.FreStyle.service;

import com.example.FreStyle.dto.NoteDto;
import jakarta.annotation.PostConstruct;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Service;
import software.amazon.awssdk.auth.credentials.AwsBasicCredentials;
import software.amazon.awssdk.auth.credentials.StaticCredentialsProvider;
import software.amazon.awssdk.regions.Region;
import software.amazon.awssdk.services.dynamodb.DynamoDbClient;
import software.amazon.awssdk.services.dynamodb.model.*;

import java.time.Instant;
import java.util.*;
import java.util.stream.Collectors;

@Service
public class NoteService {

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

    // テスト用コンストラクタ
    public NoteService(DynamoDbClient dynamoDbClient, String tableName) {
        this.dynamoDbClient = dynamoDbClient;
        this.tableName = tableName;
    }

    // Spring DI用デフォルトコンストラクタ
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

    public List<NoteDto> getNotesByUserId(Integer userId) {
        QueryRequest queryRequest = QueryRequest.builder()
                .tableName(tableName)
                .keyConditionExpression("user_id = :user_id")
                .expressionAttributeValues(Map.of(
                        ":user_id", AttributeValue.builder().n(userId.toString()).build()
                ))
                .scanIndexForward(false)
                .build();

        QueryResponse response = dynamoDbClient.query(queryRequest);

        return response.items().stream()
                .map(this::toDto)
                .collect(Collectors.toList());
    }

    public NoteDto createNote(Integer userId, String title) {
        String noteId = UUID.randomUUID().toString();
        long now = Instant.now().toEpochMilli();

        Map<String, AttributeValue> item = new HashMap<>();
        item.put("user_id", AttributeValue.builder().n(userId.toString()).build());
        item.put("note_id", AttributeValue.builder().s(noteId).build());
        item.put("title", AttributeValue.builder().s(title).build());
        item.put("content", AttributeValue.builder().s("").build());
        item.put("is_pinned", AttributeValue.builder().bool(false).build());
        item.put("created_at", AttributeValue.builder().n(String.valueOf(now)).build());
        item.put("updated_at", AttributeValue.builder().n(String.valueOf(now)).build());

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

    public void updateNote(Integer userId, String noteId, String title, String content, Boolean isPinned) {
        long now = Instant.now().toEpochMilli();

        Map<String, AttributeValue> key = new HashMap<>();
        key.put("user_id", AttributeValue.builder().n(userId.toString()).build());
        key.put("note_id", AttributeValue.builder().s(noteId).build());

        UpdateItemRequest updateRequest = UpdateItemRequest.builder()
                .tableName(tableName)
                .key(key)
                .updateExpression("SET title = :title, content = :content, is_pinned = :is_pinned, updated_at = :updated_at")
                .expressionAttributeValues(Map.of(
                        ":title", AttributeValue.builder().s(title).build(),
                        ":content", AttributeValue.builder().s(content).build(),
                        ":is_pinned", AttributeValue.builder().bool(isPinned).build(),
                        ":updated_at", AttributeValue.builder().n(String.valueOf(now)).build()
                ))
                .build();

        dynamoDbClient.updateItem(updateRequest);
    }

    public void deleteNote(Integer userId, String noteId) {
        Map<String, AttributeValue> key = new HashMap<>();
        key.put("user_id", AttributeValue.builder().n(userId.toString()).build());
        key.put("note_id", AttributeValue.builder().s(noteId).build());

        DeleteItemRequest deleteRequest = DeleteItemRequest.builder()
                .tableName(tableName)
                .key(key)
                .build();

        dynamoDbClient.deleteItem(deleteRequest);
    }

    private NoteDto toDto(Map<String, AttributeValue> item) {
        NoteDto dto = new NoteDto();
        dto.setUserId(Integer.parseInt(item.get("user_id").n()));
        dto.setNoteId(item.get("note_id").s());
        dto.setTitle(item.containsKey("title") ? item.get("title").s() : "");
        dto.setContent(item.containsKey("content") ? item.get("content").s() : "");
        dto.setIsPinned(item.containsKey("is_pinned") ? item.get("is_pinned").bool() : false);
        dto.setCreatedAt(item.containsKey("created_at") ? Long.parseLong(item.get("created_at").n()) : 0L);
        dto.setUpdatedAt(item.containsKey("updated_at") ? Long.parseLong(item.get("updated_at").n()) : 0L);
        return dto;
    }
}
