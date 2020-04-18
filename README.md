# share-routine
ハッカソン用アプリ

![image](https://user-images.githubusercontent.com/11689944/79634706-33fd9600-81a7-11ea-8eb9-6e5a7390770c.png)

## 実行

超雑ですがroutines.go内の定数(Environment)で環境の切り替えをしてます。

### ローカル(Environment=local)

```
$ go build
$ go run share-routine
```

### 本番(Environment=local以外)
GCPのCloudRunにデプロイする場合。
プロジェクト名は任意で設定してください

```
$ gcloud builds submit --tag gcr.io/share-my-routine/routines
$ gcloud run deploy --image gcr.io/share-my-routine/routines --platform managed
```

## インフラまわり
Google Cloud Platformで完結させました。

* デプロイ
  * Cloud Build
  * CloudRun 
* DB
  * Cloud FireStore
* ストレージ
  * Cloud Storege
