package com.example.FreStyle.service;

import static org.assertj.core.api.Assertions.assertThat;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.Mockito.verify;
import static org.mockito.Mockito.when;

import java.lang.reflect.Field;
import java.util.List;
import java.util.Map;

import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Nested;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.ArgumentCaptor;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import com.example.FreStyle.dto.AiChatMessageDto;

import software.amazon.awssdk.services.dynamodb.DynamoDbClient;
import software.amazon.awssdk.services.dynamodb.model.AttributeValue;
import software.amazon.awssdk.services.dynamodb.model.QueryRequest;
import software.amazon.awssdk.services.dynamodb.model.QueryResponse;

@ExtendWith(MockitoExtension.class)
@DisplayName("AiChatService")
class AiChatServiceTest {

    @Mock
    private DynamoDbClient dynamoDbClient;

    private AiChatService aiChatService;

    @BeforeEach
    void setUp() throws Exception {
        aiChatService = new AiChatService();
        Field field = AiChatService.class.getDeclaredField("dynamoDbClient");
        field.setAccessible(true);
        field.set(aiChatService, dynamoDbClient);

        Field tableField = AiChatService.class.getDeclaredField("tableName");
        tableField.setAccessible(true);
        tableField.set(aiChatService, "test-ai-chat-table");
    }

    @Nested
    @DisplayName("getChatHistory")
    class GetChatHistory {

        @Test
        @DisplayName("チャット履歴を取得する")
        void returnsChatHistory() {
            Map<String, AttributeValue> item = Map.of(
                "content", AttributeValue.builder().s("テストメッセージ").build(),
                "is_user", AttributeValue.builder().bool(true).build(),
                "timestamp", AttributeValue.builder().n("1000").build()
            );
            QueryResponse response = QueryResponse.builder().items(List.of(item)).build();
            when(dynamoDbClient.query(any(QueryRequest.class))).thenReturn(response);

            List<AiChatMessageDto> result = aiChatService.getChatHistory(1);

            assertThat(result).hasSize(1);
            assertThat(result.get(0).getContent()).isEqualTo("テストメッセージ");
            assertThat(result.get(0).isUser()).isTrue();
            assertThat(result.get(0).getTimestamp()).isEqualTo(1000L);
        }

        @Test
        @DisplayName("空のチャット履歴を返す")
        void returnsEmptyHistory() {
            QueryResponse response = QueryResponse.builder().items(List.of()).build();
            when(dynamoDbClient.query(any(QueryRequest.class))).thenReturn(response);

            List<AiChatMessageDto> result = aiChatService.getChatHistory(1);

            assertThat(result).isEmpty();
        }

        @Test
        @DisplayName("複数メッセージを取得する")
        void returnsMultipleMessages() {
            Map<String, AttributeValue> item1 = Map.of(
                "content", AttributeValue.builder().s("メッセージ1").build(),
                "is_user", AttributeValue.builder().bool(true).build(),
                "timestamp", AttributeValue.builder().n("1000").build()
            );
            Map<String, AttributeValue> item2 = Map.of(
                "content", AttributeValue.builder().s("メッセージ2").build(),
                "is_user", AttributeValue.builder().bool(false).build(),
                "timestamp", AttributeValue.builder().n("2000").build()
            );
            QueryResponse response = QueryResponse.builder().items(List.of(item1, item2)).build();
            when(dynamoDbClient.query(any(QueryRequest.class))).thenReturn(response);

            List<AiChatMessageDto> result = aiChatService.getChatHistory(1);

            assertThat(result).hasSize(2);
            assertThat(result.get(0).getContent()).isEqualTo("メッセージ1");
            assertThat(result.get(1).getContent()).isEqualTo("メッセージ2");
            assertThat(result.get(1).isUser()).isFalse();
        }

        @Test
        @DisplayName("sender_idでクエリする")
        void queriesBySenderId() {
            QueryResponse response = QueryResponse.builder().items(List.of()).build();
            when(dynamoDbClient.query(any(QueryRequest.class))).thenReturn(response);

            aiChatService.getChatHistory(42);

            ArgumentCaptor<QueryRequest> captor = ArgumentCaptor.forClass(QueryRequest.class);
            verify(dynamoDbClient).query(captor.capture());
            QueryRequest captured = captor.getValue();
            assertThat(captured.tableName()).isEqualTo("test-ai-chat-table");
            assertThat(captured.keyConditionExpression()).isEqualTo("sender_id = :sender_id");
            assertThat(captured.expressionAttributeValues().get(":sender_id").n()).isEqualTo("42");
        }

        @Test
        @DisplayName("昇順でクエリする")
        void queriesInAscendingOrder() {
            QueryResponse response = QueryResponse.builder().items(List.of()).build();
            when(dynamoDbClient.query(any(QueryRequest.class))).thenReturn(response);

            aiChatService.getChatHistory(1);

            ArgumentCaptor<QueryRequest> captor = ArgumentCaptor.forClass(QueryRequest.class);
            verify(dynamoDbClient).query(captor.capture());
            assertThat(captor.getValue().scanIndexForward()).isTrue();
        }
    }
}
