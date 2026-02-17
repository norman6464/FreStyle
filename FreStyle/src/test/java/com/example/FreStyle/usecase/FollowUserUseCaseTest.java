package com.example.FreStyle.usecase;

import static org.assertj.core.api.Assertions.assertThat;
import static org.assertj.core.api.Assertions.assertThatThrownBy;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.Mockito.verify;
import static org.mockito.Mockito.when;

import org.springframework.dao.DataIntegrityViolationException;

import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.Spy;
import org.mockito.junit.jupiter.MockitoExtension;

import com.example.FreStyle.dto.FriendshipDto;
import com.example.FreStyle.entity.Friendship;
import com.example.FreStyle.entity.User;
import com.example.FreStyle.mapper.FriendshipMapper;
import com.example.FreStyle.repository.FriendshipRepository;

@ExtendWith(MockitoExtension.class)
@DisplayName("FollowUserUseCase テスト")
class FollowUserUseCaseTest {

    @Mock
    private FriendshipRepository friendshipRepository;

    @Spy
    private FriendshipMapper friendshipMapper;

    @InjectMocks
    private FollowUserUseCase followUserUseCase;

    private User follower;
    private User following;

    @BeforeEach
    void setUp() {
        follower = new User();
        follower.setId(1);
        follower.setName("ユーザーA");
        following = new User();
        following.setId(2);
        following.setName("ユーザーB");
    }

    @Test
    @DisplayName("ユーザーをフォローできる")
    void execute_followsUser() {
        when(friendshipRepository.existsByFollowerIdAndFollowingId(1, 2)).thenReturn(false);
        when(friendshipRepository.existsByFollowerIdAndFollowingId(2, 1)).thenReturn(false);
        when(friendshipRepository.save(any(Friendship.class))).thenAnswer(inv -> {
            Friendship f = inv.getArgument(0);
            f.setId(100);
            return f;
        });

        FriendshipDto result = followUserUseCase.execute(follower, following);

        assertThat(result).isNotNull();
        assertThat(result.userId()).isEqualTo(2);
        assertThat(result.username()).isEqualTo("ユーザーB");
        assertThat(result.mutual()).isFalse();
        verify(friendshipRepository).save(any(Friendship.class));
    }

    @Test
    @DisplayName("相互フォローの場合mutualがtrueになる")
    void execute_mutualFollow() {
        when(friendshipRepository.existsByFollowerIdAndFollowingId(1, 2)).thenReturn(false);
        when(friendshipRepository.existsByFollowerIdAndFollowingId(2, 1)).thenReturn(true);
        when(friendshipRepository.save(any(Friendship.class))).thenAnswer(inv -> {
            Friendship f = inv.getArgument(0);
            f.setId(101);
            return f;
        });

        FriendshipDto result = followUserUseCase.execute(follower, following);

        assertThat(result).isNotNull();
        assertThat(result.mutual()).isTrue();
    }

    @Test
    @DisplayName("既にフォロー済みの場合は例外をスローする")
    void execute_alreadyFollowing() {
        when(friendshipRepository.existsByFollowerIdAndFollowingId(1, 2)).thenReturn(true);

        assertThatThrownBy(() -> followUserUseCase.execute(follower, following))
                .isInstanceOf(IllegalArgumentException.class);
    }

    @Test
    @DisplayName("自分自身はフォローできない")
    void execute_cannotFollowSelf() {
        assertThatThrownBy(() -> followUserUseCase.execute(follower, follower))
                .isInstanceOf(IllegalArgumentException.class);
    }

    @Test
    @DisplayName("レース条件でDB制約違反が発生した場合はIllegalArgumentExceptionに変換される")
    void execute_raceConditionThrowsIllegalArgument() {
        when(friendshipRepository.existsByFollowerIdAndFollowingId(1, 2)).thenReturn(false);
        when(friendshipRepository.save(any(Friendship.class)))
                .thenThrow(new DataIntegrityViolationException("Duplicate entry"));

        assertThatThrownBy(() -> followUserUseCase.execute(follower, following))
                .isInstanceOf(IllegalArgumentException.class)
                .hasMessage("既にフォローしています");
    }
}
