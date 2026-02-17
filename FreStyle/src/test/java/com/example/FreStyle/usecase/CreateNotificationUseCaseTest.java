package com.example.FreStyle.usecase;

import static org.assertj.core.api.Assertions.assertThat;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.Mockito.verify;

import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.ArgumentCaptor;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import com.example.FreStyle.entity.Notification;
import com.example.FreStyle.entity.User;
import com.example.FreStyle.repository.NotificationRepository;

@ExtendWith(MockitoExtension.class)
@DisplayName("CreateNotificationUseCase テスト")
class CreateNotificationUseCaseTest {

    @Mock
    private NotificationRepository notificationRepository;

    @InjectMocks
    private CreateNotificationUseCase createNotificationUseCase;

    private User testUser;

    @BeforeEach
    void setUp() {
        testUser = new User();
        testUser.setId(1);
        testUser.setName("テストユーザー");
    }

    @Test
    @DisplayName("通知を作成して保存できる")
    void execute_createsNotification() {
        createNotificationUseCase.execute(testUser, "NEW_MESSAGE", "新しいメッセージ", "田中さんからメッセージが届きました", 10);

        verify(notificationRepository).save(any(Notification.class));
    }

    @Test
    @DisplayName("relatedIdがnullでも通知を作成できる")
    void execute_createsNotificationWithoutRelatedId() {
        createNotificationUseCase.execute(testUser, "SYSTEM", "システム通知", "お知らせがあります", null);

        verify(notificationRepository).save(any(Notification.class));
    }

    @Test
    @DisplayName("保存される通知のフィールドが正しく設定される")
    void execute_savedNotificationHasCorrectFields() {
        createNotificationUseCase.execute(testUser, "NEW_MESSAGE", "新しいメッセージ", "田中さんから", 10);

        ArgumentCaptor<Notification> captor = ArgumentCaptor.forClass(Notification.class);
        verify(notificationRepository).save(captor.capture());

        Notification saved = captor.getValue();
        assertThat(saved.getUser()).isEqualTo(testUser);
        assertThat(saved.getType()).isEqualTo("NEW_MESSAGE");
        assertThat(saved.getTitle()).isEqualTo("新しいメッセージ");
        assertThat(saved.getMessage()).isEqualTo("田中さんから");
        assertThat(saved.getRelatedId()).isEqualTo(10);
    }

    @Test
    @DisplayName("relatedIdがnullの場合も正しくエンティティに設定される")
    void execute_savedNotificationHasNullRelatedId() {
        createNotificationUseCase.execute(testUser, "SYSTEM", "お知らせ", "メンテナンス予定", null);

        ArgumentCaptor<Notification> captor = ArgumentCaptor.forClass(Notification.class);
        verify(notificationRepository).save(captor.capture());

        Notification saved = captor.getValue();
        assertThat(saved.getUser()).isEqualTo(testUser);
        assertThat(saved.getType()).isEqualTo("SYSTEM");
        assertThat(saved.getTitle()).isEqualTo("お知らせ");
        assertThat(saved.getMessage()).isEqualTo("メンテナンス予定");
        assertThat(saved.getRelatedId()).isNull();
    }
}
