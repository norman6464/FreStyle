package database

import (
	"log"
	"time"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// seedPhpExercises は php_exercises テーブルに初期演習データを投入する。
// ON CONFLICT DO NOTHING で冪等。起動のたびに安全に呼び出せる。
func seedPhpExercises(db *gorm.DB) error {
	now := time.Now()
	exercises := []domain.PhpExercise{
		{
			ID:             1,
			OrderIndex:     1,
			Category:       "基礎",
			Title:          "こんにちは世界",
			Description:    "PHPでの最初のプログラムです。\n`echo` を使って「Hello, World!」と表示してみましょう。\n\n`echo` は文字列を画面に出力する命令です。",
			StarterCode:    "<?php\n// ここにコードを書いてください\necho \"Hello, World!\\n\";\n",
			HintText:       "`echo \"文字列\";` の形式で書きます。改行は `\\n` を使います。",
			ExpectedOutput: "Hello, World!",
			CreatedAt:      now,
			UpdatedAt:      now,
		},
		{
			ID:             2,
			OrderIndex:     2,
			Category:       "基礎",
			Title:          "変数と文字列",
			Description:    "変数を使って名前を格納し、文字列を連結して表示しましょう。\n\n変数は `$` で始まります。文字列の連結は `.` を使います。",
			StarterCode:    "<?php\n$name = \"フレスタイル\";\n$message = \"ようこそ、\" . $name . \" へ！\";\necho $message . \"\\n\";\n",
			HintText:       "変数は `$変数名 = 値;` で宣言。文字列連結は `.` オペレータを使います。",
			ExpectedOutput: "ようこそ、フレスタイル へ！",
			CreatedAt:      now,
			UpdatedAt:      now,
		},
		{
			ID:             3,
			OrderIndex:     3,
			Category:       "基礎",
			Title:          "数値計算",
			Description:    "PHPで四則演算を行いましょう。\n\n整数・小数の演算や、余り (`%`)、べき乗 (`**`) も試してみてください。",
			StarterCode:    "<?php\n$a = 10;\n$b = 3;\necho $a + $b . \"\\n\"; // 13\necho $a - $b . \"\\n\"; // 7\necho $a * $b . \"\\n\"; // 30\necho $a / $b . \"\\n\"; // 3.333...\necho $a % $b . \"\\n\"; // 1\n",
			HintText:       "算術演算子: `+` `-` `*` `/` `%`（余り）`**`（べき乗）",
			ExpectedOutput: "13\n7\n30\n3.3333333333333\n1",
			CreatedAt:      now,
			UpdatedAt:      now,
		},
		{
			ID:             4,
			OrderIndex:     4,
			Category:       "制御構文",
			Title:          "条件分岐（if文）",
			Description:    "点数を受け取り、条件によってメッセージを変えましょう。\n\n`if` / `elseif` / `else` を使います。",
			StarterCode:    "<?php\n$score = 75;\n\nif ($score >= 80) {\n    echo \"優秀\\n\";\n} elseif ($score >= 60) {\n    echo \"合格\\n\";\n} else {\n    echo \"不合格\\n\";\n}\n",
			HintText:       "`if (条件) { ... } elseif (条件) { ... } else { ... }` の構造で書きます。",
			ExpectedOutput: "合格",
			CreatedAt:      now,
			UpdatedAt:      now,
		},
		{
			ID:             5,
			OrderIndex:     5,
			Category:       "配列",
			Title:          "配列の基本",
			Description:    "配列を宣言し、要素にアクセスしてみましょう。\n\nPHPの配列は `[]` または `array()` で作れます。",
			StarterCode:    "<?php\n$fruits = [\"りんご\", \"バナナ\", \"オレンジ\"];\n\necho $fruits[0] . \"\\n\"; // りんご\necho count($fruits) . \"\\n\"; // 3\n\n$fruits[] = \"ぶどう\";\necho count($fruits) . \"\\n\"; // 4\n",
			HintText:       "インデックスは 0 から始まります。`count()` で要素数を取得できます。",
			ExpectedOutput: "りんご\n3\n4",
			CreatedAt:      now,
			UpdatedAt:      now,
		},
		{
			ID:             6,
			OrderIndex:     6,
			Category:       "制御構文",
			Title:          "for文",
			Description:    "for 文を使って 1 から 5 までの数字を表示しましょう。",
			StarterCode:    "<?php\nfor ($i = 1; $i <= 5; $i++) {\n    echo $i . \"\\n\";\n}\n",
			HintText:       "`for (初期値; 条件; 増分) { ... }` の形式です。",
			ExpectedOutput: "1\n2\n3\n4\n5",
			CreatedAt:      now,
			UpdatedAt:      now,
		},
		{
			ID:             7,
			OrderIndex:     7,
			Category:       "制御構文",
			Title:          "while文",
			Description:    "while 文を使ってカウントダウンを表示しましょう。",
			StarterCode:    "<?php\n$count = 5;\nwhile ($count > 0) {\n    echo $count . \"\\n\";\n    $count--;\n}\necho \"発射！\\n\";\n",
			HintText:       "`while (条件) { ... }` ループ内で条件を更新するのを忘れずに。",
			ExpectedOutput: "5\n4\n3\n2\n1\n発射！",
			CreatedAt:      now,
			UpdatedAt:      now,
		},
		{
			ID:             8,
			OrderIndex:     8,
			Category:       "配列",
			Title:          "foreach文と連想配列",
			Description:    "連想配列（キーと値のペア）を作り、foreach で全要素を表示しましょう。",
			StarterCode:    "<?php\n$person = [\n    \"name\" => \"田中太郎\",\n    \"age\"  => 23,\n    \"job\"  => \"エンジニア\",\n];\n\nforeach ($person as $key => $value) {\n    echo $key . \": \" . $value . \"\\n\";\n}\n",
			HintText:       "`foreach ($array as $key => $value)` でキーと値を同時に取得できます。",
			ExpectedOutput: "name: 田中太郎\nage: 23\njob: エンジニア",
			CreatedAt:      now,
			UpdatedAt:      now,
		},
		{
			ID:             9,
			OrderIndex:     9,
			Category:       "関数",
			Title:          "関数の定義と呼び出し",
			Description:    "2つの数を受け取り、その合計を返す関数を作りましょう。",
			StarterCode:    "<?php\nfunction add($a, $b) {\n    return $a + $b;\n}\n\necho add(3, 4) . \"\\n\";  // 7\necho add(10, 20) . \"\\n\"; // 30\n",
			HintText:       "`function 関数名(引数) { return 戻り値; }` の形式で定義します。",
			ExpectedOutput: "7\n30",
			CreatedAt:      now,
			UpdatedAt:      now,
		},
		{
			ID:             10,
			OrderIndex:     10,
			Category:       "文字列",
			Title:          "文字列操作",
			Description:    "PHP の文字列関数を使ってみましょう。\n\n`strlen()`, `strtoupper()`, `str_replace()`, `substr()` を練習します。",
			StarterCode:    "<?php\n$str = \"Hello, PHP World!\";\n\necho strlen($str) . \"\\n\";              // 17\necho strtoupper($str) . \"\\n\";          // HELLO, PHP WORLD!\necho str_replace(\"PHP\", \"Go\", $str) . \"\\n\"; // Hello, Go World!\necho substr($str, 7, 3) . \"\\n\";        // PHP\n",
			HintText:       "`strlen` で長さ、`strtoupper` で大文字変換、`str_replace` で置換、`substr(文字列, 開始, 長さ)` で部分文字列を取得します。",
			ExpectedOutput: "17\nHELLO, PHP WORLD!\nHello, Go World!\nPHP",
			CreatedAt:      now,
			UpdatedAt:      now,
		},
		{
			ID:             11,
			OrderIndex:     11,
			Category:       "OOP",
			Title:          "クラスの基本",
			Description:    "クラスを定義してオブジェクトを作成しましょう。\n\nコンストラクタ、プロパティ、メソッドを実装します。",
			StarterCode:    "<?php\nclass Animal {\n    public $name;\n    public $sound;\n\n    public function __construct($name, $sound) {\n        $this->name  = $name;\n        $this->sound = $sound;\n    }\n\n    public function speak() {\n        echo $this->name . \" は \" . $this->sound . \" と鳴きます\\n\";\n    }\n}\n\n$dog = new Animal(\"犬\", \"ワン\");\n$cat = new Animal(\"猫\", \"ニャン\");\n$dog->speak();\n$cat->speak();\n",
			HintText:       "`class クラス名 { ... }` でクラスを定義。`new クラス名()` でオブジェクトを生成します。",
			ExpectedOutput: "犬 は ワン と鳴きます\n猫 は ニャン と鳴きます",
			CreatedAt:      now,
			UpdatedAt:      now,
		},
		{
			ID:             12,
			OrderIndex:     12,
			Category:       "OOP",
			Title:          "継承",
			Description:    "基底クラスを継承して子クラスを作りましょう。\n\n`extends` キーワードと `parent::` を使います。",
			StarterCode:    "<?php\nclass Shape {\n    protected $color;\n\n    public function __construct($color) {\n        $this->color = $color;\n    }\n\n    public function describe() {\n        echo \"色: \" . $this->color . \"\\n\";\n    }\n}\n\nclass Circle extends Shape {\n    private $radius;\n\n    public function __construct($color, $radius) {\n        parent::__construct($color);\n        $this->radius = $radius;\n    }\n\n    public function describe() {\n        parent::describe();\n        echo \"半径: \" . $this->radius . \"\\n\";\n    }\n}\n\n$c = new Circle(\"赤\", 5);\n$c->describe();\n",
			HintText:       "`class 子クラス extends 親クラス` で継承。`parent::メソッド()` で親のメソッドを呼び出せます。",
			ExpectedOutput: "色: 赤\n半径: 5",
			CreatedAt:      now,
			UpdatedAt:      now,
		},
	}

	result := db.Clauses(clause.OnConflict{DoNothing: true}).Create(&exercises)
	if result.Error != nil {
		return result.Error
	}
	log.Printf("seed: php_exercises %d 件を挿入（既存はスキップ）", result.RowsAffected)
	return nil
}
