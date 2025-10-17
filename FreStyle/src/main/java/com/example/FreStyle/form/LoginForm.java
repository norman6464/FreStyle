package com.example.FreStyle.form;


// lombokの依存関係を追加しているがなぜか機能しなくなったので自分でコードを実装をしている
public class LoginForm {
    private String email;
    private String password;

    // コンストラクタ（引数あり）
    public LoginForm(String email, String password) {
        this.email = email;
        this.password = password;
    }
    
    public LoginForm() {
      
    }

    // ゲッター
    public String getEmail() {
        return email;
    }

    public String getPassword() {
        return password;
    }

    // セッター
    public void setEmail(String email) {
        this.email = email;
    }

    public void setPassword(String password) {
        this.password = password;
    }
}
