package com.example.FreStyle.usecase;

import org.springframework.stereotype.Service;

import com.example.FreStyle.dto.PresignedUrlResponse;
import com.example.FreStyle.entity.User;
import com.example.FreStyle.service.ProfileImageService;
import com.example.FreStyle.service.UserIdentityService;

import lombok.RequiredArgsConstructor;

@Service
@RequiredArgsConstructor
public class GenerateProfileImageUrlUseCase {

    private final ProfileImageService profileImageService;
    private final UserIdentityService userIdentityService;

    public PresignedUrlResponse execute(String sub, String fileName, String contentType) {
        User user = userIdentityService.findUserBySub(sub);
        return profileImageService.generatePresignedUrl(user.getId(), fileName, contentType);
    }
}
