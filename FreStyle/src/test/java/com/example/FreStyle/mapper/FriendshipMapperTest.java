package com.example.FreStyle.mapper;

import static org.assertj.core.api.Assertions.assertThat;

import java.time.LocalDateTime;

import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Nested;
import org.junit.jupiter.api.Test;

import com.example.FreStyle.dto.FriendshipDto;
import com.example.FreStyle.entity.Friendship;
import com.example.FreStyle.entity.User;

/**
 * FriendshipMapper のDTO変換ロジック検証。
 * フォロー関係を表すFriendshipエンティティを、
 * 「誰をフォローしているか（Following視点）」と「誰にフォローされているか（Follower視点）」
 * のどちらの視点のDTOに変換するかで挙動が変わることを網羅する。
 */
@DisplayName("FriendshipMapper")
class FriendshipMapperTest {

    private final FriendshipMapper mapper = new FriendshipMapper();

    private User createUser(Integer id, String name, String iconUrl, String bio, String status) {
        User user = new User();
        user.setId(id);
        user.setName(name);
        user.setIconUrl(iconUrl);
        user.setBio(bio);
        user.setStatus(status);
        user.setEmail("u" + id + "@example.com");
        return user;
    }

    private Friendship createFriendship(Integer id, User follower, User following, LocalDateTime createdAt) {
        Friendship friendship = new Friendship();
        friendship.setId(id);
        friendship.setFollower(follower);
        friendship.setFollowing(following);
        friendship.setCreatedAt(createdAt);
        return friendship;
    }

    @Nested
    @DisplayName("toFollowingDto：フォロー中のユーザー情報をDTO化")
    class ToFollowingDtoTest {

        @Test
        @DisplayName("followingユーザーの情報がDTOに反映される")
        void mapsFollowingUser() {
            User follower = createUser(1, "自分", "me.png", "自己紹介", "online");
            User following = createUser(2, "相手", "you.png", "相手の自己紹介", "away");
            Friendship friendship = createFriendship(100, follower, following, LocalDateTime.of(2026, 1, 1, 12, 0));

            FriendshipDto dto = mapper.toFollowingDto(friendship, true);

            assertThat(dto.id()).isEqualTo(100);
            assertThat(dto.userId()).isEqualTo(2);
            assertThat(dto.username()).isEqualTo("相手");
            assertThat(dto.iconUrl()).isEqualTo("you.png");
            assertThat(dto.bio()).isEqualTo("相手の自己紹介");
            assertThat(dto.status()).isEqualTo("away");
            assertThat(dto.mutual()).isTrue();
            assertThat(dto.createdAt()).isEqualTo("2026-01-01T12:00");
        }

        @Test
        @DisplayName("mutual=false の場合は相互フォローでないとしてDTOに反映される")
        void mutualFalseIsReflected() {
            User follower = createUser(1, "me", null, null, null);
            User following = createUser(2, "you", null, null, null);
            Friendship friendship = createFriendship(1, follower, following, LocalDateTime.now());

            FriendshipDto dto = mapper.toFollowingDto(friendship, false);

            assertThat(dto.mutual()).isFalse();
        }
    }

    @Nested
    @DisplayName("toFollowerDto：フォロワーユーザー情報をDTO化")
    class ToFollowerDtoTest {

        @Test
        @DisplayName("followerユーザーの情報がDTOに反映される")
        void mapsFollowerUser() {
            User follower = createUser(3, "フォロワー", "f.png", "紹介", "busy");
            User following = createUser(4, "自分", "me.png", "me", null);
            Friendship friendship = createFriendship(200, follower, following, LocalDateTime.of(2026, 2, 14, 9, 30));

            FriendshipDto dto = mapper.toFollowerDto(friendship, true);

            assertThat(dto.id()).isEqualTo(200);
            assertThat(dto.userId()).isEqualTo(3);
            assertThat(dto.username()).isEqualTo("フォロワー");
            assertThat(dto.iconUrl()).isEqualTo("f.png");
            assertThat(dto.bio()).isEqualTo("紹介");
            assertThat(dto.status()).isEqualTo("busy");
            assertThat(dto.mutual()).isTrue();
            assertThat(dto.createdAt()).isEqualTo("2026-02-14T09:30");
        }
    }

    @Nested
    @DisplayName("エッジケース")
    class EdgeCases {

        @Test
        @DisplayName("createdAtがnullのとき、DTOのcreatedAtもnullになる")
        void nullCreatedAtIsPassedThrough() {
            User follower = createUser(1, "a", null, null, null);
            User following = createUser(2, "b", null, null, null);
            Friendship friendship = createFriendship(1, follower, following, null);

            FriendshipDto dto = mapper.toFollowingDto(friendship, false);

            assertThat(dto.createdAt()).isNull();
        }

        @Test
        @DisplayName("ユーザーのiconUrl/bio/statusがnullのとき、DTOにもnullで透過される")
        void nullUserAttributesArePassedThrough() {
            User follower = createUser(1, "name", null, null, null);
            User following = createUser(2, "target", null, null, null);
            Friendship friendship = createFriendship(1, follower, following, LocalDateTime.now());

            FriendshipDto dto = mapper.toFollowingDto(friendship, false);

            assertThat(dto.iconUrl()).isNull();
            assertThat(dto.bio()).isNull();
            assertThat(dto.status()).isNull();
        }
    }
}
