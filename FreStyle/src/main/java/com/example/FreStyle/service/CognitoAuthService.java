package com.example.FreStyle.service;

import java.security.InvalidParameterException;
import java.util.ArrayList;
import java.util.Base64;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

import javax.crypto.Mac;
import javax.crypto.spec.SecretKeySpec;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Service;

import software.amazon.awssdk.auth.credentials.AwsBasicCredentials;
import software.amazon.awssdk.auth.credentials.StaticCredentialsProvider;
import software.amazon.awssdk.regions.Region;
import software.amazon.awssdk.services.cognitoidentityprovider.CognitoIdentityProviderClient;
import software.amazon.awssdk.services.cognitoidentityprovider.model.*;

/**
 * AWS Cognito 認証系サービス
 * - サインアップ
 * - メール確認
 * - ログイン
 */
@Service
public class CognitoAuthService {

    private final CognitoIdentityProviderClient cognitoClient;

    @Value("${cognito.client-id}")
    private String clientId;

    @Value("${cognito.client-secret}")
    private String clientSecret;

    public CognitoAuthService(
            @Value("${aws.access-key}") String accessKey,
            @Value("${aws.secret-key}") String secretKey,
            @Value("${aws.region}") String region) {

        this.cognitoClient = CognitoIdentityProviderClient.builder()
                .credentialsProvider(StaticCredentialsProvider.create(
                        AwsBasicCredentials.create(accessKey, secretKey)))
                .region(Region.of(region))
                .build();
    }
    
    
    public void updateUserProfile(String accessToken, String name) {
      try {
        // 更新対象属性をリストアップ
        List<AttributeType> attrsBuilder = new ArrayList<>();
        if (name != null && !name.isEmpty()) {
          attrsBuilder.add(AttributeType.builder().name("name").value(name).build());
        }
        
        if (attrsBuilder.isEmpty()) {
          throw new IllegalArgumentException("更新する属性が指定されていません。");
        }
         
        // CognitoのUpdateUserAttributes APIの呼び出し
        UpdateUserAttributesRequest request = UpdateUserAttributesRequest.builder()
                  .accessToken(accessToken)
                  .userAttributes(attrsBuilder)
                  .build();
                  
        cognitoClient.updateUserAttributes(request);
        System.out.println("Profile update");
      } catch (NotAuthorizedException e) {
        System.out.println("NotAuthorizedException");
        throw new RuntimeException("invalid parameter retry login");
      } catch (InvalidParameterException e) {
        System.out.println("InvalidParameterException");
        throw new RuntimeException("invalid parameter");
      } catch (InvalidUserPoolConfigurationException e) {
        System.out.println("InvalidUserPoolConfigurationException");
        throw new RuntimeException("user configuration setting");
      } catch (Exception e) {
        throw new RuntimeException("proceed update profile error");
      }
    }

    // サインアップ
    public void signUpUser(String email, String password, String name) {
        AttributeType emailAttr = AttributeType.builder()
                .name("email").value(email).build();

        AttributeType nameAttr = AttributeType.builder()
                .name("name").value(name).build();

        SignUpRequest request = SignUpRequest.builder()
                .clientId(clientId)
                .secretHash(calculateSecretHash(email))
                .username(email)
                .password(password)
                .userAttributes(emailAttr, nameAttr)
                .build();

        try {
            SignUpResponse response = cognitoClient.signUp(request);
            if (response.userConfirmed()) {
                System.out.println("ユーザーは既に確認済みです。");
            } else {
                System.out.println("確認コードを送信しました。");
            }
        } catch (UsernameExistsException e) {
            throw new RuntimeException("このメールアドレスは既に登録されています。");
        } catch (InvalidPasswordException e) {
            throw new RuntimeException("パスワードが要件を満たしていません。");
        } catch (InvalidParameterException e) {
            throw new RuntimeException("入力値が無効です。");
        } catch (Exception e) {
            throw new RuntimeException("サインアップ中にエラーが発生しました: " + e.getMessage(), e);
        }
    }

    // ユーザー確認（メール認証）
    public void confirmUserSignup(String email, String confirmationCode) {
        ConfirmSignUpRequest request = ConfirmSignUpRequest.builder()
                .clientId(clientId)
                .secretHash(calculateSecretHash(email))
                .username(email)
                .confirmationCode(confirmationCode)
                .build();

        try {
            ConfirmSignUpResponse response = cognitoClient.confirmSignUp(request);
            System.out.println("ユーザー確認に成功しました: " + response);
        } catch (UserNotFoundException e) {
            throw new RuntimeException("ユーザーが見つかりません。");
        } catch (CodeMismatchException e) {
            throw new RuntimeException("確認コードが正しくありません。");
        } catch (ExpiredCodeException e) {
            throw new RuntimeException("確認コードの有効期限が切れています。");
        } catch (Exception e) {
            throw new RuntimeException("ユーザー確認中にエラーが発生しました: " + e.getMessage(), e);
        }
    }

