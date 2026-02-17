package com.example.FreStyle.usecase;

import static org.assertj.core.api.Assertions.assertThat;
import static org.mockito.Mockito.verify;
import static org.mockito.Mockito.when;

import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import com.example.FreStyle.repository.NotificationRepository;

@ExtendWith(MockitoExtension.class)
@DisplayName("GetUnreadNotificationCountUseCase テスト")
class GetUnreadNotificationCountUseCaseTest {

    @Mock
    private NotificationRepository notificationRepository;

    @InjectMocks
    private GetUnreadNotificationCountUseCase getUnreadNotificationCountUseCase;

    @Test
    @DisplayName("未読通知数を取得できる")
    void execute_returnsUnreadCount() {
        when(notificationRepository.countByUserIdAndIsReadFalse(1)).thenReturn(5L);

        long result = getUnreadNotificationCountUseCase.execute(1);

        assertThat(result).isEqualTo(5L);
    }

    @Test
    @DisplayName("未読通知がない場合は0を返す")
    void execute_returnsZeroWhenNoUnread() {
        when(notificationRepository.countByUserIdAndIsReadFalse(1)).thenReturn(0L);

        long result = getUnreadNotificationCountUseCase.execute(1);

        assertThat(result).isEqualTo(0L);
    }

    @Test
    @DisplayName("大量の未読通知数を正しく返す")
    void execute_returnsLargeCount() {
        when(notificationRepository.countByUserIdAndIsReadFalse(1)).thenReturn(9999L);

        long result = getUnreadNotificationCountUseCase.execute(1);

        assertThat(result).isEqualTo(9999L);
        verify(notificationRepository).countByUserIdAndIsReadFalse(1);
    }

    @Test
    @DisplayName("異なるユーザーIDで正しいパラメータが渡される")
    void execute_passesCorrectUserId() {
        when(notificationRepository.countByUserIdAndIsReadFalse(42)).thenReturn(3L);

        long result = getUnreadNotificationCountUseCase.execute(42);

        assertThat(result).isEqualTo(3L);
        verify(notificationRepository).countByUserIdAndIsReadFalse(42);
    }
}
