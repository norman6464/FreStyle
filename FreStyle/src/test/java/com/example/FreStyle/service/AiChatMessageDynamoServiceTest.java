package com.example.FreStyle.service;

import com.example.FreStyle.dto.AiChatMessageResponseDto;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Nested;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.ArgumentCaptor;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;
import software.amazon.awssdk.services.dynamodb.DynamoDbClient;
import software.amazon.awssdk.services.dynamodb.model.*;

import java.util.HashMap;
import java.util.List;
import java.util.Map;

import static org.assertj.core.api.Assertions.assertThat;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.Mockito.*;

@ExtendWith(MockitoExtension.class)
class AiChatMessageDynamoServiceTest {

    @Mock
    private DynamoDbClient dynamoDbClient;

    private AiChatMessageDynamoService service;

    @BeforeEach
    void setUp() {
        service = new AiChatMessageDynamoService(dynamoDbClient, "fre_style_ai_chat");
    }

    @Nested
    @DisplayName("save - メッセージ保存")
    class SaveTest {

        @Test
        @DisplayName("メッセージを保存してDTOを返す")
        void shouldSaveMessageAndReturnDto() {
            when(dynamoDbClient.putItem(any(PutItemRequest.class)))
                    .thenReturn(PutItemResponse.builder().build());

            AiChatMessageResponseDto result = service.save(1, 10, "user", "こんにちは");

            assertThat(result).isNotNull();
            assertThat(result.id()).isNotBlank();
            assertThat(result.sessionId()).isEqualTo(1);
            assertThat(result.userId()).isEqualTo(10);
            assertThat(result.role()).isEqualTo("user");
            assertThat(result.content()).isEqualTo("こんにちは");
            assertThat(result.createdAt()).isGreaterThan(0L);

            ArgumentCaptor<PutItemRequest> captor = ArgumentCaptor.forClass(PutItemRequest.class);
            verify(dynamoDbClient).putItem(captor.capture());
            PutItemRequest request = captor.getValue();
            assertThat(request.tableName()).isEqualTo("fre_style_ai_chat");
            assertThat(request.item().get("session_id").n()).isEqualTo("1");
            assertThat(request.item().get("user_id").n()).isEqualTo("10");
            assertThat(request.item().get("role").s()).isEqualTo("user");
            assertThat(request.item().get("content").s()).isEqualTo("こんにちは");
        }

        @Test
        @DisplayName("assistantロールのメッセージを保存する")
        void shouldSaveAssistantMessage() {
            when(dynamoDbClient.putItem(any(PutItemRequest.class)))
                    .thenReturn(PutItemResponse.builder().build());

            AiChatMessageResponseDto result = service.save(1, 10, "assistant", "回答です");

            assertThat(result.role()).isEqualTo("assistant");
            assertThat(result.content()).isEqualTo("回答です");
        }
    }

    @Nested
    @DisplayName("findBySessionId - セッション別メッセージ取得")
    class FindBySessionIdTest {

        @Test
        @DisplayName("セッションのメッセージを昇順で取得する")
        void shouldReturnMessagesSortedByCreatedAtAsc() {
            Map<String, AttributeValue> item1 = createMessageItem(1, "msg-1", 10, "user", "質問", 1000L);
            Map<String, AttributeValue> item2 = createMessageItem(1, "msg-2", 10, "assistant", "回答", 2000L);

            QueryResponse response = QueryResponse.builder()
                    .items(List.of(item1, item2))
                    .build();
            when(dynamoDbClient.query(any(QueryRequest.class))).thenReturn(response);

            List<AiChatMessageResponseDto> result = service.findBySessionId(1);

            assertThat(result).hasSize(2);
            assertThat(result.get(0).id()).isEqualTo("msg-1");
            assertThat(result.get(0).role()).isEqualTo("user");
            assertThat(result.get(0).content()).isEqualTo("質問");
            assertThat(result.get(1).id()).isEqualTo("msg-2");
            assertThat(result.get(1).role()).isEqualTo("assistant");

            ArgumentCaptor<QueryRequest> captor = ArgumentCaptor.forClass(QueryRequest.class);
            verify(dynamoDbClient).query(captor.capture());
            QueryRequest request = captor.getValue();
            assertThat(request.tableName()).isEqualTo("fre_style_ai_chat");
            assertThat(request.scanIndexForward()).isTrue();
        }

        @Test
        @DisplayName("メッセージが存在しない場合は空リストを返す")
        void shouldReturnEmptyListWhenNoMessages() {
            QueryResponse response = QueryResponse.builder()
                    .items(List.of())
                    .build();
            when(dynamoDbClient.query(any(QueryRequest.class))).thenReturn(response);

            List<AiChatMessageResponseDto> result = service.findBySessionId(999);

            assertThat(result).isEmpty();
        }
    }

