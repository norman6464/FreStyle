package com.example.FreStyle.usecase;

import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import com.example.FreStyle.entity.User;
import com.example.FreStyle.service.UserIdentityService;
import com.example.FreStyle.service.UserProfileService;

import lombok.RequiredArgsConstructor;

@Service
@RequiredArgsConstructor
public class DeleteUserProfileUseCase {

    private final UserProfileService userProfileService;
    private final UserIdentityService userIdentityService;

    @Transactional
    public void execute(String sub) {
        User user = userIdentityService.findUserBySub(sub);
        userProfileService.deleteProfile(user.getId());
    }
}
