package com.example.FreStyle.usecase;

import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import com.example.FreStyle.dto.UserProfileDto;
import com.example.FreStyle.entity.User;
import com.example.FreStyle.form.UserProfileForm;
import com.example.FreStyle.service.UserIdentityService;
import com.example.FreStyle.service.UserProfileService;

import lombok.RequiredArgsConstructor;

@Service
@RequiredArgsConstructor
public class UpdateUserProfileUseCase {

    private final UserProfileService userProfileService;
    private final UserIdentityService userIdentityService;

    @Transactional
    public UserProfileDto execute(String sub, UserProfileForm form) {
        User user = userIdentityService.findUserBySub(sub);
        return userProfileService.updateProfile(user.getId(), form);
    }
}
