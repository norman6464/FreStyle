package com.example.FreStyle.usecase;

import com.example.FreStyle.service.CognitoAuthService;
import com.example.FreStyle.service.UserService;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;

@Service
@RequiredArgsConstructor
@Slf4j
public class CognitoConfirmUseCase {

    private final CognitoAuthService cognitoAuthService;
    private final UserService userService;

    public void execute(String email, String code) {
        log.info("CognitoConfirmUseCase: サインアップ確認処理開始 - email: {}", email);

        cognitoAuthService.confirmUserSignup(email, code);
        userService.activeUser(email);

        log.info("CognitoConfirmUseCase: サインアップ確認処理完了 - email: {}", email);
    }
}
