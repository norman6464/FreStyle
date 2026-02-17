package com.example.FreStyle.usecase;

import static org.assertj.core.api.Assertions.assertThatThrownBy;
import static org.mockito.Mockito.verify;
import static org.mockito.Mockito.when;

import java.util.Optional;

import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import com.example.FreStyle.entity.Friendship;
import com.example.FreStyle.entity.User;
import com.example.FreStyle.exception.ResourceNotFoundException;
import com.example.FreStyle.repository.FriendshipRepository;

@ExtendWith(MockitoExtension.class)
@DisplayName("UnfollowUserUseCase テスト")
class UnfollowUserUseCaseTest {

    @Mock
    private FriendshipRepository friendshipRepository;

    @InjectMocks
    private UnfollowUserUseCase unfollowUserUseCase;

    private User follower;
    private User following;

    @BeforeEach
    void setUp() {
        follower = new User();
        follower.setId(1);
        following = new User();
        following.setId(2);
    }

    @Test
    @DisplayName("フォローを解除できる")
    void execute_unfollowsUser() {
        Friendship friendship = new Friendship();
        friendship.setId(10);
        friendship.setFollower(follower);
        friendship.setFollowing(following);

        when(friendshipRepository.findByFollowerIdAndFollowingId(1, 2))
                .thenReturn(Optional.of(friendship));

        unfollowUserUseCase.execute(follower, 2);

        verify(friendshipRepository).delete(friendship);
    }

    @Test
    @DisplayName("フォローしていない場合は例外をスローする")
    void execute_notFollowing() {
        when(friendshipRepository.findByFollowerIdAndFollowingId(1, 2))
                .thenReturn(Optional.empty());

        assertThatThrownBy(() -> unfollowUserUseCase.execute(follower, 2))
                .isInstanceOf(ResourceNotFoundException.class);
    }
}
