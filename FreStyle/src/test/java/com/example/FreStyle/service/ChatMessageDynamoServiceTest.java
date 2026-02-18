package com.example.FreStyle.service;

import com.example.FreStyle.dto.ChatMessageDto;
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
class ChatMessageDynamoServiceTest {

    @Mock
    private DynamoDbClient dynamoDbClient;

    private ChatMessageDynamoService service;

    @BeforeEach
    void setUp() {
        service = new ChatMessageDynamoService(dynamoDbClient, "fre_style_chat");
    }

    @Nested
    @DisplayName("save - メッセージ保存")
    class SaveTest {

        @Test
        @DisplayName("メッセージを保存してDTOを返す")
        void shouldSaveMessageAndReturnDto() {
            when(dynamoDbClient.putItem(any(PutItemRequest.class)))
                    .thenReturn(PutItemResponse.builder().build());

            ChatMessageDto result = service.save(10, 1, "こんにちは");

            assertThat(result).isNotNull();
            assertThat(result.id()).isNotBlank();
            assertThat(result.roomId()).isEqualTo(10);
            assertThat(result.senderId()).isEqualTo(1);
            assertThat(result.senderName()).isNull();
            assertThat(result.content()).isEqualTo("こんにちは");
            assertThat(result.createdAt()).isGreaterThan(0L);

            ArgumentCaptor<PutItemRequest> captor = ArgumentCaptor.forClass(PutItemRequest.class);
            verify(dynamoDbClient).putItem(captor.capture());
            PutItemRequest request = captor.getValue();
            assertThat(request.tableName()).isEqualTo("fre_style_chat");
            assertThat(request.item().get("room_id").n()).isEqualTo("10");
            assertThat(request.item().get("sender_id").n()).isEqualTo("1");
            assertThat(request.item().get("content").s()).isEqualTo("こんにちは");
        }
    }

    @Nested
    @DisplayName("findByRoomId - ルーム別メッセージ取得")
    class FindByRoomIdTest {

        @Test
        @DisplayName("ルームのメッセージを昇順で取得する")
        void shouldReturnMessagesSortedByCreatedAtAsc() {
            Map<String, AttributeValue> item1 = createMessageItem(10, "msg-1", 1, "質問", 1000L);
            Map<String, AttributeValue> item2 = createMessageItem(10, "msg-2", 2, "回答", 2000L);

            QueryResponse response = QueryResponse.builder()
                    .items(List.of(item1, item2))
                    .build();
            when(dynamoDbClient.query(any(QueryRequest.class))).thenReturn(response);

            List<ChatMessageDto> result = service.findByRoomId(10);

            assertThat(result).hasSize(2);
            assertThat(result.get(0).id()).isEqualTo("msg-1");
            assertThat(result.get(0).senderId()).isEqualTo(1);
            assertThat(result.get(0).content()).isEqualTo("質問");
            assertThat(result.get(1).id()).isEqualTo("msg-2");
            assertThat(result.get(1).senderId()).isEqualTo(2);

            ArgumentCaptor<QueryRequest> captor = ArgumentCaptor.forClass(QueryRequest.class);
            verify(dynamoDbClient).query(captor.capture());
            QueryRequest request = captor.getValue();
            assertThat(request.tableName()).isEqualTo("fre_style_chat");
            assertThat(request.scanIndexForward()).isTrue();
        }

        @Test
        @DisplayName("メッセージが存在しない場合は空リストを返す")
        void shouldReturnEmptyListWhenNoMessages() {
            QueryResponse response = QueryResponse.builder()
                    .items(List.of())
                    .build();
            when(dynamoDbClient.query(any(QueryRequest.class))).thenReturn(response);

            List<ChatMessageDto> result = service.findByRoomId(999);

            assertThat(result).isEmpty();
        }
    }

    @Nested
    @DisplayName("deleteByRoomIdAndCreatedAt - メッセージ削除")
    class DeleteTest {

        @Test
        @DisplayName("指定のメッセージを削除する")
        void shouldDeleteMessage() {
            when(dynamoDbClient.deleteItem(any(DeleteItemRequest.class)))
                    .thenReturn(DeleteItemResponse.builder().build());

            service.deleteByRoomIdAndCreatedAt(10, 1000L);

            ArgumentCaptor<DeleteItemRequest> captor = ArgumentCaptor.forClass(DeleteItemRequest.class);
            verify(dynamoDbClient).deleteItem(captor.capture());
            DeleteItemRequest request = captor.getValue();
            assertThat(request.tableName()).isEqualTo("fre_style_chat");
            assertThat(request.key().get("room_id").n()).isEqualTo("10");
            assertThat(request.key().get("created_at").n()).isEqualTo("1000");
        }
    }

