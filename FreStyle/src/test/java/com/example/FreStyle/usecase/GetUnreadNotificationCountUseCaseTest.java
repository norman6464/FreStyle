package com.example.FreStyle.usecase;

import static org.assertj.core.api.Assertions.assertThat;
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
}