    @Nested
    @DisplayName("findByUserId - ユーザー別メッセージ取得")
    class FindByUserIdTest {

        @Test
        @DisplayName("ユーザーの全メッセージをGSI経由で取得する")
        void shouldReturnMessagesByUserIdUsingGSI() {
            Map<String, AttributeValue> item1 = createMessageItem(1, "msg-1", 10, "user", "質問1", 1000L);
            Map<String, AttributeValue> item2 = createMessageItem(2, "msg-2", 10, "user", "質問2", 2000L);

            QueryResponse response = QueryResponse.builder()
                    .items(List.of(item1, item2))
                    .build();
            when(dynamoDbClient.query(any(QueryRequest.class))).thenReturn(response);

            List<AiChatMessageResponseDto> result = service.findByUserId(10);

            assertThat(result).hasSize(2);
            assertThat(result.get(0).userId()).isEqualTo(10);
            assertThat(result.get(1).userId()).isEqualTo(10);

            ArgumentCaptor<QueryRequest> captor = ArgumentCaptor.forClass(QueryRequest.class);
            verify(dynamoDbClient).query(captor.capture());
            QueryRequest request = captor.getValue();
            assertThat(request.indexName()).isEqualTo("user_id-created_at-index");
            assertThat(request.scanIndexForward()).isTrue();
        }
    }

    @Nested
    @DisplayName("countBySessionId - メッセージ数カウント")
    class CountBySessionIdTest {

        @Test
        @DisplayName("セッションのメッセージ数を返す")
        void shouldReturnMessageCount() {
            QueryResponse response = QueryResponse.builder()
                    .count(5)
                    .build();
            when(dynamoDbClient.query(any(QueryRequest.class))).thenReturn(response);

            Long count = service.countBySessionId(1);

            assertThat(count).isEqualTo(5L);

            ArgumentCaptor<QueryRequest> captor = ArgumentCaptor.forClass(QueryRequest.class);
            verify(dynamoDbClient).query(captor.capture());
            QueryRequest request = captor.getValue();
            assertThat(request.select()).isEqualTo(Select.COUNT);
        }
    }

    @Nested
    @DisplayName("deleteBySessionId - セッション別メッセージ削除")
    class DeleteBySessionIdTest {

        @Test
        @DisplayName("セッションの全メッセージを削除する")
        void shouldDeleteAllMessagesInSession() {
            Map<String, AttributeValue> item1 = createMessageItem(1, "msg-1", 10, "user", "質問", 1000L);
            Map<String, AttributeValue> item2 = createMessageItem(1, "msg-2", 10, "assistant", "回答", 2000L);

            QueryResponse queryResponse = QueryResponse.builder()
                    .items(List.of(item1, item2))
                    .build();
            when(dynamoDbClient.query(any(QueryRequest.class))).thenReturn(queryResponse);
            when(dynamoDbClient.batchWriteItem(any(BatchWriteItemRequest.class)))
                    .thenReturn(BatchWriteItemResponse.builder().build());

            service.deleteBySessionId(1);

            verify(dynamoDbClient).query(any(QueryRequest.class));
            ArgumentCaptor<BatchWriteItemRequest> captor = ArgumentCaptor.forClass(BatchWriteItemRequest.class);
            verify(dynamoDbClient).batchWriteItem(captor.capture());
            BatchWriteItemRequest request = captor.getValue();
            assertThat(request.requestItems().get("fre_style_ai_chat")).hasSize(2);
        }

        @Test
        @DisplayName("メッセージが存在しない場合はバッチ削除を実行しない")
        void shouldNotCallBatchDeleteWhenNoMessages() {
            QueryResponse queryResponse = QueryResponse.builder()
                    .items(List.of())
                    .build();
            when(dynamoDbClient.query(any(QueryRequest.class))).thenReturn(queryResponse);

            service.deleteBySessionId(999);

            verify(dynamoDbClient).query(any(QueryRequest.class));
            verify(dynamoDbClient, never()).batchWriteItem(any(BatchWriteItemRequest.class));
        }
    }

    private Map<String, AttributeValue> createMessageItem(
            Integer sessionId, String messageId, Integer userId,
            String role, String content, Long createdAt) {
        Map<String, AttributeValue> item = new HashMap<>();
        item.put("session_id", AttributeValue.builder().n(sessionId.toString()).build());
        item.put("message_id", AttributeValue.builder().s(messageId).build());
        item.put("user_id", AttributeValue.builder().n(userId.toString()).build());
        item.put("role", AttributeValue.builder().s(role).build());
        item.put("content", AttributeValue.builder().s(content).build());
        item.put("created_at", AttributeValue.builder().n(createdAt.toString()).build());
        return item;
    }
}
