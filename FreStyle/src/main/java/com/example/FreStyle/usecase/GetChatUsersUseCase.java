package com.example.FreStyle.usecase;

import java.util.List;

import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import com.example.FreStyle.dto.UserDto;
import com.example.FreStyle.entity.User;
import com.example.FreStyle.service.UserIdentityService;
import com.example.FreStyle.service.UserService;

import lombok.RequiredArgsConstructor;

@Service
@RequiredArgsConstructor
public class GetChatUsersUseCase {

    private final UserIdentityService userIdentityService;
    private final UserService userService;

    @Transactional(readOnly = true)
    public List<UserDto> execute(String sub, String query) {
        User user = userIdentityService.findUserBySub(sub);
        return userService.findUsersWithRoomId(user.getId(), query);
    }
}
