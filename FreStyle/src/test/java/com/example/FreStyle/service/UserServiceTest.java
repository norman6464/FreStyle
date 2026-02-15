package com.example.FreStyle.service;

import com.example.FreStyle.dto.LoginUserDto;
import com.example.FreStyle.entity.User;
import com.example.FreStyle.form.ProfileForm;
import com.example.FreStyle.form.SignupForm;
import com.example.FreStyle.repository.ChatRoomRepository;
import com.example.FreStyle.repository.UserRepository;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import java.util.Optional;

import static org.junit.jupiter.api.Assertions.*;
import static org.mockito.Mockito.*;

@ExtendWith(MockitoExtension.class)
class UserServiceTest {

    @Mock
    private UserRepository userRepository;

    @Mock
    private UserIdentityService userIdentityService;

    @Mock
    private ChatRoomRepository chatRoomRepository;

    @InjectMocks
    private UserService userService;

    private User createUser(Integer id, String name, String email, boolean active) {
        User user = new User();
        user.setId(id);
        user.setName(name);
        user.setEmail(email);
        user.setIsActive(active);
        return user;
    }

    @Test
    @DisplayName("registerUser: ユーザーを登録する")
    void registerUser_savesUser() {
        SignupForm form = new SignupForm("test@example.com", "password", "テスト");
        when(userRepository.existsByEmail("test@example.com")).thenReturn(false);

        userService.registerUser(form);

        verify(userRepository).save(any(User.class));
    }

    @Test
    @DisplayName("registerUser: メールアドレスが重複している場合例外")
    void registerUser_throwsWhenDuplicate() {
        SignupForm form = new SignupForm("dup@example.com", "password", "テスト");
        when(userRepository.existsByEmail("dup@example.com")).thenReturn(true);

        RuntimeException ex = assertThrows(RuntimeException.class,
                () -> userService.registerUser(form));
        assertEquals("このメールアドレスは既に使用されています。", ex.getMessage());
    }

    @Test
    @DisplayName("findUserById: ユーザーを返す")
    void findUserById_returnsUser() {
        User user = createUser(1, "テスト", "test@example.com", true);
        when(userRepository.findById(1)).thenReturn(Optional.of(user));

        User result = userService.findUserById(1);

        assertEquals("テスト", result.getName());
    }

    @Test
    @DisplayName("findUserById: 存在しない場合例外")
    void findUserById_throwsWhenNotFound() {
        when(userRepository.findById(999)).thenReturn(Optional.empty());

        RuntimeException ex = assertThrows(RuntimeException.class,
                () -> userService.findUserById(999));
        assertEquals("ユーザーが見つかりません。", ex.getMessage());
    }

    @Test
    @DisplayName("findUserByEmail: アクティブなユーザーを返す")
    void findUserByEmail_returnsActiveUser() {
        User user = createUser(1, "テスト", "test@example.com", true);
        when(userRepository.findByEmail("test@example.com")).thenReturn(Optional.of(user));

        User result = userService.findUserByEmail("test@example.com");

        assertEquals("テスト", result.getName());
    }

    @Test
    @DisplayName("findUserByEmail: 非アクティブの場合例外")
    void findUserByEmail_throwsWhenInactive() {
        User user = createUser(1, "テスト", "test@example.com", false);
        when(userRepository.findByEmail("test@example.com")).thenReturn(Optional.of(user));

        RuntimeException ex = assertThrows(RuntimeException.class,
                () -> userService.findUserByEmail("test@example.com"));
        assertEquals("メール認証は完了していないため、ログインできません。", ex.getMessage());
    }

    @Test
    @DisplayName("findUserByEmail: 存在しない場合例外")
    void findUserByEmail_throwsWhenNotFound() {
        when(userRepository.findByEmail("none@example.com")).thenReturn(Optional.empty());

        RuntimeException ex = assertThrows(RuntimeException.class,
                () -> userService.findUserByEmail("none@example.com"));
        assertEquals("ユーザーが存在しません。", ex.getMessage());
    }

