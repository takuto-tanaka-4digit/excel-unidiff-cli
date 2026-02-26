---
name: euni-commands
description: euniコマンドの使い方。ローカル運用、CI組み込み、異常時切り分け、終了コード/UGコード解釈を最短で引くための実行手順。
---

# euni Commands

`excel-unidiff-cli` の CLI (`euni`) 実行手順。  
目的: 「何をいつ叩くか」を即決できる状態にする。

## 先に把握

- `euni` はサブコマンド必須。`help` サブコマンドなし。
- `--format` は `text|json` のみ。
- `--recursive` 既定 `false`（rootのみ）。
- 書き込みは `apply` と `init-policy` だけ。
- 終了コード:
  - `0`: 問題なし
  - `1`: findings あり（運用上NG）
  - `2`: 実行エラー

## コマンド早見

- `euni init-policy --repo .`
  - `.euni.yml` を作成
- `euni check --repo . --policy ./.euni.yml`
  - 非破壊チェック（追跡済みファイル）
- `euni apply --repo . --policy ./.euni.yml [--dry-run]`
  - policyとの差分を `git config --local` へ適用
- `euni doctor --repo . --policy ./.euni.yml`
  - `check + scan` 統合診断
- `euni scan --repo .`
  - Unicodeパス分析（追跡 + 未追跡）
- `euni version`
  - `euni <version> (<commit>)`

## 実行フロー（推奨）

1. 初回導入
```bash
euni init-policy --repo .
euni check --repo . --recursive --policy ./.euni.yml
```

2. 修復前確認
```bash
euni apply --repo . --recursive --policy ./.euni.yml --dry-run
```

3. 修復実行
```bash
euni apply --repo . --recursive --policy ./.euni.yml
euni check --repo . --recursive --policy ./.euni.yml
```

4. 詰まり時
```bash
euni doctor --repo . --recursive --policy ./.euni.yml
euni scan --repo . --recursive
```

## CI 定型

JSONを stdout に1オブジェクトで出す。

```bash
euni check --repo . --recursive --policy ./.euni.yml --non-interactive --format json > euni-report.json
```

- `schema/euni-report.schema.json` で検証
- プロセス終了コードと `report.exitCode` の一致を検証
- artifact として `euni-report.json` を保存

## オプション適用範囲

- `check|apply|doctor|scan`:
  - `--repo --recursive --format --quiet --non-interactive --log-file`
- `check|apply|doctor` のみ:
  - `--policy`
- `apply` のみ:
  - `--dry-run`
- `init-policy` のみ:
  - `--force`
- `version`:
  - オプション不可

## UGコード最小運用メモ

- findings系（exit `1`）:
  - `UG004` drift
  - `UG005` NFC衝突
  - `UG011` 結合文字
  - `UG012` 非標準FS要因
  - `UG013` policy path曖昧
- error系（exit `2`）:
  - `UG001` repo不正
  - `UG002` git実行失敗
  - `UG003` policy読込失敗
  - `UG006` submodule未初期化
  - `UG007` policy構造不正
  - `UG008` 既存policy上書き拒否
  - `UG009` 非対応コマンド/オプション
  - `UG010` repo境界外gitdir

## 典型ミス

- `euni help` を叩く -> `UG009`
- `scan` に `--policy` を付ける -> `UG009`
- `version` にオプションを付ける -> `UG009`
- `--recursive` なしで submodule 問題を見逃す
