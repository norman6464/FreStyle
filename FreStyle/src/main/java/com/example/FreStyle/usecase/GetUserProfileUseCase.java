package com.example.FreStyle.usecase;

import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import com.example.FreStyle.dto.UserProfileDto;
import com.example.FreStyle.entity.User;
import com.example.FreStyle.service.UserIdentityService;
import com.example.FreStyle.service.UserProfileService;

import lombok.RequiredArgsConstructor;

@Service
@RequiredArgsConstructor
public class GetUserProfileUseCase {

    private final UserProfileService userProfileService;
    private final UserIdentityService userIdentityService;

    @Transactional(readOnly = true)
    public UserProfileDto execute(String sub) {
        User user = userIdentityService.findUserBySub(sub);
        return userProfileService.getProfileByUserId(user.getId());
    }
}
