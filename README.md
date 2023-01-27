# protoc-gen-go-errors
用于生成在grpc调用时，需要被处理的错误的创建和判断方法；为每一个错误携带 `http status code` 以及 `biz code(业务错误码)`.

## 介绍
通过如下定义
```proto
enum TestErrorReason {
    option (errors.settings) = {
        default_http_code: 500
        start_biz_code: 100001
    };

    TestNotFound = 0 [(errors.code) = {http_code:404}];
    TestBusy = 1;
    TestIncrease = 2 [(errors.code) = {http_code:502 biz_code:100010}];
    TestRedirect = 3 [(errors.code) = {http_code:302}];
}
```

可以生成如下go文件

```go

var bizErrorCodeMap map[string]int = map[string]int{

	"errors.test_TestErrorReason_TestNotFound": 100001,
	"errors.test_TestErrorReason_TestBusy":     100002,
	"errors.test_TestErrorReason_TestIncrease": 100010,
	"errors.test_TestErrorReason_TestRedirect": 100011,
}

func IsTestnotfound(err error) bool {
	if err == nil {
		return false
	}
	e := errors.FromError(err)
	return e.Reason == "errors.test_TestErrorReason_TestNotFound" && e.Code == 404
}

func ErrorTestnotfound(format string, args ...interface{}) *errors.Error {
	return errors.New(404, "errors.test_TestErrorReason_TestNotFound", fmt.Sprintf(format, args...))
}

// ...

func BizErrorCode(err error) int {
	if err == nil {
		return 0
	}
	e := errors.FromError(err)
	return bizErrorCodeMap[e.Reason]
}

```

## 编译及插件使用注意事项
protoc命令行使用插件时，参数名称和插件应用程序的名称保持对应关系。例如：
>   protoc -I . -I ..\gerr --go_out=paths=source_relative:. --go-errors_out=paths=source_relative:. test.proto

--go-errors_out 参数，对应的插件应用程序名称应该是: protoc-gen-go-errors

因此，在go build生成插件应用程序时，需要指定应用程序的名称为：
> go build -o protoc-gen-go-errors

同时，插件应用应位于GOPATH/bin目录，以使protoc可以找到对应的应用程序

## 使用案例

1. build the binary util
```shell
# 下载依赖
make init

# 安装可执行文件
make build
```
2. use it

```shell
make test
```