    @Test
    @DisplayName("registerUserOIDC: 新規ユーザーを作成してIdentityを登録する")
    void registerUserOIDC_createsNewUser() {
        when(userRepository.findByEmail("oidc@example.com")).thenReturn(Optional.empty());
        User saved = createUser(10, "OIDC User", "oidc@example.com", true);
        when(userRepository.save(any())).thenReturn(saved);

        User result = userService.registerUserOIDC("OIDC User", "oidc@example.com", "google", "sub123");

        assertEquals(10, result.getId());
        verify(userIdentityService).registerUserIdentity(saved, "google", "sub123");
    }

    @Test
    @DisplayName("registerUserOIDC: 既存ユーザーにIdentityを追加する")
    void registerUserOIDC_addsIdentityToExistingUser() {
        User existing = createUser(5, "既存ユーザー", "existing@example.com", true);
        when(userRepository.findByEmail("existing@example.com")).thenReturn(Optional.of(existing));

        User result = userService.registerUserOIDC("既存ユーザー", "existing@example.com", "google", "sub456");

        assertEquals(5, result.getId());
        verify(userRepository, never()).save(any());
        verify(userIdentityService).registerUserIdentity(existing, "google", "sub456");
    }

    @Test
    @DisplayName("findLoginUserBySub: LoginUserDtoを返す")
    void findLoginUserBySub_returnsDto() {
        User user = createUser(1, "テスト", "test@example.com", true);
        when(userIdentityService.findUserBySub("sub123")).thenReturn(user);

        LoginUserDto dto = userService.findLoginUserBySub("sub123");

        assertEquals("sub123", dto.getSub());
        assertEquals("テスト", dto.getName());
        assertEquals("test@example.com", dto.getEmail());
    }

    @Test
    @DisplayName("updateUser: プロフィールを更新する")
    void updateUser_updatesProfile() {
        User user = createUser(1, "旧名", "test@example.com", true);
        when(userIdentityService.findUserBySub("sub123")).thenReturn(user);

        ProfileForm form = new ProfileForm("新名", "新しいBio");
        userService.updateUser(form, "sub123");

        verify(userRepository).save(user);
        assertEquals("新名", user.getName());
        assertEquals("新しいBio", user.getBio());
    }

    @Test
    @DisplayName("activeUser: ユーザーをアクティブにする")
    void activeUser_activatesUser() {
        User user = createUser(1, "テスト", "test@example.com", false);
        when(userRepository.findByEmail("test@example.com")).thenReturn(Optional.of(user));

        userService.activeUser("test@example.com");

        assertTrue(user.getIsActive());
        verify(userRepository).save(user);
    }

    @Test
    @DisplayName("activeUser: 既にアクティブなら何もしない")
    void activeUser_skipsIfAlreadyActive() {
        User user = createUser(1, "テスト", "test@example.com", true);
        when(userRepository.findByEmail("test@example.com")).thenReturn(Optional.of(user));

        userService.activeUser("test@example.com");

        verify(userRepository, never()).save(any());
    }

    @Test
    @DisplayName("checkUserIsActive: アクティブなユーザーでは例外なし")
    void checkUserIsActive_passesForActiveUser() {
        User user = createUser(1, "テスト", "test@example.com", true);
        when(userRepository.findByEmail("test@example.com")).thenReturn(Optional.of(user));

        assertDoesNotThrow(() -> userService.checkUserIsActive("test@example.com"));
    }

    @Test
    @DisplayName("checkUserIsActive: 非アクティブの場合例外")
    void checkUserIsActive_throwsForInactiveUser() {
        User user = createUser(1, "テスト", "test@example.com", false);
        when(userRepository.findByEmail("test@example.com")).thenReturn(Optional.of(user));

        RuntimeException ex = assertThrows(RuntimeException.class,
                () -> userService.checkUserIsActive("test@example.com"));
        assertEquals("メール認証は完了していないためログインできません。", ex.getMessage());
    }

    @Test
    @DisplayName("getTotalUserCount: ユーザー総数を返す")
    void getTotalUserCount_returnsCount() {
        when(userRepository.count()).thenReturn(42L);

        Long count = userService.getTotalUserCount();

        assertEquals(42L, count);
    }
}
