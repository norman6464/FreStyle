package com.example.FreStyle.service;

import com.example.FreStyle.entity.User;
import com.example.FreStyle.entity.UserIdentity;
import com.example.FreStyle.repository.UserIdentityRepository;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.ArgumentCaptor;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import java.util.Optional;

import static org.junit.jupiter.api.Assertions.*;
import static org.mockito.Mockito.*;

@ExtendWith(MockitoExtension.class)
class UserIdentityServiceTest {

    @Mock
    private UserIdentityRepository userIdentityRepository;

    @InjectMocks
    private UserIdentityService userIdentityService;

    @Test
    @DisplayName("registerUserIdentity: 新規IDを登録する")
    void registerUserIdentity_savesNewIdentity() {
        User user = new User();
        user.setId(1);
        when(userIdentityRepository.findByProviderAndProviderSub("google", "sub123"))
                .thenReturn(Optional.empty());

        userIdentityService.registerUserIdentity(user, "google", "sub123");

        ArgumentCaptor<UserIdentity> captor = ArgumentCaptor.forClass(UserIdentity.class);
        verify(userIdentityRepository).save(captor.capture());
        assertEquals(user, captor.getValue().getUser());
        assertEquals("google", captor.getValue().getProvider());
        assertEquals("sub123", captor.getValue().getProviderSub());
    }

    @Test
    @DisplayName("registerUserIdentity: 既存IDの場合は保存しない")
    void registerUserIdentity_skipsIfExists() {
        UserIdentity existing = new UserIdentity();
        when(userIdentityRepository.findByProviderAndProviderSub("google", "sub123"))
                .thenReturn(Optional.of(existing));

        userIdentityService.registerUserIdentity(new User(), "google", "sub123");

        verify(userIdentityRepository, never()).save(any());
    }

    @Test
    @DisplayName("findUserBySub: ユーザーを返す")
    void findUserBySub_returnsUser() {
        User user = new User();
        user.setId(1);
        user.setName("テスト");
        UserIdentity identity = new UserIdentity();
        identity.setUser(user);
        when(userIdentityRepository.findByProviderSub("sub123"))
                .thenReturn(Optional.of(identity));

        User result = userIdentityService.findUserBySub("sub123");

        assertEquals("テスト", result.getName());
    }

    @Test
    @DisplayName("findUserBySub: 存在しないsubで例外")
    void findUserBySub_throwsWhenNotFound() {
        when(userIdentityRepository.findByProviderSub("invalid"))
                .thenReturn(Optional.empty());

        RuntimeException ex = assertThrows(RuntimeException.class,
                () -> userIdentityService.findUserBySub("invalid"));
        assertEquals("ユーザーが存在しません。", ex.getMessage());
    }
}
