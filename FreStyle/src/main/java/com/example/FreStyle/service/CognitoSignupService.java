package com.example.FreStyle.service;

import java.security.InvalidParameterException;
import java.util.Base64;

import javax.crypto.Mac;
import javax.crypto.spec.SecretKeySpec;

import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Service;

import software.amazon.awssdk.auth.credentials.AwsBasicCredentials;
import software.amazon.awssdk.auth.credentials.StaticCredentialsProvider;
import software.amazon.awssdk.regions.Region;
import software.amazon.awssdk.services.cognitoidentityprovider.CognitoIdentityProviderClient;
import software.amazon.awssdk.services.cognitoidentityprovider.model.AttributeType;
import software.amazon.awssdk.services.cognitoidentityprovider.model.CodeMismatchException;
import software.amazon.awssdk.services.cognitoidentityprovider.model.ConfirmSignUpRequest;
import software.amazon.awssdk.services.cognitoidentityprovider.model.ConfirmSignUpResponse;
import software.amazon.awssdk.services.cognitoidentityprovider.model.ExpiredCodeException;
import software.amazon.awssdk.services.cognitoidentityprovider.model.InvalidPasswordException;
import software.amazon.awssdk.services.cognitoidentityprovider.model.SignUpRequest;
import software.amazon.awssdk.services.cognitoidentityprovider.model.SignUpResponse;
import software.amazon.awssdk.services.cognitoidentityprovider.model.UserNotFoundException;
import software.amazon.awssdk.services.cognitoidentityprovider.model.UsernameExistsException;


// クライアントシークレットを有効にしている場合は、HMAC-SHA256で署名をつけてclient_id、client_secret、emailを計算してリクエストに付与するようになる

@Service
public class CognitoSignupService {
  
  // AWS Cognitoにリクエストをするためのクライアント
  private final CognitoIdentityProviderClient cognitoClient;
  
  @Value("${cognito.client-id}")
  private String clientId;
  
  @Value("${cognito.client-secret}")
  private String clientSecret;
  
  // クライアントの設定
  public CognitoSignupService(
    @Value("${aws.access-key}") String accessKey,
    @Value("${aws.secret-key}") String secretKey,
    @Value("${aws.region}") String region){
    
    this.cognitoClient = CognitoIdentityProviderClient.builder()
            .credentialsProvider(StaticCredentialsProvider.create(
                  AwsBasicCredentials.create(accessKey, secretKey))) 
                  .region(Region.of(region))
                  .build(); 
  }
  
  // サインアップ処理
  public void signUpUser(String email, String password, String name) {
    // 属性定義
    AttributeType emailAttr = AttributeType.builder()
            .name("email")
            .value(email)
            .build();
    
    AttributeType nameAttr = AttributeType.builder()
            .name("name")
            .value(name)
            .build();
            
    SignUpRequest request = SignUpRequest.builder()
            .clientId(clientId)
            .secretHash(calculateHash(email)) // Cognitoのクライアントシークレットで暗号化し
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
      throw new RuntimeException("パスワードが要件を満たして言いません");
    } catch (InvalidParameterException e) {
      throw new RuntimeException("入力値が無効です。");
    } catch (Exception e) {
      System.out.println(e.getMessage());
      throw new RuntimeException("サインアップ処理中にエラーが発生しました。" + e.getMessage());
    }
  }
  
  // サインアップ後のメール検証
  public void confirmUserSignup(String email, String confirmationCode) {
      ConfirmSignUpRequest request = ConfirmSignUpRequest.builder()
            .clientId(clientId)
            .secretHash(calculateHash(email))
            .username(email)
            .confirmationCode(confirmationCode)
            .build();
            
      try {
        ConfirmSignUpResponse response = cognitoClient.confirmSignUp(request);
          System.out.println("ユーザー確認に成功しました:" + response);
      } catch (UserNotFoundException e) {
        throw new RuntimeException("ユーザーが見つかりません。");
      } catch (CodeMismatchException e) {
        throw new RuntimeException("確認コードが正しくありません。");
      } catch (ExpiredCodeException e) {
        throw new RuntimeException("確認コードの有効期限が切れています。");
      } catch (Exception e) {
        throw new RuntimeException("ユーザー確認中にエラーが発生しました。" + e.getMessage());
      }       
  }
  
  
  // クライアントシークレットでの計算
  // SECRET_HASHを計算してサインアップリクエストに付与するには、HMAC-SHA256アルゴリズムで
  public String calculateHash(String username) {
    try {
      String message = username + clientId;
      Mac mac = Mac.getInstance("HmacSHA256");
      SecretKeySpec keySpec = new SecretKeySpec(clientSecret.getBytes("UTF-8"),"HmacSHA256");
      mac.init(keySpec);
      byte[] rawHmac = mac.doFinal(message.getBytes("UTF-8"));
      return Base64.getEncoder().encodeToString(rawHmac);
    } catch (Exception e) {
      throw new RuntimeException("Error while calculating SECRET_HASH", e);
    }
  }
  
}
