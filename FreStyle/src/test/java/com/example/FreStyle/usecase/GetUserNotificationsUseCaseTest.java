package com.example.FreStyle.usecase;

import static org.assertj.core.api.Assertions.assertThat;
import static org.mockito.Mockito.when;

import java.time.LocalDateTime;
import java.util.List;

import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import com.example.FreStyle.dto.NotificationDto;
import com.example.FreStyle.entity.Notification;
import com.example.FreStyle.entity.User;
import com.example.FreStyle.repository.NotificationRepository;

@ExtendWith(MockitoExtension.class)
@DisplayName("GetUserNotificationsUseCase テスト")
class GetUserNotificationsUseCaseTest {

    @Mock
    private NotificationRepository notificationRepository;

    @InjectMocks
    private GetUserNotificationsUseCase getUserNotificationsUseCase;

    @Test
    @DisplayName("ユーザーの通知一覧を取得できる")
    void execute_returnsNotifications() {
        User user = new User();
        user.setId(1);
        Notification notification = new Notification();
        notification.setId(10);
        notification.setUser(user);
        notification.setType("NEW_MESSAGE");
        notification.setTitle("新しいメッセージ");
        notification.setMessage("田中さんからメッセージが届きました");
        notification.setIsRead(false);
        notification.setRelatedId(5);
        notification.setCreatedAt(LocalDateTime.of(2026, 1, 15, 10, 30));

        when(notificationRepository.findByUserIdOrderByCreatedAtDesc(1))
                .thenReturn(List.of(notification));

        List<NotificationDto> result = getUserNotificationsUseCase.execute(1);

        assertThat(result).hasSize(1);
        assertThat(result.get(0).id()).isEqualTo(10);
        assertThat(result.get(0).type()).isEqualTo("NEW_MESSAGE");
        assertThat(result.get(0).title()).isEqualTo("新しいメッセージ");
        assertThat(result.get(0).isRead()).isFalse();
    }

    @Test
    @DisplayName("通知がない場合は空リストを返す")
    void execute_returnsEmptyList() {
        when(notificationRepository.findByUserIdOrderByCreatedAtDesc(1))
                .thenReturn(List.of());

        List<NotificationDto> result = getUserNotificationsUseCase.execute(1);

        assertThat(result).isEmpty();
    }
}
