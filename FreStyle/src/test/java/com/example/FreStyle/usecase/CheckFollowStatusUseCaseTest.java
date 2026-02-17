package com.example.FreStyle.usecase;

import static org.assertj.core.api.Assertions.assertThat;
import static org.mockito.Mockito.when;

import java.util.Map;

import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import com.example.FreStyle.repository.FriendshipRepository;

@ExtendWith(MockitoExtension.class)
@DisplayName("CheckFollowStatusUseCase テスト")
class CheckFollowStatusUseCaseTest {

    @Mock
    private FriendshipRepository friendshipRepository;

    @InjectMocks
    private CheckFollowStatusUseCase checkFollowStatusUseCase;

    @Test
    @DisplayName("相互フォロー状態を確認できる")
    void execute_mutualFollow() {
        when(friendshipRepository.existsByFollowerIdAndFollowingId(1, 2)).thenReturn(true);
        when(friendshipRepository.existsByFollowerIdAndFollowingId(2, 1)).thenReturn(true);

        Map<String, Object> result = checkFollowStatusUseCase.execute(1, 2);

        assertThat(result.get("isFollowing")).isEqualTo(true);
        assertThat(result.get("isFollowedBy")).isEqualTo(true);
        assertThat(result.get("isMutual")).isEqualTo(true);
    }

    @Test
    @DisplayName("片方のみフォローの状態を確認できる")
    void execute_oneWayFollow() {
        when(friendshipRepository.existsByFollowerIdAndFollowingId(1, 2)).thenReturn(true);
        when(friendshipRepository.existsByFollowerIdAndFollowingId(2, 1)).thenReturn(false);

        Map<String, Object> result = checkFollowStatusUseCase.execute(1, 2);

        assertThat(result.get("isFollowing")).isEqualTo(true);
        assertThat(result.get("isFollowedBy")).isEqualTo(false);
        assertThat(result.get("isMutual")).isEqualTo(false);
    }

    @Test
    @DisplayName("フォローなしの状態を確認できる")
    void execute_noFollow() {
        when(friendshipRepository.existsByFollowerIdAndFollowingId(1, 2)).thenReturn(false);
        when(friendshipRepository.existsByFollowerIdAndFollowingId(2, 1)).thenReturn(false);

        Map<String, Object> result = checkFollowStatusUseCase.execute(1, 2);

        assertThat(result.get("isFollowing")).isEqualTo(false);
        assertThat(result.get("isFollowedBy")).isEqualTo(false);
        assertThat(result.get("isMutual")).isEqualTo(false);
    }
}
