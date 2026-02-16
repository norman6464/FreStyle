package com.example.FreStyle.usecase;

import com.example.FreStyle.service.CognitoAuthService;
import com.example.FreStyle.service.UserService;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import static org.assertj.core.api.Assertions.assertThatThrownBy;
import static org.mockito.Mockito.*;

@ExtendWith(MockitoExtension.class)
class CognitoConfirmUseCaseTest {

    @Mock private CognitoAuthService cognitoAuthService;
    @Mock private UserService userService;

    @InjectMocks
    private CognitoConfirmUseCase cognitoConfirmUseCase;

    @Test
    @DisplayName("確認成功時にCognito確認とユーザー有効化が行われる")
    void confirmsAndActivatesUser() {
        cognitoConfirmUseCase.execute("test@example.com", "123456");

        verify(cognitoAuthService).confirmUserSignup("test@example.com", "123456");
        verify(userService).activeUser("test@example.com");
    }

    @Test
    @DisplayName("Cognito確認失敗時にユーザー有効化が呼ばれない")
    void doesNotActivateWhenCognitoFails() {
        doThrow(new RuntimeException("Code mismatch"))
                .when(cognitoAuthService).confirmUserSignup("test@example.com", "wrong-code");

        assertThatThrownBy(() -> cognitoConfirmUseCase.execute("test@example.com", "wrong-code"))
                .isInstanceOf(RuntimeException.class);

        verify(userService, never()).activeUser(any());
    }
}
