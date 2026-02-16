package com.example.FreStyle.usecase;

import org.springframework.security.oauth2.jwt.Jwt;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import com.example.FreStyle.form.ProfileForm;
import com.example.FreStyle.service.CognitoAuthService;
import com.example.FreStyle.service.UserService;

import lombok.RequiredArgsConstructor;

@Service
@RequiredArgsConstructor
public class UpdateProfileUseCase {

    private final CognitoAuthService cognitoAuthService;
    private final UserService userService;

    @Transactional
    public void execute(Jwt jwt, ProfileForm form) {
        String sub = jwt.getSubject();
        boolean isOidcUser = jwt.hasClaim("cognito:groups");

        if (!isOidcUser) {
            String accessToken = jwt.getTokenValue();
            cognitoAuthService.updateUserProfile(accessToken, form.getName());
        }

        userService.updateUser(form, sub);
    }
}
