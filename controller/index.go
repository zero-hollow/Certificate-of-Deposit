package controller

// import "C"
import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	"git.huawei.com/goclient/api"
	"git.huawei.com/goclient/response"
	"git.huawei.com/goclient/utils"
	"github.com/deatil/go-cryptobin/cryptobin/crypto"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gomodule/redigo/redis"
	"github.com/jmoiron/sqlx"
)

// 存证结构
type Deposit struct {
	TxHash    string `db:"txhash"`
	Phone     string `db:"phone"`
	TimeStamp int64  `db:"timestamp"`
	IsModify  bool   `db:"ismodify"`
	DataKey   string `db:"datakey"`
}

// 加密密钥
const Key = "dfertf12dfertf12"

// 缓存连接
var Conn, _ = redis.Dial("tcp", "localhost:6379")

//本地数据库连接
//var Db, err = sqlx.Open("mysql", "root:123456@tcp(127.0.0.1:3306)/credite")

// 协程通道
var ResponseChannel = make(chan *response.Response, 15)

// 协程安全锁
var lock sync.Mutex

// 判断手机号有效性
func verifyMobileFormat(mobileNum string) bool {
	regular := "^1[345789]{1}\\d{9}$"
	reg := regexp.MustCompile(regular)
	return reg.MatchString(mobileNum)
}

// 判断哈希有效性
func VerifyHashFormat(HashNum string) bool {
	hash := strings.TrimSpace(HashNum)
	reg := regexp.MustCompile(`[\W|_]{1,}`)
	if len(reg.ReplaceAllString(hash, "")) == 64 {
		return true
	} else {
		return false
	}
}

// 数据上链存储
func UpChain(c *gin.Context) *response.Response {
	Db, err := sqlx.Open("mysql", "root:123456@tcp(127.0.0.1:3306)/credite")
	if err != nil {
		fmt.Println(err)
	}
	//并发操作只能对副本进行
	cCp := c.Copy()
	fmt.Println(cCp)
	//校验数据有效性
	phone := cCp.Query("phone")
	data := cCp.Query("data")
	if verifyMobileFormat(phone) && json.Valid([]byte(data)) {
		go func() {
			//数据加密
			cyptdata := crypto.FromString(data).SetKey(Key).Aes().ECB().PKCS7Padding().Encrypt().ToBase64String()

			//生成datakey，即手机号+时间戳的hash
			timestamp := time.Now().Unix()
			timestampbyte := make([]byte, 8)
			binary.BigEndian.PutUint64(timestampbyte, uint64(timestamp))
			headers := bytes.Join([][]byte{[]byte(phone), timestampbyte}, []byte{})
			datakeySha := sha256.Sum256(headers)
			datakey := hex.EncodeToString(datakeySha[:])


			lock.Lock()
			//上链存储
			//TxHashID, err := api.SaveRecode(phone, strconv.FormatInt(timestamp, 10), cyptdata)
			TxHashID, err := api.SaveRecode(datakey,cyptdata)
			//本地存储
			//Db.Exec("insert into deposit(txhash,phone,timestamp,datakey,ismodify)values(?,?,?,?,?)", TxHashID, phone, timestamp, datakey, false)
			Db.Exec("insert into deposit(txhash,phone,timestamp,datakey,ismodify)values(?,?,?,?,?)", TxHashID, phone, timestamp, datakey, false)
			//删除对应缓存
			Conn.Do("DEL", phone)
			Conn.Do("DEL", TxHashID)
			lock.Unlock()

			if err != nil {
				ResponseChannel <- response.Resp().Json(gin.H{"status": utils.GetCode(err), "data": "", "msg": utils.GetMsg(err)})
			} else {
				ResponseChannel <- response.Resp().Json(gin.H{"status": 200, "data": TxHashID, "msg": "上链成功"})
			}
		}()
		return <-ResponseChannel
	} else {
		return response.Resp().Json(gin.H{"status": 601, "data": "", "msg": "参数无效"})
	}
}

