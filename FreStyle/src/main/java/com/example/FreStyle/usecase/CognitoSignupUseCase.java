package com.example.FreStyle.usecase;

import com.example.FreStyle.form.SignupForm;
import com.example.FreStyle.service.CognitoAuthService;
import com.example.FreStyle.service.UserService;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;

@Service
@RequiredArgsConstructor
@Slf4j
public class CognitoSignupUseCase {

    private final CognitoAuthService cognitoAuthService;
    private final UserService userService;

    public void execute(SignupForm form) {
        log.info("CognitoSignupUseCase: サインアップ処理開始 - email: {}", form.getEmail());

        cognitoAuthService.signUpUser(form.getEmail(), form.getPassword(), form.getName());
        userService.registerUser(form);

        log.info("CognitoSignupUseCase: サインアップ処理完了 - email: {}", form.getEmail());
    }
}
