# sheet-env

Google スプレッドシートに保存した秘匿値を、一時的な環境変数としてコマンドに注入する CLI ツールです。

```bash
sheet-env npm run dev
sheet-env python app.py
sheet-env -- node server.js --port 3000
```

**Go 製のシングルバイナリ** — Node.js / Python などのランタイム不要。バイナリを PATH に置くだけで動きます。

---

## 目次

1. [仕組み](#仕組み)
2. [インストール](#インストール)
3. [セットアップ](#セットアップ)
4. [設定リファレンス](#設定リファレンス)
5. [CLIリファレンス](#cliリファレンス)
6. [配布・リリース方法](#配布リリース方法)
7. [セキュリティ](#セキュリティ)

---

## 仕組み

```
プロジェクト/.env          プロジェクト/.sheetenv
┌──────────────────────┐  ┌──────────────────────────────────┐
│ DB_URL=localhost      │  │ SHEET_ENV_SPREADSHEET_ID=abc123  │
│ API_KEY=$sheet   ─────┼──┼──► Google スプレッドシート      │
│ SECRET=$sheet    ─────┼──┼──► (API_KEY と SECRET の値を取得)│
└──────────────────────┘  └──────────────────────────────────┘
           │
           ▼
  環境変数をマージして子プロセスを起動
  (シークレットはメモリ上のみ・ファイルや履歴に残らない)
```

1. `.env` を読み込み、値が `$sheet` の変数を抽出する
2. Google Sheets API でその変数名に対応する値をスプレッドシートから取得する
3. 取得した値と `.env` のローカル変数をマージし、指定コマンドを子プロセスとして起動する

**優先順位（高 → 低）**

```
スプレッドシートの値  >  .env のローカル値  >  既存の環境変数
```

---

## インストール

### バイナリをダウンロード（推奨）

[GitHub Releases](https://github.com/peko-thunder/sheet-password/releases) からお使いの OS・アーキテクチャに合ったファイルをダウンロードします。

| ファイル名 | 対象環境 |
|---|---|
| `sheet-env-linux-amd64` | Linux x86\_64 |
| `sheet-env-linux-arm64` | Linux ARM64 (Raspberry Pi 等) |
| `sheet-env-darwin-amd64` | macOS Intel |
| `sheet-env-darwin-arm64` | macOS Apple Silicon (M1/M2/M3) |
| `sheet-env-windows-amd64.exe` | Windows x64 |

```bash
# Linux / macOS の例
curl -L https://github.com/peko-thunder/sheet-password/releases/latest/download/sheet-env-linux-amd64 \
  -o /usr/local/bin/sheet-env
chmod +x /usr/local/bin/sheet-env

sheet-env --version
```

```powershell
# Windows PowerShell の例
Invoke-WebRequest -Uri https://github.com/peko-thunder/sheet-password/releases/latest/download/sheet-env-windows-amd64.exe `
  -OutFile sheet-env.exe
# sheet-env.exe を PATH の通ったフォルダに移動する
```

### ソースからビルド

Go 1.22 以上が必要です。

```bash
git clone https://github.com/peko-thunder/sheet-password.git
cd sheet-password

make build     # カレント OS・アーキテクチャ向け → ./sheet-env
make install   # $GOPATH/bin にインストール
```

---

## セットアップ

### 1. Google Cloud でサービスアカウントを作成する（チーム・CI 向け）

個人のローカル利用だけなら **[方法 B](#方法b--application-default-credentials個人利用向け)** のほうが手軽です。

1. [Google Cloud Console](https://console.cloud.google.com/) でプロジェクトを作成（または既存を使用）
2. **「APIとサービス」→「有効なAPIとサービス」** から **Google Sheets API** を有効化
3. **「IAMと管理」→「サービスアカウント」** でサービスアカウントを作成
4. 作成したサービスアカウントの **「キー」→「鍵を追加」→「JSON」** でキーファイルをダウンロード
5. キーファイルを安全な場所に保存（例: `~/.config/sheet-env/credentials.json`）

#### 方法 B — Application Default Credentials（個人利用向け）

Google Cloud SDK が入っていれば、サービスアカウント不要です。

```bash
gcloud auth application-default login
# ブラウザで Google アカウントにログインするだけ
```

### 2. スプレッドシートを用意する

新規スプレッドシートを作成し、**A 列に変数名・B 列に値**を記載します。

| A（変数名） | B（値） |
|---|---|
| SECRET\_API\_KEY | sk-abc123... |
| AWS\_SECRET\_ACCESS\_KEY | AKIA... |
| DATABASE\_PASSWORD | hunter2 |

- ヘッダー行を設ける場合は `.sheetenv` で `SHEET_ENV_HEADER_ROW=true` を設定してください。
- サービスアカウントを使う場合は、スプレッドシートをサービスアカウントのメールアドレスと**閲覧者**として共有してください。

スプレッドシート URL の `spreadsheets/d/` 以降のランダムな文字列が **Spreadsheet ID** です。

```
https://docs.google.com/spreadsheets/d/<ここが Spreadsheet ID>/edit
```

### 3. `.sheetenv` をプロジェクトルートに作成する

```dotenv
# .sheetenv
SHEET_ENV_SPREADSHEET_ID=1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgVE2upms
SHEET_ENV_SHEET_NAME=Sheet1
SHEET_ENV_CREDENTIALS=/path/to/credentials.json   # サービスアカウントの場合のみ
```

> **重要**: `.sheetenv` は `.gitignore` に追加してリポジトリにコミットしないでください。

### 4. プロジェクトの `.env` を編集する

スプレッドシートから取得したい変数の値を `$sheet` に書き換えます。

```dotenv
# .env
DATABASE_URL=postgres://localhost/mydb   # ローカル値 — そのまま使われる
SECRET_API_KEY=$sheet                    # スプレッドシートから取得
AWS_SECRET_ACCESS_KEY=$sheet             # スプレッドシートから取得
```

### 5. コマンドをラップして実行する

```bash
sheet-env npm run dev
sheet-env python app.py
sheet-env go run .
sheet-env -- node server.js --port 3000
sheet-env -e .env.staging -- npm start
```

実行時に `$sheet` の値をスプレッドシートから取得し、子プロセスの環境変数として注入します。

---

## 設定リファレンス

### `.sheetenv` ファイル（ツール設定）

プロジェクトルートに置きます。シェルの環境変数が既に設定されている場合は、そちらが優先されます。

| 変数名 | 必須 | デフォルト | 説明 |
|---|---|---|---|
| `SHEET_ENV_SPREADSHEET_ID` | ✅ | — | スプレッドシートの ID |
| `SHEET_ENV_SHEET_NAME` | | `Sheet1` | シート名 |
| `SHEET_ENV_CREDENTIALS` | | — | サービスアカウント JSON のパス。未設定時は ADC を使用 |
| `SHEET_ENV_KEY_COLUMN` | | `A` | 変数名が入っている列 |
| `SHEET_ENV_VALUE_COLUMN` | | `B` | 値が入っている列 |
| `SHEET_ENV_HEADER_ROW` | | `false` | `true` にすると 1 行目をヘッダーとしてスキップ |

`GOOGLE_APPLICATION_CREDENTIALS`（Google 標準の環境変数）も `SHEET_ENV_CREDENTIALS` の代替として認識します。

### `.env` ファイル（プロジェクト設定）

| 値 | 動作 |
|---|---|
| 通常の文字列 | そのまま環境変数として渡す |
| `$sheet` | 実行時にスプレッドシートから取得して注入する |

`.env` の書式は標準的な dotenv 形式に対応しています。

```dotenv
# コメント
PLAIN_VALUE=hello world
QUOTED_VALUE="spaces are fine"
EXPORT_STYLE=export KEY=value      # export プレフィックスも認識
SECRET=$sheet
```

---

## CLI リファレンス

```
sheet-env [options] [--] <command> [args...]
```

| オプション | 省略形 | デフォルト | 説明 |
|---|---|---|---|
| `--env-file <path>` | `-e` | `.env` | 読み込む .env ファイルのパス |
| `--config <path>` | `-c` | `.sheetenv` | ツール設定ファイルのパス |
| `--verbose` | | | デバッグ情報を stderr に出力 |
| `--help` | `-h` | | ヘルプを表示 |
| `--version` | `-v` | | バージョンを表示 |

`--` を使うと、その後ろのすべてのトークンをコマンドとして扱います（オプションの競合を避けたいときに使用）。

```bash
# --port がコマンド側の引数だと明示したい場合
sheet-env -- node server.js --port 3000

# 複数の環境向けに .env を切り替える
sheet-env -e .env.staging npm run build
```

---

## 配布・リリース方法

### ローカルで全プラットフォーム向けバイナリをビルドする

```bash
make build-all
```

`dist/` ディレクトリに以下が生成されます。

```
dist/
├── sheet-env-linux-amd64
├── sheet-env-linux-arm64
├── sheet-env-darwin-amd64
├── sheet-env-darwin-arm64
└── sheet-env-windows-amd64.exe
```

### リリースアーカイブを作成する

```bash
make release
```

`dist/` にバイナリと圧縮アーカイブが生成されます。

```
dist/
├── sheet-env-v1.2.0-linux-amd64.tar.gz
├── sheet-env-v1.2.0-linux-arm64.tar.gz
├── sheet-env-v1.2.0-darwin-amd64.tar.gz
├── sheet-env-v1.2.0-darwin-arm64.tar.gz
└── sheet-env-v1.2.0-windows-amd64.zip
```

バージョンは `git describe --tags` から自動取得されます。事前に `git tag v1.2.0` でタグを打ってください。

### GitHub Releases にアップロードする

```bash
# タグを打つ
git tag v1.2.0
git push origin v1.2.0

# アーカイブを生成
make release

# gh CLI でリリースを作成してアーカイブをアップロード
gh release create v1.2.0 dist/*.tar.gz dist/*.zip \
  --title "v1.2.0" \
  --notes "リリースノートをここに書く"
```

### Makefile コマンド一覧

| コマンド | 説明 |
|---|---|
| `make build` | カレント OS・アーキテクチャ向けにビルド → `./sheet-env` |
| `make install` | `$GOPATH/bin` にインストール |
| `make build-all` | 全プラットフォーム向けにビルド → `dist/` |
| `make release` | `build-all` + 圧縮アーカイブを生成 |
| `make clean` | バイナリと `dist/` を削除 |
| `make test` | テストを実行 |
| `make vet` | `go vet` を実行 |

---

## セキュリティ

### やること

- `.sheetenv` と `credentials.json` は必ず `.gitignore` に追加する
- スプレッドシートの共有設定は「リンクを知っている全員」ではなく、サービスアカウントや特定ユーザーのみに限定する
- サービスアカウントには閲覧権限のみ付与する（書き込み不要）

### やらないこと

- `credentials.json` を `~/.bashrc` や `~/.zshrc` に直書きしない
- スプレッドシートに秘匿値と公開値を混在させない（別シートに分けることを推奨）

### ツールの安全性

- このツールはスプレッドシートを **読み取り専用** (`spreadsheets.readonly`) スコープでしか参照しません
- 取得したシークレットは子プロセスのメモリにのみ存在し、ファイルや shell の履歴には書き込まれません
- 子プロセスの起動は `shell: false`（シェルを介さない直接 exec）のため、シークレットがコマンド履歴に残りません
