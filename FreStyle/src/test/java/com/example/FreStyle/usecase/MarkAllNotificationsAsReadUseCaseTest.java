package com.example.FreStyle.usecase;

import static org.mockito.ArgumentMatchers.anyList;
import static org.mockito.Mockito.never;
import static org.mockito.Mockito.verify;
import static org.mockito.Mockito.when;

import java.time.LocalDateTime;
import java.util.List;

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
@DisplayName("MarkAllNotificationsAsReadUseCase テスト")
class MarkAllNotificationsAsReadUseCaseTest {

    @Mock
    private NotificationRepository notificationRepository;

    @InjectMocks
    private MarkAllNotificationsAsReadUseCase markAllNotificationsAsReadUseCase;

    @Test
    @DisplayName("未読通知を全て既読にできる")
    void execute_marksAllAsRead() {
        User user = new User();
        user.setId(1);
        Notification n1 = new Notification(1, user, "NEW_MESSAGE", "通知1", "メッセージ1", false, null, LocalDateTime.now());
        Notification n2 = new Notification(2, user, "GOAL_ACHIEVED", "通知2", "メッセージ2", false, null, LocalDateTime.now());

        when(notificationRepository.findByUserIdAndIsReadFalseOrderByCreatedAtDesc(1))
                .thenReturn(List.of(n1, n2));

        markAllNotificationsAsReadUseCase.execute(1);

        verify(notificationRepository).saveAll(anyList());
    }

    @Test
    @DisplayName("未読通知がない場合は何もしない")
    void execute_doesNothingWhenNoUnread() {
        when(notificationRepository.findByUserIdAndIsReadFalseOrderByCreatedAtDesc(1))
                .thenReturn(List.of());

        markAllNotificationsAsReadUseCase.execute(1);

        verify(notificationRepository, never()).saveAll(anyList());
    }
}
