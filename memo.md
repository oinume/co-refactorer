# Design

## 1. Ask to ChatGPT

prompt: このPRを参考にして、以下のファイルをリファクタリングしてください。

## 2. Function calling

pr: https://github.com/hoge/fuga/pull/12345
files:
  - x/a.go
  - y/b.go


## 3. Get pull request info and file content

- Get pull request info via GitHub API
  - title
  - description
  - diff
- Get file content
  - path
  - content

## 4. Make prompt and request to ChatGPT

prompt

あなたはGo言語のエキスパートです。以下の差分を参考にして、以下の x/a.go, y/b.go を同じように書き換えて、書き換えた結果であるファイルの内容を出力してほしいです。

差分
```
<PRの差分を入れる>
```

x/a.go
```
<ファイルの中身を入れる>
```

y/b.go
```
<ファイルの中身を入れる>
```

## 5. 理想は書き換えられたファイルをローカルであれば置換したい

