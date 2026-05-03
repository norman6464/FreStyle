-- PHP 学習教材シードデータ
-- 実行: EC2 踏み台サーバー経由で RDS に接続して実行
-- 冪等性: INSERT IGNORE で重複スキップ

INSERT IGNORE INTO php_exercises
  (id, order_index, category, title, description, starter_code, hint_text, expected_output, created_at, updated_at)
VALUES
(1, 1, '基礎', 'こんにちは世界',
'PHPでの最初のプログラムです。\n`echo` を使って「Hello, World!」と表示してみましょう。\n\n`echo` は文字列を画面に出力する命令です。',
'<?php
// ここにコードを書いてください
echo "Hello, World!\n";
',
'`echo "文字列";` の形式で書きます。改行は `\n` を使います。',
'Hello, World!',
NOW(), NOW()),

(2, 2, '基礎', '変数と文字列',
'変数を使って名前を格納し、文字列を連結して表示しましょう。\n\n変数は `$` で始まります。文字列の連結は `.` を使います。',
'<?php
$name = "フレスタイル";
$message = "ようこそ、" . $name . " へ！";
echo $message . "\n";
',
'変数は `$変数名 = 値;` で宣言。文字列連結は `.` オペレータを使います。',
'ようこそ、フレスタイル へ！',
NOW(), NOW()),

(3, 3, '基礎', '数値計算',
'PHPで四則演算を行いましょう。\n\n整数・小数の演算や、余り (`%`)、べき乗 (`**`) も試してみてください。',
'<?php
$a = 10;
$b = 3;

echo $a + $b . "\n";  // 足し算
echo $a - $b . "\n";  // 引き算
echo $a * $b . "\n";  // 掛け算
echo $a / $b . "\n";  // 割り算
echo $a % $b . "\n";  // 余り
echo $a ** $b . "\n"; // べき乗
',
'`/` は小数を返します。整数の割り算には `intdiv()` が便利です。',
'13
7
30
3.3333333333333
1
1000',
NOW(), NOW()),

(4, 4, '制御構文', '条件分岐（if文）',
'点数に応じて成績を判定するプログラムを作りましょう。\n\n`if` / `elseif` / `else` を使って、条件によって処理を分岐させます。',
'<?php
$score = 75;

if ($score >= 90) {
    echo "優\n";
} elseif ($score >= 70) {
    echo "良\n";
} elseif ($score >= 60) {
    echo "可\n";
} else {
    echo "不可\n";
}
',
'`>=` は「以上」の比較演算子です。条件は上から順に評価されます。',
'良',
NOW(), NOW()),

(5, 5, '配列', '配列の基本',
'配列を作成して要素を操作しましょう。\n\nPHPの配列は `[]` または `array()` で作成できます。',
'<?php
$fruits = ["りんご", "バナナ", "いちご"];

echo $fruits[0] . "\n";        // 先頭要素
echo count($fruits) . "\n";    // 要素数

$fruits[] = "メロン";           // 末尾に追加
echo count($fruits) . "\n";

foreach ($fruits as $fruit) {
    echo $fruit . "\n";
}
',
'インデックスは 0 から始まります。`count()` で要素数を取得できます。',
'りんご
3
4
りんご
バナナ
いちご
メロン',
NOW(), NOW()),

(6, 6, '制御構文', 'for文',
'for 文を使って 1 から 10 の合計を計算しましょう。\n\nまた、九九の一部も表示してみてください。',
'<?php
// 1〜10の合計
$sum = 0;
for ($i = 1; $i <= 10; $i++) {
    $sum += $i;
}
echo "合計: " . $sum . "\n";

// 5の段
for ($i = 1; $i <= 9; $i++) {
    echo "5 × " . $i . " = " . (5 * $i) . "\n";
}
',
'`$i++` は `$i = $i + 1` と同じです。`+=` は加算代入演算子です。',
'合計: 55
5 × 1 = 5
5 × 2 = 10
5 × 3 = 15
5 × 4 = 20
5 × 5 = 25
5 × 6 = 30
5 × 7 = 35
5 × 8 = 40
5 × 9 = 45',
NOW(), NOW()),

(7, 7, '制御構文', 'while文',
'while 文を使って、ユーザーに数字を予測するゲームのシミュレーションを作りましょう。\n\n3 の倍数だけを表示するプログラムです。',
'<?php
$i = 1;
while ($i <= 30) {
    if ($i % 3 === 0) {
        echo $i . "\n";
    }
    $i++;
}
',
'`%` は余りを求める演算子。`=== 0` で「余りがちょうど 0」と比較します。',
'3
6
9
12
15
18
21
24
27
30',
NOW(), NOW()),

