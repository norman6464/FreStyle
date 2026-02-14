package com.example.FreStyle.service;

import com.example.FreStyle.dto.NoteDto;
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
class NoteServiceTest {

    @Mock
    private DynamoDbClient dynamoDbClient;

    private NoteService noteService;

    @BeforeEach
    void setUp() {
        noteService = new NoteService(dynamoDbClient, "fre_style_notes");
    }

    @Nested
    @DisplayName("findByUserId - ユーザーのノート一覧取得")
    class FindByUserIdTest {

        @Test
        @DisplayName("ユーザーのノートを更新日時降順で取得する")
        void shouldReturnNotesSortedByUpdatedAtDesc() {
            Map<String, AttributeValue> item1 = createNoteItem(1, "note-1", "タイトル1", "内容1", false, 1000L, 2000L);
            Map<String, AttributeValue> item2 = createNoteItem(1, "note-2", "タイトル2", "内容2", true, 1500L, 3000L);

            QueryResponse response = QueryResponse.builder()
                    .items(List.of(item1, item2))
                    .build();
            when(dynamoDbClient.query(any(QueryRequest.class))).thenReturn(response);

            List<NoteDto> result = noteService.findByUserId(1);

            assertThat(result).hasSize(2);
            assertThat(result.get(0).getNoteId()).isEqualTo("note-1");
            assertThat(result.get(1).getNoteId()).isEqualTo("note-2");
        }

        @Test
        @DisplayName("ノートが存在しない場合は空リストを返す")
        void shouldReturnEmptyListWhenNoNotes() {
            QueryResponse response = QueryResponse.builder()
                    .items(List.of())
                    .build();
            when(dynamoDbClient.query(any(QueryRequest.class))).thenReturn(response);

            List<NoteDto> result = noteService.findByUserId(1);

            assertThat(result).isEmpty();
        }
    }

    @Nested
    @DisplayName("save - ノート作成")
    class SaveTest {

        @Test
        @DisplayName("新しいノートを作成して返す")
        void shouldCreateNoteAndReturnDto() {
            when(dynamoDbClient.putItem(any(PutItemRequest.class)))
                    .thenReturn(PutItemResponse.builder().build());

            NoteDto result = noteService.save(1, "新しいノート");

            assertThat(result).isNotNull();
            assertThat(result.getUserId()).isEqualTo(1);
            assertThat(result.getTitle()).isEqualTo("新しいノート");
            assertThat(result.getContent()).isEmpty();
            assertThat(result.getIsPinned()).isFalse();
            assertThat(result.getNoteId()).isNotBlank();

            verify(dynamoDbClient, times(1)).putItem(any(PutItemRequest.class));
        }
    }

    @Nested
    @DisplayName("update - ノート更新")
    class UpdateTest {

        @Test
        @DisplayName("タイトルと内容を更新する")
        void shouldUpdateTitleAndContent() {
            when(dynamoDbClient.updateItem(any(UpdateItemRequest.class)))
                    .thenReturn(UpdateItemResponse.builder().build());

            noteService.update(1, "note-1", "更新タイトル", "更新内容", false);

            ArgumentCaptor<UpdateItemRequest> captor = ArgumentCaptor.forClass(UpdateItemRequest.class);
            verify(dynamoDbClient).updateItem(captor.capture());

            UpdateItemRequest request = captor.getValue();
            assertThat(request.tableName()).isEqualTo("fre_style_notes");
            assertThat(request.key().get("user_id").n()).isEqualTo("1");
            assertThat(request.key().get("note_id").s()).isEqualTo("note-1");
        }
    }

    @Nested
    @DisplayName("delete - ノート削除")
    class DeleteTest {

        @Test
        @DisplayName("指定されたノートを削除する")
        void shouldDeleteNote() {
            when(dynamoDbClient.deleteItem(any(DeleteItemRequest.class)))
                    .thenReturn(DeleteItemResponse.builder().build());

            noteService.delete(1, "note-1");

            ArgumentCaptor<DeleteItemRequest> captor = ArgumentCaptor.forClass(DeleteItemRequest.class);
            verify(dynamoDbClient).deleteItem(captor.capture());

            DeleteItemRequest request = captor.getValue();
            assertThat(request.tableName()).isEqualTo("fre_style_notes");
            assertThat(request.key().get("user_id").n()).isEqualTo("1");
            assertThat(request.key().get("note_id").s()).isEqualTo("note-1");
        }
    }

    private Map<String, AttributeValue> createNoteItem(
            Integer userId, String noteId, String title, String content,
            boolean isPinned, Long createdAt, Long updatedAt) {
        Map<String, AttributeValue> item = new HashMap<>();
        item.put("user_id", AttributeValue.builder().n(userId.toString()).build());
        item.put("note_id", AttributeValue.builder().s(noteId).build());
        item.put("title", AttributeValue.builder().s(title).build());
        item.put("content", AttributeValue.builder().s(content).build());
        item.put("is_pinned", AttributeValue.builder().bool(isPinned).build());
        item.put("created_at", AttributeValue.builder().n(createdAt.toString()).build());
        item.put("updated_at", AttributeValue.builder().n(updatedAt.toString()).build());
        return item;
    }
}
