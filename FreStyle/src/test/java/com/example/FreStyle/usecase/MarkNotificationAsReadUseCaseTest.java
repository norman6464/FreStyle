package com.example.FreStyle.usecase;

import static org.junit.jupiter.api.Assertions.assertThrows;
import static org.mockito.Mockito.verify;
import static org.mockito.Mockito.when;

import java.time.LocalDateTime;
import java.util.Optional;

import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import com.example.FreStyle.entity.Notification;
import com.example.FreStyle.entity.User;
import com.example.FreStyle.exception.ResourceNotFoundException;
import com.example.FreStyle.repository.NotificationRepository;

@ExtendWith(MockitoExtension.class)
@DisplayName("MarkNotificationAsReadUseCase テスト")
class MarkNotificationAsReadUseCaseTest {

    @Mock
    private NotificationRepository notificationRepository;

    @InjectMocks
    private MarkNotificationAsReadUseCase markNotificationAsReadUseCase;

    @Test
    @DisplayName("通知を既読にできる")
    void execute_marksAsRead() {
        User user = new User();
        user.setId(1);
        Notification notification = new Notification();
        notification.setId(10);
        notification.setUser(user);
        notification.setType("NEW_MESSAGE");
        notification.setTitle("テスト");
        notification.setMessage("テストメッセージ");
        notification.setIsRead(false);
        notification.setCreatedAt(LocalDateTime.now());

        when(notificationRepository.findById(10)).thenReturn(Optional.of(notification));

        markNotificationAsReadUseCase.execute(1, 10);

        verify(notificationRepository).save(notification);
    }

    @Test
    @DisplayName("存在しない通知IDの場合は例外をスローする")
    void execute_throwsWhenNotFound() {
        when(notificationRepository.findById(999)).thenReturn(Optional.empty());

        assertThrows(ResourceNotFoundException.class,
                () -> markNotificationAsReadUseCase.execute(1, 999));
    }

    @Test
    @DisplayName("他のユーザーの通知を既読にしようとすると例外をスローする")
    void execute_throwsWhenUnauthorized() {
        User user = new User();
        user.setId(2);
        Notification notification = new Notification();
        notification.setId(10);
        notification.setUser(user);
        notification.setType("NEW_MESSAGE");
        notification.setTitle("テスト");
        notification.setMessage("テストメッセージ");
        notification.setIsRead(false);
        notification.setCreatedAt(LocalDateTime.now());

        when(notificationRepository.findById(10)).thenReturn(Optional.of(notification));

        assertThrows(ResourceNotFoundException.class,
                () -> markNotificationAsReadUseCase.execute(1, 10));
    }
}