    @Nested
    @DisplayName("findLatestByRoomId - ルームの最新メッセージ取得")
    class FindLatestByRoomIdTest {

        @Test
        @DisplayName("最新のメッセージを1件取得する")
        void shouldReturnLatestMessage() {
            Map<String, AttributeValue> item = createMessageItem(10, "msg-latest", 1, "最新", 5000L);
            QueryResponse response = QueryResponse.builder()
                    .items(List.of(item))
                    .build();
            when(dynamoDbClient.query(any(QueryRequest.class))).thenReturn(response);

            ChatMessageDto result = service.findLatestByRoomId(10);

            assertThat(result).isNotNull();
            assertThat(result.id()).isEqualTo("msg-latest");
            assertThat(result.content()).isEqualTo("最新");

            ArgumentCaptor<QueryRequest> captor = ArgumentCaptor.forClass(QueryRequest.class);
            verify(dynamoDbClient).query(captor.capture());
            QueryRequest request = captor.getValue();
            assertThat(request.scanIndexForward()).isFalse();
            assertThat(request.limit()).isEqualTo(1);
        }

        @Test
        @DisplayName("メッセージが存在しない場合はnullを返す")
        void shouldReturnNullWhenNoMessages() {
            QueryResponse response = QueryResponse.builder()
                    .items(List.of())
                    .build();
            when(dynamoDbClient.query(any(QueryRequest.class))).thenReturn(response);

            ChatMessageDto result = service.findLatestByRoomId(999);

            assertThat(result).isNull();
        }
    }

    @Nested
    @DisplayName("findLatestByRoomIds - 複数ルームの最新メッセージ一括取得")
    class FindLatestByRoomIdsTest {

        @Test
        @DisplayName("複数ルームの最新メッセージをMapで返す")
        void shouldReturnLatestMessagesForMultipleRooms() {
            Map<String, AttributeValue> item1 = createMessageItem(10, "msg-1", 1, "ルーム10最新", 5000L);
            Map<String, AttributeValue> item2 = createMessageItem(20, "msg-2", 2, "ルーム20最新", 6000L);

            // ルーム10のクエリ
            QueryResponse response1 = QueryResponse.builder().items(List.of(item1)).build();
            // ルーム20のクエリ
            QueryResponse response2 = QueryResponse.builder().items(List.of(item2)).build();

            when(dynamoDbClient.query(any(QueryRequest.class)))
                    .thenReturn(response1)
                    .thenReturn(response2);

            Map<Integer, ChatMessageDto> result = service.findLatestByRoomIds(List.of(10, 20));

            assertThat(result).hasSize(2);
            assertThat(result.get(10).content()).isEqualTo("ルーム10最新");
            assertThat(result.get(20).content()).isEqualTo("ルーム20最新");
        }

        @Test
        @DisplayName("メッセージがないルームはMapに含まれない")
        void shouldExcludeRoomsWithNoMessages() {
            Map<String, AttributeValue> item1 = createMessageItem(10, "msg-1", 1, "あり", 5000L);
            QueryResponse response1 = QueryResponse.builder().items(List.of(item1)).build();
            QueryResponse emptyResponse = QueryResponse.builder().items(List.of()).build();

            when(dynamoDbClient.query(any(QueryRequest.class)))
                    .thenReturn(response1)
                    .thenReturn(emptyResponse);

            Map<Integer, ChatMessageDto> result = service.findLatestByRoomIds(List.of(10, 20));

            assertThat(result).hasSize(1);
            assertThat(result).containsKey(10);
            assertThat(result).doesNotContainKey(20);
        }
    }

    private Map<String, AttributeValue> createMessageItem(
            Integer roomId, String messageId, Integer senderId,
            String content, Long createdAt) {
        Map<String, AttributeValue> item = new HashMap<>();
        item.put("room_id", AttributeValue.builder().n(roomId.toString()).build());
        item.put("message_id", AttributeValue.builder().s(messageId).build());
        item.put("sender_id", AttributeValue.builder().n(senderId.toString()).build());
        item.put("content", AttributeValue.builder().s(content).build());
        item.put("created_at", AttributeValue.builder().n(createdAt.toString()).build());
        return item;
    }
}
