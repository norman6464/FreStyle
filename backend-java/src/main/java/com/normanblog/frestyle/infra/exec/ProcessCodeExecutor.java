package com.normanblog.frestyle.infra.exec;

import com.normanblog.frestyle.dto.CodeExecuteResponse;
import java.io.IOException;
import java.io.InputStream;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.ArrayList;
import java.util.Comparator;
import java.util.List;
import java.util.Map;
import java.util.concurrent.TimeUnit;
import java.util.regex.Pattern;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

/**
 * ユーザーコードを子プロセスとして実行する本番サンドボックス。java / php / go に対応する。
 *
 * <p>セキュリティの肝:
 *
 * <ul>
 *   <li><b>環境クリーン化</b>: 子プロセスに backend のシークレット(DB/AWS/Cognito 等)を一切渡さない
 *       (PATH/LANG + 言語ごとの最小限だけ通す)。これが無いと提出コードから機密が抜ける
 *   <li><b>言語別ハードニング</b>: php は disable_functions / open_basedir、go は単一ファイル実行で隔離
 *   <li><b>timeout</b>: 上限秒で子孫ごと destroyForcibly。無限ループを止める
 *   <li><b>出力上限</b>: stdout/stderr を上限バイトで打ち切り、巨大出力でメモリを潰されない
 *   <li><b>JVM ヒープ上限 / 専用 temp / 非 root</b>
 * </ul>
 *
 * <p>ネットワーク egress の遮断は infra(セキュリティグループ等)側で担保する前提。
 */
public class ProcessCodeExecutor implements CodeExecutor {

  private static final Logger log = LoggerFactory.getLogger(ProcessCodeExecutor.class);

  // 実行を禁止する PHP 関数(ファイル操作・OS 操作・ネットワーク・env 露出などを封じる)。
  private static final String PHP_DISABLE_FUNCTIONS =
      String.join(
          ",",
          "exec", "system", "shell_exec", "passthru", "popen", "proc_open",
          "pcntl_exec", "proc_get_status", "proc_terminate", "proc_close",
          "file_get_contents", "file_put_contents", "file", "fopen", "fwrite",
          "fread", "fclose", "fgets", "fputs", "file_exists", "unlink",
          "rename", "copy", "mkdir", "rmdir", "opendir", "readdir", "closedir",
          "glob", "scandir", "tempnam", "tmpfile",
          "socket_create", "fsockopen", "pfsockopen",
          "curl_init", "curl_exec", "curl_multi_exec",
          "dl", "phpinfo", "posix_kill", "posix_setuid",
          "getenv", "putenv", "apache_getenv",
          "syslog", "openlog", "closelog");

  // go の package 宣言を「行頭(空白許容)」で判定する(文字列/コメント内の誤通過を防ぐ)。
  private static final Pattern GO_PACKAGE_MAIN = Pattern.compile("(?m)^\\s*package\\s+main\\b");

  private final long timeoutSeconds;
  private final int maxOutputBytes;

  public ProcessCodeExecutor(long timeoutSeconds, int maxOutputBytes) {
    this.timeoutSeconds = timeoutSeconds;
    this.maxOutputBytes = maxOutputBytes;
  }

  @Override
  public CodeExecuteResponse execute(String language, String code, String stdin) {
    if (!supportedLanguages().contains(language)) {
      return new CodeExecuteResponse("", "未対応の言語です: " + language, 1);
    }

    Path dir = null;
    try {
      dir = Files.createTempDirectory("codeexec-");
      return switch (language) {
        case "java" -> runJava(dir, code, stdin);
        case "php" -> runPhp(dir, code, stdin);
        case "go" -> runGo(dir, code, stdin);
        default -> new CodeExecuteResponse("", "未対応の言語です: " + language, 1);
      };
    } catch (IOException e) {
      log.warn("code execution setup failed", e);
      return new CodeExecuteResponse("", "実行環境を利用できません", 1);
    } finally {
      deleteRecursively(dir);
    }
  }

  // Java: 単一ファイル実行(java Main.java)。内部のクラス名は任意で良い。
  private CodeExecuteResponse runJava(Path dir, String code, String stdin) throws IOException {
    Path source = dir.resolve("Main.java");
    Files.writeString(source, code, StandardCharsets.UTF_8);

    ProcessBuilder pb =
        new ProcessBuilder(
            "java", "-Xmx64m", "-XX:+UseSerialGC", "-XX:TieredStopAtLevel=1", source.toString());
    pb.directory(dir.toFile());
    cleanEnv(pb, dir);

    return runWithLimits(pb, timeoutSeconds, stdin);
  }

  // PHP: 開始タグ必須。disable_functions / open_basedir / memory_limit / max_execution_time で堅牢化。
  private CodeExecuteResponse runPhp(Path dir, String code, String stdin) throws IOException {
    String trimmed = code.stripLeading();
    if (!trimmed.startsWith("<?php") && !trimmed.startsWith("<?=")) {
      return new CodeExecuteResponse("", "PHP コードには `<?php` 開始タグが必要です。", 1);
    }
    Path source = dir.resolve("Main.php");
    Files.writeString(source, code, StandardCharsets.UTF_8);

    ProcessBuilder pb =
        new ProcessBuilder(
            "php",
            "-d", "disable_functions=" + PHP_DISABLE_FUNCTIONS,
            "-d", "memory_limit=64M",
            "-d", "max_execution_time=5",
            "-d", "open_basedir=" + dir,
            "-d", "display_errors=stderr",
            "-d", "log_errors=0",
            // variables_order から E を外し $_ENV を空にする(getenv 抜け穴を塞ぐ)。
            "-d", "variables_order=GPCS",
            source.toString());
    pb.directory(dir.toFile());
    cleanEnv(pb, dir);

    return runWithLimits(pb, timeoutSeconds, stdin);
  }

