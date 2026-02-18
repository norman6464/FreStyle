package com.example.FreStyle.usecase;

import static org.junit.jupiter.api.Assertions.*;
import static org.mockito.Mockito.*;

import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import com.example.FreStyle.dto.ProfileDto;
import com.example.FreStyle.entity.User;
import com.example.FreStyle.exception.ResourceNotFoundException;
import com.example.FreStyle.service.UserIdentityService;

@ExtendWith(MockitoExtension.class)
@DisplayName("GetProfileUseCase テスト")
class GetProfileUseCaseTest {

    @Mock
    private UserIdentityService userIdentityService;

    @InjectMocks
    private GetProfileUseCase getProfileUseCase;

    private User testUser;

    @BeforeEach
    void setUp() {
        testUser = new User();
        testUser.setId(1);
        testUser.setName("テストユーザー");
        testUser.setEmail("test@example.com");
        testUser.setBio("テスト自己紹介");
        testUser.setIconUrl("https://cdn.example.com/profiles/1/avatar.png");
        testUser.setStatus("学習中");
    }

    @Test
    @DisplayName("subからProfileDtoを取得できる")
    void execute_ReturnsProfileDto() {
        when(userIdentityService.findUserBySub("sub-123")).thenReturn(testUser);

        ProfileDto result = getProfileUseCase.execute("sub-123");

        assertEquals("テストユーザー", result.name());
        assertEquals("テスト自己紹介", result.bio());
        assertEquals("https://cdn.example.com/profiles/1/avatar.png", result.iconUrl());
        assertEquals("学習中", result.status());
    }

    @Test
    @DisplayName("ユーザーが見つからない場合はResourceNotFoundExceptionをスロー")
    void execute_ThrowsException_WhenUserNotFound() {
        when(userIdentityService.findUserBySub("sub-999")).thenReturn(null);

        assertThrows(ResourceNotFoundException.class, () -> {
            getProfileUseCase.execute("sub-999");
        });
    }
}
