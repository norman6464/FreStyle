package domain

// CodeExecutionInput は学習者コードのサンドボックス実行入力。
// Language は "php" / "go" / "bash"。Stdin はテストケース採点で標準入力として流す。
type CodeExecutionInput struct {
	Code     string `json:"code"`
	Language string `json:"language"`
	Stdin    string `json:"stdin"`
}

// CodeExecutionResult はサンドボックス実行の結果。
type CodeExecutionResult struct {
	Stdout   string `json:"stdout"`
	Stderr   string `json:"stderr"`
	ExitCode int    `json:"exitCode"`
}
