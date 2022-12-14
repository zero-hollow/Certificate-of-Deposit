// 集成go-sdk注意事项
0、进入poissonchain-go-sdk目录
1、拷贝lib目录下的openssl文件夹到/usr/local/include/
2、拷贝lib目录下的securec文件夹到/usr/local/include/
3、保证common、proto、thirdparty包和wienerchain-go-sdk包在平级目录下，使go module可以正常引用到相关依赖
4、App集成go sdk，进行go build编译时，增加-a编译选项，保证所有代码可强制重新编译