package com.example.FreStyle.usecase;

import com.example.FreStyle.form.SignupForm;
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

import org.mockito.InOrder;

@ExtendWith(MockitoExtension.class)
class CognitoSignupUseCaseTest {

    @Mock private CognitoAuthService cognitoAuthService;
    @Mock private UserService userService;

    @InjectMocks
    private CognitoSignupUseCase cognitoSignupUseCase;

    @Test
    @DisplayName("サインアップ成功時にCognitoとDBの両方にユーザーが登録される")
    void registersUserInCognitoAndDb() {
        SignupForm form = new SignupForm("test@example.com", "password123", "テストユーザー");

        cognitoSignupUseCase.execute(form);

        verify(cognitoAuthService).signUpUser("test@example.com", "password123", "テストユーザー");
        verify(userService).registerUser(form);
    }

    @Test
    @DisplayName("Cognito登録がDB登録より先に実行される")
    void registersInCognitoBeforeDb() {
        SignupForm form = new SignupForm("test@example.com", "password123", "テストユーザー");

        cognitoSignupUseCase.execute(form);

        InOrder inOrder = inOrder(cognitoAuthService, userService);
        inOrder.verify(cognitoAuthService).signUpUser("test@example.com", "password123", "テストユーザー");
        inOrder.verify(userService).registerUser(form);
    }

    @Test
    @DisplayName("DB登録失敗時に例外が伝搬する")
    void propagatesExceptionWhenDbFails() {
        SignupForm form = new SignupForm("test@example.com", "password123", "テストユーザー");

        doThrow(new RuntimeException("DB error"))
                .when(userService).registerUser(form);

        assertThatThrownBy(() -> cognitoSignupUseCase.execute(form))
                .isInstanceOf(RuntimeException.class)
                .hasMessage("DB error");

        verify(cognitoAuthService).signUpUser("test@example.com", "password123", "テストユーザー");
    }

    @Test
    @DisplayName("Cognito登録失敗時にDB登録が呼ばれない")
    void doesNotRegisterInDbWhenCognitoFails() {
        SignupForm form = new SignupForm("test@example.com", "password123", "テストユーザー");

        doThrow(new RuntimeException("Cognito error"))
                .when(cognitoAuthService).signUpUser("test@example.com", "password123", "テストユーザー");

        assertThatThrownBy(() -> cognitoSignupUseCase.execute(form))
                .isInstanceOf(RuntimeException.class);

        verify(userService, never()).registerUser(any());
    }
}