  // Go: package main 必須。go run で単一ファイル実行。コンパイルがあるため timeout は長めにする。
  private CodeExecuteResponse runGo(Path dir, String code, String stdin) throws IOException {
    if (!GO_PACKAGE_MAIN.matcher(code).find()) {
      return new CodeExecuteResponse("", "Go コードには `package main` と `func main()` が必要です。", 1);
    }
    Path source = dir.resolve("main.go");
    Files.writeString(source, code, StandardCharsets.UTF_8);

    ProcessBuilder pb = new ProcessBuilder("go", "run", source.toString());
    pb.directory(dir.toFile());
    cleanEnv(pb, dir);
    // go は HOME / build cache が要る。専用 temp に閉じ込め、go.mod 不要の単一ファイル実行にする。
    Map<String, String> env = pb.environment();
    env.put("HOME", dir.toString());
    env.put("GOCACHE", dir.resolve(".gocache").toString());
    env.put("GOPATH", dir.resolve(".gopath").toString());
    env.put("GO111MODULE", "off");

    return runWithLimits(pb, Math.max(timeoutSeconds, 20), stdin);
  }

  // 子プロセスに親の環境(= シークレット)を継承させない。最小限だけ通す。
  private void cleanEnv(ProcessBuilder pb, Path dir) {
    pb.environment().clear();
    String path = System.getenv("PATH");
    pb.environment().put("PATH", path == null ? "/usr/local/bin:/usr/bin:/bin" : path);
    pb.environment().put("LANG", "C.UTF-8");
    pb.environment().put("TMPDIR", dir.toString());
  }

  private CodeExecuteResponse runWithLimits(ProcessBuilder pb, long timeout, String stdin) {
    Process process;
    try {
      process = pb.start();
    } catch (IOException e) {
      log.warn("failed to start process", e);
      return new CodeExecuteResponse("", "実行環境を利用できません", 1);
    }

    // stdin を流して閉じる(テストケースの入力)。無ければ即 close で EOF を渡す。
    try (var os = process.getOutputStream()) {
      if (stdin != null && !stdin.isEmpty()) {
        os.write(stdin.getBytes(StandardCharsets.UTF_8));
      }
    } catch (IOException ignored) {
      // 既にプロセスが終了して stdin が閉じている場合がある。致命ではない。
    }

    // stdout/stderr を並行に汲む(片方のパイプが詰まると deadlock するため)。
    StringBuilder stdout = new StringBuilder();
    StringBuilder stderr = new StringBuilder();
    Thread outThread = drain(process.getInputStream(), stdout);
    Thread errThread = drain(process.getErrorStream(), stderr);
    outThread.start();
    errThread.start();

    try {
      boolean finished = process.waitFor(timeout, TimeUnit.SECONDS);
      if (!finished) {
        killTree(process);
        joinQuietly(outThread);
        joinQuietly(errThread);
        return new CodeExecuteResponse(
            stdout.toString(), append(stderr, "実行がタイムアウトしました").toString(), 124);
      }
      joinQuietly(outThread);
      joinQuietly(errThread);

      return new CodeExecuteResponse(stdout.toString(), stderr.toString(), process.exitValue());
    } catch (InterruptedException e) {
      killTree(process);
      Thread.currentThread().interrupt();
      return new CodeExecuteResponse(stdout.toString(), "実行が中断されました", 1);
    }
  }

  // ユーザーコードが起動した子孫プロセスごと止める(destroyForcibly だけだと孫が残り得る)。
  private static void killTree(Process process) {
    process.descendants().forEach(ProcessHandle::destroyForcibly);
    process.destroyForcibly();
  }

  // 入力ストリームを上限バイトまで buf に読み込むスレッドを作る(start はしない)。
  private Thread drain(InputStream in, StringBuilder buf) {
    return new Thread(
        () -> {
          byte[] chunk = new byte[4096];
          int total = 0;
          try (InputStream is = in) {
            int n;
            while ((n = is.read(chunk)) != -1) {
              // 上限まではバッファに記録。上限後も read は止めず読み捨てる
              // (止めると pipe が詰まり子プロセスの write がブロックして終了できなくなる)。
              if (total < maxOutputBytes) {
                int allowed = Math.min(n, maxOutputBytes - total);
                buf.append(new String(chunk, 0, allowed, StandardCharsets.UTF_8));
                total += allowed;
              }
            }
          } catch (IOException ignored) {
            // プロセス終了でストリームが閉じるのは正常。
          }
        });
  }

  private static StringBuilder append(StringBuilder sb, String s) {
    if (sb.length() > 0) {
      sb.append('\n');
    }
    return sb.append(s);
  }

  private static void joinQuietly(Thread t) {
    try {
      t.join(2000);
    } catch (InterruptedException e) {
      Thread.currentThread().interrupt();
    }
  }

  private static void deleteRecursively(Path dir) {
    if (dir == null) {
      return;
    }
    try (var paths = Files.walk(dir)) {
      paths.sorted(Comparator.reverseOrder()).forEach(ProcessCodeExecutor::deleteQuietly);
    } catch (IOException ignored) {
      // 後始末失敗は致命ではない。
    }
  }

  private static void deleteQuietly(Path p) {
    try {
      Files.deleteIfExists(p);
    } catch (IOException ignored) {
      // ignore
    }
  }

  // 対応言語一覧。
  public static List<String> supportedLanguages() {
    List<String> langs = new ArrayList<>();
    langs.add("java");
    langs.add("php");
    langs.add("go");
    return langs;
  }
}
