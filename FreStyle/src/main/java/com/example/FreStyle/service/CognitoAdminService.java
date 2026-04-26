package com.example.FreStyle.service;

import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Service;

import lombok.extern.slf4j.Slf4j;
import software.amazon.awssdk.regions.Region;
import software.amazon.awssdk.services.cognitoidentityprovider.CognitoIdentityProviderClient;
import software.amazon.awssdk.services.cognitoidentityprovider.model.AdminAddUserToGroupRequest;
import software.amazon.awssdk.services.cognitoidentityprovider.model.AdminCreateUserRequest;
import software.amazon.awssdk.services.cognitoidentityprovider.model.AdminCreateUserResponse;
import software.amazon.awssdk.services.cognitoidentityprovider.model.AttributeType;
import software.amazon.awssdk.services.cognitoidentityprovider.model.DeliveryMediumType;
import software.amazon.awssdk.services.cognitoidentityprovider.model.UsernameExistsException;

/**
 * Cognito の管理 API（招待・グループ追加・パスワードリセット等）。
 * 通常の認証フロー (CognitoAuthService) とは別に、CompanyAdmin / SuperAdmin
 * からの管理操作用に分離する。
 */
@Service
@Slf4j
public class CognitoAdminService {

    private final CognitoIdentityProviderClient cognitoClient;
    private final String userPoolId;

    public CognitoAdminService(
            @Value("${aws.region}") String region,
            @Value("${cognito.user-pool-id}") String userPoolId
    ) {
        this.userPoolId = userPoolId;
        this.cognitoClient = CognitoIdentityProviderClient.builder()
                .region(Region.of(region))
                .build();
    }

    public record AdminCreatedUser(String username, String sub, boolean alreadyExisted) {}

    /**
     * Cognito にユーザーを管理者作成し、招待メール（Cognito 標準）を送信する。
     * - 既に同一 email で登録済みなら何もせず alreadyExisted=true を返す
     * - 一時パスワードは Cognito が自動生成、メール送信に乗る
     * - ユーザーは初回ログイン時に必ずパスワード変更を要求される
     *
     * @param email      招待先メールアドレス
     * @param displayName ユーザー表示名（name 属性）
     * @return 作成された Cognito Username と sub
     */
    public AdminCreatedUser inviteUser(String email, String displayName) {
        try {
            AdminCreateUserRequest req = AdminCreateUserRequest.builder()
                    .userPoolId(userPoolId)
                    .username(email)
                    .userAttributes(
                            AttributeType.builder().name("email").value(email).build(),
                            AttributeType.builder().name("email_verified").value("true").build(),
                            AttributeType.builder().name("name").value(displayName != null ? displayName : email).build()
                    )
                    .desiredDeliveryMediums(DeliveryMediumType.EMAIL)
                    // MessageAction を未指定にすると Cognito 標準の招待メール (一時パスワード付き) が送信される
                    .build();
            AdminCreateUserResponse resp = cognitoClient.adminCreateUser(req);
            String sub = resp.user().attributes().stream()
                    .filter(a -> "sub".equals(a.name()))
                    .map(AttributeType::value)
                    .findFirst()
                    .orElse(resp.user().username());
            log.info("[CognitoAdminService] invited user email={} sub={} status={}",
                    email, sub, resp.user().userStatusAsString());
            return new AdminCreatedUser(resp.user().username(), sub, false);
        } catch (UsernameExistsException e) {
            log.warn("[CognitoAdminService] user already exists email={}", email);
            // 既存ユーザーの sub を取得するため admin-get-user を呼ぶこともできるが、
            // ここでは alreadyExisted=true で返してアプリ側に判断させる
            return new AdminCreatedUser(email, null, true);
        }
    }

    /**
     * 任意の Cognito Group にユーザーを追加（admin / company-admin / trainee 等）。
     * Group が存在しなければ事前に AWS CLI で create-group しておくこと。
     */
    public void addUserToGroup(String username, String groupName) {
        AdminAddUserToGroupRequest req = AdminAddUserToGroupRequest.builder()
                .userPoolId(userPoolId)
                .username(username)
                .groupName(groupName)
                .build();
        cognitoClient.adminAddUserToGroup(req);
        log.info("[CognitoAdminService] added user to group username={} group={}", username, groupName);
    }
}
