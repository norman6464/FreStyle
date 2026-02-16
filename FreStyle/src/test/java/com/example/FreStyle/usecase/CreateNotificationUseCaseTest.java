package com.example.FreStyle.usecase;

import static org.mockito.ArgumentMatchers.any;
import static org.mockito.Mockito.verify;

import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
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
}
