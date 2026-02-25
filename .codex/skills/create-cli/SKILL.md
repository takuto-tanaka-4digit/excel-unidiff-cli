---
name: create-cli
description: >
  コマンドライン UI のパラメータと UX を設計する: 引数、フラグ、サブコマンド、
  ヘルプ文、出力形式、エラーメッセージ、終了コード、プロンプト、
  config/env の優先順位、安全/ドライランの挙動。
  CLI 仕様の設計（実装前）や、既存 CLI の表面設計を
  一貫性/合成可能性/発見容易性のために整理する時に使う。
---

# CLI 作成

CLI の表面（構文 + 振る舞い）を設計。人間優先、スクリプト友好。

## まずこれをやる

- `agent-scripts/skills/create-cli/references/cli-guidelines.md` を読み、既定の評価軸として適用。
- 上流/完全版ガイド: https://clig.dev/（変更提案: https://github.com/cli-guidelines/cli-guidelines）
- インターフェイス確定に必要な最小限の確認質問だけをする。

## すばやく確認

ユーザーが曖昧なら、質問した上で最良の既定で進める:

- コマンド名 + 1 文の目的。
- 主な利用者: 人間/スクリプト/両方。
- 入力元: args vs stdin; ファイル vs URL; シークレット（フラグ禁止）。
- 出力契約: 人間向けテキスト、`--json`、`--plain`、終了コード。
- 対話性: プロンプト許可? `--no-input` 必要? 破壊操作の確認は?
- 設定モデル: flags/env/config-file; 優先順位; XDG vs リポローカル。
- プラットフォーム/ランタイム制約: macOS/Linux/Windows; 単一バイナリ vs ランタイム依存。

## 成果物（出力すべきもの）

CLI 設計時は、実装可能なコンパクト仕様を出す:

- コマンドツリー + USAGE 概要。
- 引数/フラグ表（型、既定値、必須/任意、例）。
- サブコマンドの意味（役割、冪等性、状態変更）。
- 出力ルール: stdout vs stderr; TTY 判定; `--json`/`--plain`; `--quiet`/`--verbose`。
- エラー + 終了コードの対応表（主要な失敗モード）。
- 安全ルール: `--dry-run`、確認、`--force`、`--no-input`。
- 設定/env ルール + 優先順位（flags > env > project config > user config > system）。
- シェル補完（必要なら）: インストール/発見性; 生成コマンド or 同梱スクリプト。
- 例を 5–10 個（よくあるフロー、パイプ/STDIN 例を含む）。

## 既定の慣習（ユーザー指定がない限り）

- `-h/--help` は常にヘルプを表示し、他の引数は無視。
- `--version` は version を stdout に出力。
- 主要データは stdout; 診断/エラーは stderr。
- 機械出力は `--json`、安定した行ベースは `--plain` を検討。
- プロンプトは stdin が TTY の時のみ; `--no-input` で無効化。
- 破壊操作: 対話確認 + 非対話は `--force` または明示 `--confirm=...` が必要。
- `NO_COLOR`, `TERM=dumb` を尊重; `--no-color` を提供。
- Ctrl-C 対応: 速やかに終了; 後始末は最小限; 可能なら crash-only。

## テンプレート（回答に貼る）

### CLI 仕様スケルトン

次のセクションを埋め、不要なものは削除:

1. **Name**: `mycmd`
2. **One-liner**: `...`
3. **USAGE**:
   - `mycmd [global flags] <subcommand> [args]`
4. **Subcommands**:
   - `mycmd init ...`
   - `mycmd run ...`
5. **Global flags**:
   - `-h, --help`
   - `--version`
   - `-q, --quiet` / `-v, --verbose`（定義を明確に）
   - `--json` / `--plain`（必要なら）
6. **I/O contract**:
   - stdout:
   - stderr:
7. **Exit codes**:
   - `0` success
   - `1` generic failure
  - `2` 無効な使い方（解析/検証）
   - （本当に必要な場合のみコマンド固有コードを追加）
8. **Env/config**:
   - env vars:
   - config file path + precedence:
9. **例**:
   - …

## メモ

- 解析ライブラリ（言語別）の推奨は求められた時のみ; それ以外は言語非依存に保つ。
- 「パラメータ設計」の依頼なら実装に踏み込まない。
