# techstack
- `go/parser` 
- `go/ast`

# purpose

## 1.check package import
`crypto/x509`

## 2.check function are used
`net/http#StripPrefix`

`net/http#SameSiteDefaultMode`

## 3.check context functions parent is not nil
`context#WithValue`

`context#WithDeadline`

`context#WithCancel`