// 根据手机号查询
func QueryByPhone(c *gin.Context) *response.Response {
	Db, _ := sqlx.Open("mysql", "root:123456@tcp(127.0.0.1:3306)/credite")
	defer Db.Close()
	phone := c.Query("phone")
	fmt.Println(phone)
	//先在缓存中查找数据
	r, _ := redis.String(Conn.Do("Get", phone))
	if r == "" {
		if verifyMobileFormat(phone) {
			go func() {
				//先在本地数据库中找到对应交易哈希
				var deposit []Deposit
				hashlist := make([]string, 0, 3)
				Db.Select(&deposit, "select txhash,phone,timestamp,ismodify from deposit where phone=? AND ismodify=? order by timestamp Desc LIMIT 0,3", phone, false)
				for i := 0; i < len(deposit); i++ {
					hashlist = append(hashlist, deposit[i].TxHash)
				}
				result, err := api.QueryByPhone(hashlist)
				if err != nil {
					ResponseChannel <- response.Resp().Json(gin.H{"status": utils.GetCode(err), "data": "", "msg": utils.GetMsg(err)})
				} else {
					//数据解密
					//crypderesult := crypto.FromBase64String(result).SetKey(Key).Aes().ECB().PKCS7Padding().Decrypt().ToString()
					//数据暂存至缓存，过期时间为1小时
					lock.Lock()
					_, err = Conn.Do("Set", phone, result)
					if err != nil {
						fmt.Println("set phone result failed")
					}
					_, err = Conn.Do("expire", phone, 3600)
					if err != nil {
						fmt.Println("expore phone failed")
					}
					lock.Unlock()
					var result_message []api.Message
					err = json.Unmarshal([]byte(result), &result_message)
					if err != nil {
						fmt.Println("json转换错误")
					}
					ResponseChannel <- response.Resp().Json(gin.H{"status": 200, "data": result_message, "msg": "查询成功"})
				}
			}()
			return <-ResponseChannel
		} else {
			return response.Resp().Json(gin.H{"status": 601, "data": "", "msg": "参数无效"})
		}
	} else {
		return response.Resp().Json(gin.H{"status": 200, "data": r, "msg": "查询成功"})
	}
}

// 根据交易哈希查询
func QueryByHash(c *gin.Context) *response.Response {
	Db, _ := sqlx.Open("mysql", "root:123456@tcp(127.0.0.1:3306)/credite")
	defer Db.Close()
	hash := c.Query("hash")
	//先在缓存中查找数据
	r, _ := redis.String(Conn.Do("Get", hash))
	if r == "" {
		if VerifyHashFormat(hash) {
			go func() {
				var deposit []Deposit
				//先在本地数据库中寻找
				Db.Select(&deposit, "select txhash,phone,timestamp,ismodify from deposit where txhash=?", hash)
				//判断deposit中是否有值
				if len(deposit) == 0 {
					ResponseChannel <- response.Resp().Json(gin.H{"status": 603, "data": "", "msg": "交易哈希不存在"})
				} else {
					if deposit[0].IsModify {
						ResponseChannel <- response.Resp().Json(gin.H{"status": 601, "data": "", "msg": "该数据已被修改，请使用新哈希查询"})
					} else {
						//结果解析，判断没有查询结果的情况
						if VerifyHashFormat(hash) {
							result, err := api.QueryByHash(hash)
							if err != nil {
								ResponseChannel <- response.Resp().Json(gin.H{"status": utils.GetCode(err), "data": "", "msg": utils.GetMsg(err)})
							} else {
								//数据解密
								//crypderesult := crypto.FromBase64String(result).SetKey(Key).Aes().ECB().PKCS7Padding().Decrypt().ToString()
								//数据暂存至缓存，过期时间为1小时
								lock.Lock()
								_, err = Conn.Do("Set", hash, result)
								if err != nil {
									fmt.Println("set hash result failed")
								}
								_, err = Conn.Do("expire", hash, 3600)
								if err != nil {
									fmt.Println("expire hash 3600 failed")
								}
								lock.Unlock()
								// fmt.Println(JSON.parse(result))
								var result_message api.Message
								err = json.Unmarshal([]byte(result), &result_message)
								if err != nil {
									fmt.Println("json字符串转为对象错误!")
								}
								ResponseChannel <- response.Resp().Json(gin.H{"status": 200, "data": result_message, "msg": "查询成功"})
							}
						} else {
							ResponseChannel <- response.Resp().Json(gin.H{"status": 601, "data": "", "msg": "参数无效"})
						}
					}
				}

			}()
			return <-ResponseChannel
		} else {
			return response.Resp().Json(gin.H{"status": 601, "data": "", "msg": "参数无效"})
		}

	} else {
		return response.Resp().Json(gin.H{"status": 200, "data": r, "msg": "查询成功"})
	}
}

