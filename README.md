


go.mod 添加 

```
module github.com/liangdas/mqant
replace github.com/liangdas/mqant => github.com/leonlau/mqant/v2 v2.1.1
```

这样就不用修改import 文件路径了. 

在项目文件 的go.mod 中也需要 
```
replace github.com/liangdas/mqant => github.com/leonlau/mqant/v2 v2.1.1
```
