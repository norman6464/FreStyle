package com.example.FreStyle.usecase;

import static org.assertj.core.api.Assertions.assertThat;
import static org.mockito.Mockito.when;

import java.time.LocalDateTime;
import java.util.List;

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
@DisplayName("GetFollowersUseCase テスト")
class GetFollowersUseCaseTest {

    @Mock
    private FriendshipRepository friendshipRepository;

    @Spy
    private FriendshipMapper friendshipMapper;

    @InjectMocks
    private GetFollowersUseCase getFollowersUseCase;

    private User user;
    private User follower1;
    private User follower2;

    @BeforeEach
    void setUp() {
        user = new User();
        user.setId(1);
        user.setName("自分");
        follower1 = new User();
        follower1.setId(2);
        follower1.setName("フォロワーA");
        follower2 = new User();
        follower2.setId(3);
        follower2.setName("フォロワーB");
    }

    @Test
    @DisplayName("フォロワー一覧を取得できる")
    void execute_returnsFollowerList() {
        Friendship f1 = new Friendship(1, follower1, user, LocalDateTime.now());
        Friendship f2 = new Friendship(2, follower2, user, LocalDateTime.now());

        when(friendshipRepository.findByFollowingIdOrderByCreatedAtDesc(1))
                .thenReturn(List.of(f1, f2));
        when(friendshipRepository.findMutualFollowingIds(List.of(2, 3), 1))
                .thenReturn(List.of(2));

        List<FriendshipDto> result = getFollowersUseCase.execute(1);

        assertThat(result).hasSize(2);
        assertThat(result.get(0).getUserId()).isEqualTo(2);
        assertThat(result.get(0).isMutual()).isTrue();
        assertThat(result.get(1).getUserId()).isEqualTo(3);
        assertThat(result.get(1).isMutual()).isFalse();
    }

    @Test
    @DisplayName("フォロワーがいない場合は空リストを返す")
    void execute_returnsEmptyList() {
        when(friendshipRepository.findByFollowingIdOrderByCreatedAtDesc(1))
                .thenReturn(List.of());

        List<FriendshipDto> result = getFollowersUseCase.execute(1);

        assertThat(result).isEmpty();
    }
}
