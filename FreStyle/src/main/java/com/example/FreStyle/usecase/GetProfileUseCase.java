package com.example.FreStyle.usecase;

import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import com.example.FreStyle.dto.ProfileDto;
import com.example.FreStyle.entity.User;
import com.example.FreStyle.exception.ResourceNotFoundException;
import com.example.FreStyle.service.UserIdentityService;

import lombok.RequiredArgsConstructor;

@Service
@RequiredArgsConstructor
public class GetProfileUseCase {

    private final UserIdentityService userIdentityService;

    @Transactional(readOnly = true)
    public ProfileDto execute(String sub) {
        User user = userIdentityService.findUserBySub(sub);
        if (user == null) {
            throw new ResourceNotFoundException("ユーザーが見つかりません: sub=" + sub);
        }
        return new ProfileDto(user.getName(), user.getBio(), user.getIconUrl(), user.getStatus());
    }
}
