# sheet-env

Google スプレッドシートに保存した秘匿値を、一時的な環境変数としてコマンドに注入するCLIツールです。

```bash
sheet-env npm run dev
sheet-env python app.py
sheet-env -- node server.js --port 3000
```

**Go 製のシングルバイナリ** — ランタイム不要、インストールはバイナリを PATH に置くだけです。

---

## インストール

### バイナリをダウンロード（推奨）

[GitHub Releases](https://github.com/peko-thunder/sheet-password/releases) から
お使いの OS / アーキテクチャに合ったバイナリをダウンロードして PATH に配置します。

```bash
# 例: Linux amd64
curl -L https://github.com/peko-thunder/sheet-password/releases/latest/download/sheet-env-linux-amd64 \
  -o /usr/local/bin/sheet-env
chmod +x /usr/local/bin/sheet-env
```

### ソースからビルド

Go 1.22 以上が必要です。

```bash
git clone https://github.com/peko-thunder/sheet-password.git
cd sheet-password
make build          # カレントOS向け → ./sheet-env
make install        # $GOPATH/bin にインストール
make build-all      # 全プラットフォーム向け → dist/
make release        # dist/ に tar.gz / zip を生成
```

---

## セットアップ

### 1. スプレッドシートを用意する

| A（変数名）           | B（値）           |
|-----------------------|-------------------|
| SECRET_API_KEY        | sk-abc123...      |
| AWS_SECRET_ACCESS_KEY | AKIA...           |
| DATABASE_PASSWORD     | hunter2           |

### 2. 認証を設定する

**方法A — サービスアカウント（チーム・CI向け）**

1. Google Cloud Console でサービスアカウントを作成し、JSON キーをダウンロード
2. スプレッドシートをそのサービスアカウントのメールアドレスと共有（閲覧権限）
3. `.sheetenv` に `SHEET_ENV_CREDENTIALS` でパスを指定

**方法B — Application Default Credentials（個人利用向け）**

```bash
gcloud auth application-default login
```

`SHEET_ENV_CREDENTIALS` の設定は不要です。

### 3. プロジェクトに `.sheetenv` を作成する

```dotenv
# .sheetenv  ← プロジェクトルートに置く（.gitignore に追加すること）
SHEET_ENV_SPREADSHEET_ID=1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgVE2upms
SHEET_ENV_SHEET_NAME=Sheet1
SHEET_ENV_CREDENTIALS=/path/to/credentials.json   # 方法Aの場合のみ

# オプション
# SHEET_ENV_KEY_COLUMN=A       # 変数名の列  (デフォルト: A)
# SHEET_ENV_VALUE_COLUMN=B     # 値の列      (デフォルト: B)
# SHEET_ENV_HEADER_ROW=false   # 1行目をヘッダーとしてスキップ (デフォルト: false)
```

> **注意**: `.sheetenv` には Spreadsheet ID が含まれるため、`.gitignore` に追加してください。

### 4. プロジェクトの `.env` を編集する

スプレッドシートから取得したい変数の値を `$sheet` にします。

```dotenv
# .env
DATABASE_URL=postgres://localhost/mydb   # 通常の値 — そのまま使われる
SECRET_API_KEY=$sheet                    # スプレッドシートから取得
AWS_SECRET_ACCESS_KEY=$sheet             # スプレッドシートから取得
```

### 5. コマンドをラップして実行する

```bash
sheet-env npm run dev
sheet-env python app.py
sheet-env -- node server.js --port 3000
sheet-env -e .env.production -- npm start
```

実行時に `$sheet` の変数をスプレッドシートから取得し、プロセスの環境変数として注入します。  
シークレットはプロセスのメモリ上にのみ存在し、ファイルや shell の履歴には残りません。

---

## オプション

```
sheet-env [options] [--] <command> [args...]

  -e, --env-file <path>   .env ファイルのパス (デフォルト: .env)
  -c, --config <path>     設定ファイルのパス  (デフォルト: .sheetenv)
      --verbose           デバッグ情報を stderr に出力
  -h, --help              ヘルプを表示
  -v, --version           バージョンを表示
```

---

## ビルド成果物

| ファイル                          | 対象              |
|-----------------------------------|-------------------|
| `sheet-env-linux-amd64`           | Linux x86_64      |
| `sheet-env-linux-arm64`           | Linux ARM64       |
| `sheet-env-darwin-amd64`          | macOS Intel       |
| `sheet-env-darwin-arm64`          | macOS Apple Silicon |
| `sheet-env-windows-amd64.exe`     | Windows x64       |

---

## セキュリティ上の注意

- `.sheetenv` と `credentials.json` は `.gitignore` に追加し、リポジトリにコミットしないでください。
- スプレッドシートへのアクセス権限は閲覧のみ（`spreadsheets.readonly`）に絞ってあります。
- スプレッドシートの共有設定は「リンクを知っている全員」ではなく、特定のサービスアカウントや個人に限定してください。