(8, 8, '配列', 'foreach文と連想配列',
'連想配列（キーと値のペア）を使って、商品リストを作りましょう。\n\n`foreach` でキーと値を両方取り出せます。',
'<?php
$prices = [
    "コーヒー" => 300,
    "紅茶"     => 250,
    "ジュース" => 200,
];

foreach ($prices as $item => $price) {
    echo $item . ": " . $price . "円\n";
}

// 合計金額
$total = array_sum($prices);
echo "合計: " . $total . "円\n";
',
'連想配列は `[キー => 値]` の形式。`array_sum()` で数値の合計を計算できます。',
'コーヒー: 300円
紅茶: 250円
ジュース: 200円
合計: 750円',
NOW(), NOW()),

(9, 9, '関数', '関数の定義と呼び出し',
'関数を使ってコードを再利用可能にしましょう。\n\n体重 (kg) と身長 (m) から BMI を計算する関数を作ってください。',
'<?php
function calcBmi(float $weight, float $height): float {
    return round($weight / ($height ** 2), 1);
}

function judgeBmi(float $bmi): string {
    if ($bmi < 18.5) return "低体重";
    if ($bmi < 25.0) return "普通体重";
    if ($bmi < 30.0) return "肥満(1度)";
    return "肥満(2度以上)";
}

$bmi = calcBmi(65, 1.70);
echo "BMI: " . $bmi . "\n";
echo "判定: " . judgeBmi($bmi) . "\n";
',
'`round($value, $digits)` で小数点以下 $digits 桁に丸めます。型宣言 `float` を使うと型安全になります。',
'BMI: 22.5
判定: 普通体重',
NOW(), NOW()),

(10, 10, '文字列', '文字列操作',
'PHPの組み込み文字列関数を使いましょう。\n\n名前の変換・検索・置換など実用的な操作を練習します。',
'<?php
$text = "Hello, PHP World!";

echo strlen($text) . "\n";                    // 文字列の長さ
echo strtoupper($text) . "\n";               // 大文字変換
echo strtolower($text) . "\n";               // 小文字変換
echo str_replace("PHP", "FreStyle", $text) . "\n"; // 置換
echo substr($text, 7, 3) . "\n";             // 部分文字列
echo strpos($text, "PHP") . "\n";            // 検索位置
',
'`strlen()` はバイト数を返します。日本語の場合は `mb_strlen()` を使いましょう。',
'17
HELLO, PHP WORLD!
hello, php world!
Hello, FreStyle World!
PHP
7',
NOW(), NOW()),

(11, 11, 'OOP', 'クラスの基本',
'クラスを使ってオブジェクト指向プログラミングを体験しましょう。\n\n「Person」クラスを作り、名前と挨拶を管理します。',
'<?php
class Person {
    private string $name;
    private int $age;

    public function __construct(string $name, int $age) {
        $this->name = $name;
        $this->age  = $age;
    }

    public function greet(): string {
        return "こんにちは！私は " . $this->name . "（" . $this->age . "歳）です。";
    }

    public function getName(): string {
        return $this->name;
    }
}

$person = new Person("田中太郎", 25);
echo $person->greet() . "\n";
echo "名前: " . $person->getName() . "\n";
',
'`__construct()` はインスタンス生成時に呼ばれる特別なメソッドです。`$this` はそのオブジェクト自身を指します。',
'こんにちは！私は 田中太郎（25歳）です。
名前: 田中太郎',
NOW(), NOW()),

(12, 12, 'OOP', '継承',
'継承を使って既存のクラスを拡張しましょう。\n\n「Animal」クラスを継承した「Dog」と「Cat」を作ります。',
'<?php
class Animal {
    protected string $name;

    public function __construct(string $name) {
        $this->name = $name;
    }

    public function speak(): string {
        return $this->name . " が鳴きました";
    }
}

class Dog extends Animal {
    public function speak(): string {
        return $this->name . "：ワンワン！";
    }
}

class Cat extends Animal {
    public function speak(): string {
        return $this->name . "：ニャーニャー！";
    }
}

$animals = [new Dog("ポチ"), new Cat("タマ"), new Dog("シロ")];
foreach ($animals as $animal) {
    echo $animal->speak() . "\n";
}
',
'`extends` で継承、`parent::メソッド()` で親クラスのメソッドを呼べます。同じメソッド名で挙動を変えることを「オーバーライド」と呼びます。',
'ポチ：ワンワン！
タマ：ニャーニャー！
シロ：ワンワン！',
NOW(), NOW());
