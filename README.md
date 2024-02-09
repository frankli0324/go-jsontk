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
tks, err := jsontk.Tokenize(b)
//fmt.Println(f)
if err != nil {
    os.Exit(1)
}
for _, tk := range tks {
    v := jsontk.JSON([]jsontk.Token{tk})
    var err error
    switch tk.Type {
    case jsontk.NUMBER:
        _, err = v.Float64()
    case jsontk.STRING:
        _, err = v.String()
    }
    if err != nil {
        os.Exit(1)
    }
}
```

![image](https://github.com/frankli0324/go-jsontk/assets/20221896/9df5a269-7308-4336-ad85-948d60d0e527)

> the results only shows the difference between standard library and jsontk, the succeeded cases are not shown.

## Warning

EXPERIMENTAL
