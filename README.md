# sheet-env

Google スプレッドシートに保存した秘匿値を、一時的な環境変数としてコマンドに注入するCLIツールです。

```
sheet-env npm run dev
sheet-env python app.py
sheet-env -- node server.js --port 3000
```

秘密情報は `.env` や shell の環境変数に直接書かず、スプレッドシートで一元管理できます。

---

## セットアップ

### 1. インストール

```bash
npm install -g sheet-env
# または npx で都度実行
npx sheet-env npm run dev
```

### 2. スプレッドシートを用意する

スプレッドシートに以下のレイアウトで秘匿値を記載します（ヘッダー行は任意）。

| A（変数名）           | B（値）           |
|-----------------------|-------------------|
| SECRET_API_KEY        | sk-abc123...      |
| AWS_SECRET_ACCESS_KEY | AKIA...           |
| DATABASE_PASSWORD     | hunter2           |

### 3. 認証を設定する

**方法① — サービスアカウント（チーム・CI向け）**

1. Google Cloud Console でサービスアカウントを作成し、JSON キーをダウンロード
2. スプレッドシートをそのサービスアカウントのメールアドレスと共有（閲覧権限）
3. `.sheetenv` に `SHEET_ENV_CREDENTIALS` でパスを指定

**方法② — Application Default Credentials（個人利用向け、簡単）**

```bash
gcloud auth application-default login
```

`SHEET_ENV_CREDENTIALS` の設定は不要です。

### 4. プロジェクトに `.sheetenv` を作成する

```dotenv
# .sheetenv  ← プロジェクトルートに置く（.gitignore に追加すること）
SHEET_ENV_SPREADSHEET_ID=1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgVE2upms
SHEET_ENV_SHEET_NAME=Sheet1
SHEET_ENV_CREDENTIALS=/path/to/credentials.json   # 方法①の場合のみ

# オプション
# SHEET_ENV_KEY_COLUMN=A       # 変数名の列  (デフォルト: A)
# SHEET_ENV_VALUE_COLUMN=B     # 値の列      (デフォルト: B)
# SHEET_ENV_HEADER_ROW=false   # 1行目をヘッダーとしてスキップ (デフォルト: false)
```

> **注意**: `.sheetenv` には Spreadsheet ID が含まれるため、`.gitignore` に追加してください。

### 5. プロジェクトの `.env` を編集する

スプレッドシートから取得したい変数の値を `$sheet` にします。

```dotenv
# .env
DATABASE_URL=postgres://localhost/mydb   # 通常の値 — そのまま使われる
SECRET_API_KEY=$sheet                    # スプレッドシートから取得
AWS_SECRET_ACCESS_KEY=$sheet             # スプレッドシートから取得
```

### 6. コマンドをラップして実行する

```bash
sheet-env npm run dev
sheet-env python app.py
sheet-env -- node server.js --port 3000
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

## セキュリティ上の注意

- `.sheetenv` と `credentials.json` は `.gitignore` に追加し、リポジトリにコミットしないでください。
- スプレッドシートへのアクセス権限は最小限（閲覧のみ）に絞ってください。
- スプレッドシートの共有設定は「リンクを知っている全員」ではなく、サービスアカウントや特定ユーザーに限定してください。

---

## ビルド（開発者向け）

```bash
npm install
npm run build   # dist/ に出力
```
