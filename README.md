# jsontk

![image](https://github.com/frankli0324/go-jsontk/assets/20221896/37b70d26-f28f-4616-88f0-3a6683610f00)
> from *Top Gear* S29E2

又不是不能用，你就说比标准库快不快吧

## Features

```go
data := []byte(`{"tokenize":{"json":true,"into":["parts",1.1]}}`)

// Tokenize
res, _ := Tokenize(data)
for _, tk := range res.store {
    fmt.Printf("%s->%s\n", tk.Type.String(), string(tk.Value))
}
// Iterate
res, _ := Iterate(data, func(typ TokenType, idx, len int) {
    // emits same result as `Tokenize`, except with "idx" and "len" information
})
// Validate
err := Validate(data)

// Iterator
var iter Iterator // can be reused
iter.Reset(data)
...
```

## Correctness

tested with `https://github.com/nst/JSONTestSuite` using demonstrated code:

* Tokenize API

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

* Iterator API

```go
func walk(iter *jsontk.Iterator) {
    var tk jsontk.Token
    switch iter.Peek() {
    case jsontk.BEGIN_OBJECT:
        iter.NextObject(func(key *jsontk.Token) bool {
            _, ok := key.Unquote()
            if !ok {
                os.Exit(1)
            }
            walk(iter)
            return true
        })
    case jsontk.BEGIN_ARRAY:
        iter.NextArray(func(idx int) bool {
            walk(iter)
            return true
        })
    case jsontk.STRING:
        _, ok := iter.NextToken(&tk).Unquote()
        if !ok {
            os.Exit(1)
        }
    case jsontk.NUMBER:
        _, err := iter.NextToken(&tk).Number().Float64()
        if err != nil {
            os.Exit(1)
        }
    case jsontk.INVALID:
        os.Exit(1)
    default:
        iter.Skip()
    }
}

var iter jsontk.Iterator
iter.Reset(b)
walk(&iter)
if iter.Error != nil {
    os.Exit(1)
}
_, _, l := iter.Next()
if l != 0 {
    os.Exit(1)
}
os.Exit(0)
```

![image](https://github.com/user-attachments/assets/baa460bd-450b-4bcf-a45c-d00f60cf15aa)

> the results only shows the difference between standard library and jsontk, the succeeded cases are not shown.

## Warning / Disclaimer

EXPERIMENTAL

This library is solely designed to extract payload fields in an **insecure** manner.
You might not want to use this library unless you understand what you're doing
