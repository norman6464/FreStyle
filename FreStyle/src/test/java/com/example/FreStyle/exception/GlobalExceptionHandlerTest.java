package com.example.FreStyle.exception;

import static org.assertj.core.api.Assertions.assertThat;
import static org.mockito.Mockito.*;

import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.junit.jupiter.MockitoExtension;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.context.request.WebRequest;

import com.example.FreStyle.dto.ErrorResponseDto;

@ExtendWith(MockitoExtension.class)
@DisplayName("GlobalExceptionHandler")
class GlobalExceptionHandlerTest {

    @InjectMocks
    private GlobalExceptionHandler handler;

    private WebRequest mockRequest(String uri) {
        WebRequest request = mock(WebRequest.class);
        when(request.getDescription(false)).thenReturn("uri=" + uri);
        return request;
    }

    @Test
    @DisplayName("ResourceNotFoundExceptionで404を返す")
    void returnsNotFoundOnResourceNotFoundException() {
        WebRequest request = mockRequest("/api/notes/999");
        ResourceNotFoundException ex = new ResourceNotFoundException("ノートが見つかりません");

        ResponseEntity<ErrorResponseDto> response = handler.handleResourceNotFoundException(ex, request);

        assertThat(response.getStatusCode()).isEqualTo(HttpStatus.NOT_FOUND);
        assertThat(response.getBody()).isNotNull();
        assertThat(response.getBody().getStatus()).isEqualTo(404);
        assertThat(response.getBody().getError()).isEqualTo("Not Found");
        assertThat(response.getBody().getMessage()).isEqualTo("ノートが見つかりません");
        assertThat(response.getBody().getPath()).isEqualTo("/api/notes/999");
    }

    @Test
    @DisplayName("UnauthorizedExceptionで403を返す")
    void returnsForbiddenOnUnauthorizedException() {
        WebRequest request = mockRequest("/api/notes/1");
        UnauthorizedException ex = new UnauthorizedException("アクセス権限がありません");

        ResponseEntity<ErrorResponseDto> response = handler.handleUnauthorizedException(ex, request);

        assertThat(response.getStatusCode()).isEqualTo(HttpStatus.FORBIDDEN);
        assertThat(response.getBody()).isNotNull();
        assertThat(response.getBody().getStatus()).isEqualTo(403);
        assertThat(response.getBody().getError()).isEqualTo("Forbidden");
        assertThat(response.getBody().getMessage()).isEqualTo("アクセス権限がありません");
        assertThat(response.getBody().getPath()).isEqualTo("/api/notes/1");
    }

    @Test
    @DisplayName("BusinessExceptionで400を返す")
    void returnsBadRequestOnBusinessException() {
        WebRequest request = mockRequest("/api/auth/signup");
        BusinessException ex = new BusinessException("メールアドレスが無効です");

        ResponseEntity<ErrorResponseDto> response = handler.handleBusinessException(ex, request);

        assertThat(response.getStatusCode()).isEqualTo(HttpStatus.BAD_REQUEST);
        assertThat(response.getBody()).isNotNull();
        assertThat(response.getBody().getStatus()).isEqualTo(400);
        assertThat(response.getBody().getError()).isEqualTo("Bad Request");
        assertThat(response.getBody().getMessage()).isEqualTo("メールアドレスが無効です");
        assertThat(response.getBody().getPath()).isEqualTo("/api/auth/signup");
    }

    @Test
    @DisplayName("汎用Exceptionで500を返す")
    void returnsInternalServerErrorOnGenericException() {
        WebRequest request = mockRequest("/api/data");
        Exception ex = new Exception("予期しないエラー");

        ResponseEntity<ErrorResponseDto> response = handler.handleGenericException(ex, request);

        assertThat(response.getStatusCode()).isEqualTo(HttpStatus.INTERNAL_SERVER_ERROR);
        assertThat(response.getBody()).isNotNull();
        assertThat(response.getBody().getStatus()).isEqualTo(500);
        assertThat(response.getBody().getError()).isEqualTo("Internal Server Error");
        assertThat(response.getBody().getMessage()).isEqualTo("サーバーエラーが発生しました。管理者にお問い合わせください。");
        assertThat(response.getBody().getPath()).isEqualTo("/api/data");
    }

    @Test
    @DisplayName("IllegalArgumentExceptionで400を返す")
    void returnsBadRequestOnIllegalArgumentException() {
        WebRequest request = mockRequest("/api/notes/1/images");
        IllegalArgumentException ex = new IllegalArgumentException("許可されていないファイル形式です");

        ResponseEntity<ErrorResponseDto> response = handler.handleIllegalArgumentException(ex, request);

        assertThat(response.getStatusCode()).isEqualTo(HttpStatus.BAD_REQUEST);
        assertThat(response.getBody()).isNotNull();
        assertThat(response.getBody().getStatus()).isEqualTo(400);
        assertThat(response.getBody().getError()).isEqualTo("Bad Request");
        assertThat(response.getBody().getMessage()).isEqualTo("許可されていないファイル形式です");
        assertThat(response.getBody().getPath()).isEqualTo("/api/notes/1/images");
    }

    @Test
    @DisplayName("IllegalStateExceptionで400を返す")
    void returnsBadRequestOnIllegalStateException() {
        WebRequest request = mockRequest("/api/chat/users/1/create");
        IllegalStateException ex = new IllegalStateException("無効な状態です");

        ResponseEntity<ErrorResponseDto> response = handler.handleIllegalStateException(ex, request);

        assertThat(response.getStatusCode()).isEqualTo(HttpStatus.BAD_REQUEST);
        assertThat(response.getBody()).isNotNull();
        assertThat(response.getBody().getStatus()).isEqualTo(400);
        assertThat(response.getBody().getError()).isEqualTo("Bad Request");
        assertThat(response.getBody().getMessage()).isEqualTo("無効な状態です");
        assertThat(response.getBody().getPath()).isEqualTo("/api/chat/users/1/create");
    }

    @Test
    @DisplayName("エラーレスポンスにタイムスタンプが含まれる")
    void responseContainsTimestamp() {
        WebRequest request = mockRequest("/api/test");
        ResourceNotFoundException ex = new ResourceNotFoundException("テスト");

        ResponseEntity<ErrorResponseDto> response = handler.handleResourceNotFoundException(ex, request);

        assertThat(response.getBody()).isNotNull();
        assertThat(response.getBody().getTimestamp()).isNotNull();
    }
}