// 修改上链数据
func Modify(c *gin.Context) *response.Response {
	Db, _ := sqlx.Open("mysql", "root:123456@tcp(127.0.0.1:3306)/credite")
	defer Db.Close()
	//并发操作只能对副本进行
	cCp := c.Copy()
	//校验数据有效性
	phone := cCp.Query("phone")
	hash := cCp.Query("hash")
	data := cCp.Query("data")
	if verifyMobileFormat(phone) && VerifyHashFormat(hash) && json.Valid([]byte(data)) {
		go func() {
			//数据加密
			cyptdata := crypto.FromString(data).SetKey(Key).Aes().ECB().PKCS7Padding().Encrypt().ToBase64String()
			//timestamp := time.Now().Unix()

			lock.Lock()
			//先验证手机号和交易hash对应的东西是否存在
			var deposit []Deposit
			Db.Select(&deposit, "select txhash,phone,timestamp,ismodify from deposit where txhash=? and phone=?", hash, phone)
			fmt.Println("---------------------------")
			fmt.Println(deposit)
			if len(deposit) == 0 {
				ResponseChannel <- response.Resp().Json(gin.H{"status": 603, "data": "", "msg": "该条信息不存在"})
			} else {
				//数据修改
				//TxHashID, err := api.ChangeRecode(phone, strconv.FormatInt(timestamp, 10), cyptdata)

				//生成datakey，即手机号+时间戳的hash
				timestamp := time.Now().Unix()
				timestampbyte := make([]byte, 8)
				binary.BigEndian.PutUint64(timestampbyte, uint64(timestamp))
				headers := bytes.Join([][]byte{[]byte(phone), timestampbyte}, []byte{})
				datakeysha := sha256.Sum256(headers)
				datakey := hex.EncodeToString(datakeysha[:])
			

				TxHashID, err := api.ChangeRecode(datakey, cyptdata)
				//本地数据更新
				//Db.Exec("insert into deposit(txhash,phone,timestamp,ismodify)values(?,?,?,?)", TxHashID, phone, timestamp, false)
				Db.Exec("insert into deposit(txhash,phone,timestamp,datakey,ismodify)values(?,?,?,?,?)", TxHashID, phone, timestamp, datakey, false)
				Db.Exec("update deposit set ismodify=? where txhash=?", true, hash)
				//删除对应缓存
				Conn.Do("DEL", phone)
				Conn.Do("DEL", hash)
				lock.Unlock()

				if err != nil {
					ResponseChannel <- response.Resp().Json(gin.H{"status": utils.GetCode(err), "data": "", "msg": utils.GetMsg(err)})
				} else {
					ResponseChannel <- response.Resp().Json(gin.H{"status": 200, "data": TxHashID, "msg": "修改成功"})
				}
			}
		}()
		return <-ResponseChannel
	} else {
		return response.Resp().Json(gin.H{"status": 601, "data": "", "msg": "参数无效"})
	}
}

// 404页面
func NoRoute(c *gin.Context) {
	c.String(http.StatusNotFound, "404 not found")
}
