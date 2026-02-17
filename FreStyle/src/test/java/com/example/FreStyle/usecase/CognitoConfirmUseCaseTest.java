package com.example.FreStyle.usecase;

import com.example.FreStyle.service.CognitoAuthService;
import com.example.FreStyle.service.UserService;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InOrder;
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
    @DisplayName("Cognito確認→ユーザー有効化の順序で実行される")
    void executesInCorrectOrder() {
        cognitoConfirmUseCase.execute("order@example.com", "111111");

        InOrder inOrder = inOrder(cognitoAuthService, userService);
        inOrder.verify(cognitoAuthService).confirmUserSignup("order@example.com", "111111");
        inOrder.verify(userService).activeUser("order@example.com");
    }

    @Test
    @DisplayName("ユーザー有効化失敗時に例外が伝搬しCognito確認は完了済み")
    void propagatesUserServiceExceptionAfterCognitoConfirm() {
        doThrow(new RuntimeException("DB error"))
                .when(userService).activeUser("fail@example.com");

        assertThatThrownBy(() -> cognitoConfirmUseCase.execute("fail@example.com", "123456"))
                .isInstanceOf(RuntimeException.class)
                .hasMessage("DB error");

        verify(cognitoAuthService).confirmUserSignup("fail@example.com", "123456");
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