    // ログイン
    public Map<String, String> login(String email, String password) {
        Map<String, String> authParams = new HashMap<>();
        authParams.put("USERNAME", email);
        authParams.put("PASSWORD", password);
        authParams.put("SECRET_HASH", calculateSecretHash(email));

        InitiateAuthRequest authRequest = InitiateAuthRequest.builder()
                .authFlow(AuthFlowType.USER_PASSWORD_AUTH)
                .clientId(clientId)
                .authParameters(authParams)
                .build();

        try {
            InitiateAuthResponse response = cognitoClient.initiateAuth(authRequest);
            AuthenticationResultType result = response.authenticationResult();

            Map<String, String> tokens = new HashMap<>();
            tokens.put("accessToken", result.accessToken());
            tokens.put("idToken", result.idToken());
            tokens.put("refreshToken", result.refreshToken());
            return tokens;

        } catch (NotAuthorizedException e) {
            throw new RuntimeException("メールアドレスまたはパスワードが間違っています。");
        } catch (UserNotConfirmedException e) {
            throw new RuntimeException("メール確認が完了していません。");
        } catch (Exception e) {
            throw new RuntimeException("ログイン中にエラーが発生しました: " + e.getMessage(), e);
        }
    }
    
    
    public void forgotPassword(String email) {
        ForgotPasswordRequest request = ForgotPasswordRequest.builder()
        .clientId(clientId)
        .username(email)
        .secretHash(calculateSecretHash(email))
        .build();
        
        try {
            cognitoClient.forgotPassword(request);
            System.out.println("send confirm code" + email);
        } catch (UserNotFoundException e) {
            System.out.println("this mail not register");
            throw new RuntimeException("このメールアドレスは登録されていません。");
        } catch (Exception e) {
            System.out.println("password reset request failed");
            throw new RuntimeException("パスワードリセットのリクエストに失敗しました。" + e.getMessage());
        }
    }
    
    public void confirmForgotPassword(String email, String confirmationCode, String newPassword) {
        ConfirmForgotPasswordRequest request = ConfirmForgotPasswordRequest.builder()
            .clientId(clientId)
            .username(email)
            .secretHash(calculateSecretHash(email))
            .confirmationCode(confirmationCode)
            .password(newPassword)
            .build();
            
            try {
                cognitoClient.confirmForgotPassword(request);
                System.out.println("password reset finish");
            } catch (CodeMismatchException e) {
                System.out.println("CodeMismatchException class error");
                throw new RuntimeException("確認コードが間違っています。");
            } catch (ExpiredCodeException e) {
                System.out.println("ExpiredCodeException class error");
                throw new RuntimeException("確認コードの有効期限が切れています。");
            } catch (Exception e) {
                System.out.println("password reset processing error");
                throw new RuntimeException("パスワードがリセット中にエラーが発生しました。");
            }
    }
    
    // ハッシュ値を計算をする
    // Cognitoのアプリクライアントでシークレットが有効になっている場合はすべてにSecret_Hashを付与する必要がある
    private String calculateSecretHash(String username) {
        try {
            String message = username + clientId;
            Mac mac = Mac.getInstance("HmacSHA256");
            SecretKeySpec keySpec = new SecretKeySpec(clientSecret.getBytes("UTF-8"), "HmacSHA256");
            mac.init(keySpec);
            byte[] rawHmac = mac.doFinal(message.getBytes("UTF-8"));
            return Base64.getEncoder().encodeToString(rawHmac);
        } catch (Exception e) {
            throw new RuntimeException("SECRET_HASHの計算中にエラーが発生しました。", e);
        }
    }
    
    // -----------------------
    // リフレッシュトークンでアクセストークン、IDトークンを再発行
    // だがリフレッシュトークンのローテーションの設定をしていないのでリフレッシュトークンの取得はできない
    // -----------------------
    public Map<String, String> refreshAccessToken(String refreshToken, String username) {
        System.out.println("[CognitoAuthService refreshAccessToken] リフレッシュトークンでのアクセストークン再発行を開始");
        System.out.println("[CognitoAuthService refreshAccessToken] REFRESH_TOKEN: " + 
                          (refreshToken != null ? refreshToken.substring(0, Math.min(20, refreshToken.length())) + "..." : "null"));
        
        Map<String, String> authParams = new HashMap<>();
        authParams.put("REFRESH_TOKEN", refreshToken);
        
        String secretHash = calculateSecretHash(username);
        authParams.put("SECRET_HASH", secretHash);
        System.out.println("[CognitoAuthService refreshAccessToken] SECRET_HASH: " + secretHash.substring(0, Math.min(20, secretHash.length())) + "...");
        
        InitiateAuthRequest authRequest = InitiateAuthRequest.builder()
            .authFlow(AuthFlowType.REFRESH_TOKEN)
            .clientId(clientId)
            .authParameters(authParams)
            .build();
        
        System.out.println("[CognitoAuthService refreshAccessToken] Cognito initializeAuth API を呼び出し中...");
        try {
            InitiateAuthResponse response = cognitoClient.initiateAuth(authRequest);
            AuthenticationResultType result = response.authenticationResult();
            
            System.out.println("[CognitoAuthService refreshAccessToken] ✅ トークン再発行成功");
            
            Map<String, String> tokens = new HashMap<>();
            
            tokens.put("accessToken", result.accessToken());
            tokens.put("idToken", result.idToken());
            System.out.println("[CognitoAuthService refreshAccessToken] accessToken: " + 
                              (result.accessToken() != null ? result.accessToken().substring(0, Math.min(20, result.accessToken().length())) + "..." : "null"));
            return tokens;
        } catch (NotAuthorizedException e) {
            System.out.println("[CognitoAuthService refreshAccessToken] ❌ NotAuthorizedException: " + e.getMessage());
            throw new RuntimeException("リフレッシュトークンが無効です。再ログインしてください。");
        } catch (Exception e) {
            System.out.println("[CognitoAuthService refreshAccessToken] ❌ Exception: " + e.getClass().getSimpleName() + " - " + e.getMessage());
            e.printStackTrace();
            throw new RuntimeException("アクセストークン再発行中にエラーが発生しました。");
        }
    }
}
