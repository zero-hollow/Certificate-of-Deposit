# common
common provides common modules for other projects, currently includes:
* logger
* viper-tool
* cryptomgr

## 集成指导
集成common模块的其它模块，需要通过go.mod引用common模块，引用方式如下：
```
// common和gmssl的路径根据实际情况指定
replace (
	git.huawei.com/poissonsearch/wienerchain/common => ../../../common  
	gmssl => ../../../thirdparty/GmSSL/gmssl
)
```
