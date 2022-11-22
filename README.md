# pon-tf-registory

Custom Terraform registory

```
$ go run main.go
$ ngrok http 8080
# make insert
$ terraform init
```

### docker

```
$ docker run -p 8080:8080 -e PGP_ID=sample test-tf-registory
```

## 参考

https://qiita.com/cappyzawa/items/8be66af768c11c45f414
