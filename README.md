# jsontk

![image](https://github.com/frankli0324/go-jsontk/assets/20221896/37b70d26-f28f-4616-88f0-3a6683610f00)
> from *Top Gear* S29E2

又不是不能用，你就说比标准库快不快吧

## features

```jsonc
{
	"with" :  	// comments!
     true,
 "allows": 1, // a bit of error
}
```

## compatibility

tested with `https://github.com/nst/JSONTestSuite` using this code:

```go
func walk(tks *jsontk.JSON) {
    var err error
    switch tks.Type() {
    case jsontk.BEGIN_OBJECT:
        for _, k := range tks.Keys() {
            walk(tks.Get(k))
        }
    case jsontk.BEGIN_ARRAY:
        for i := tks.Len(); i > 0; i-- {
            walk(tks.Index(i - 1))
        }
    case jsontk.NUMBER:
        _, err = tks.Float64()
    case jsontk.STRING:
        _, err = tks.String()
    }
    if err != nil {
        os.Exit(1)
    }
}

tks, err := jsontk.Tokenize(b)
//fmt.Println(f)
if err != nil {
    os.Exit(1)
}
walk(tks)
```

![image](https://github.com/frankli0324/go-jsontk/assets/20221896/1f504938-1994-4cd9-aa5d-fcb162659a52)

> the results only shows the difference between standard library and jsontk, the succeeded cases are not shown.

## Warning / Disclaimer

EXPERIMENTAL

This library is solely designed to extract payload fields in an **insecure** manner.
You might not want to use this library unless you understand what you're doing
