package client

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"github.com/boltdb/bolt"
	"os"
	"strconv"
)

/**
 * @author: SuZhiXiaoWei
 * @DateTime: 2022/4/16 18:23
 **/
const (
	DB_PASH  = "./chain.db" //数据库保存地址
	BK_USER  = "user"   //存储用户的桶 key = username value = pswd
	BK_MONEY = "money" //存储余额的桶 key = username value = money
	BK_STATUS = "status" //存储当前是否登入的状态 key = nowuser value= now username
	ST_NOWUSER  = "nowuser" //状态桶的key值
)

/**
用户注册
*/
func Register() {
	res := flag.NewFlagSet("register", flag.ExitOnError)
	name := res.String("name", "", "账户名")
	pswd := res.String("pswd", "", "密码")
	res.Parse(os.Args[2:])

	db, err := bolt.Open(DB_PASH, 0600, nil)
	defer db.Close()
	if err != nil {
		fmt.Println("DB打开失败")
		return
	}
	err = db.Update(func(tx *bolt.Tx) error {
		//通过状态桶判断当前是否有用户登入
		status := tx.Bucket([]byte(BK_STATUS))
		//如果当前用户存在
		if status != nil {
			st := status.Get([]byte(ST_NOWUSER))
			if !bytes.Equal(st, []byte("")) {
				return errors.New("当前为登录状态，不能执行注册操作")
			}
		}
		//当前用户不存在
		user := tx.Bucket([]byte(BK_USER))
		//判断用户信息桶 若不存在，则创建
		if user == nil {
			user, err = tx.CreateBucket([]byte(BK_USER))
			if err != nil {
				return err
			}
		} else {
			//桶存在，判断该用户是否已被注册
			get_user := user.Get([]byte(*name))
			if get_user != nil {
				return errors.New("用户已被注册")
			}
		}
		//用户没用被注册，则注册新用户
		err = user.Put([]byte(*name), []byte(*pswd))
		if err != nil {
			return errors.New("新建账户失败" + err.Error())
		}

		//判断余额桶是否存在 不存在则创建
		bk_money := tx.Bucket([]byte(BK_MONEY))
		if bk_money == nil {
			bk_money, err = tx.CreateBucket([]byte(BK_MONEY))
			if err != nil {
				return err
			}
		}
		//创建一个用户对应的余额桶，并初始余额值为0
		err = bk_money.Put([]byte(*name), []byte("0"))
		if err != nil {
			return err
		}
		return err
	})

	if err != nil {
		fmt.Println("注册失败：" + err.Error())
		return
	}
	fmt.Println("注册成功")
}

/*
	用户登入
*/
func Login() {
	login := flag.NewFlagSet("login", flag.ExitOnError)
	name := login.String("name", "", "用户名")
	pswd := login.String("pswd", "", "密码")
	login.Parse(os.Args[2:])
	db, err := bolt.Open(DB_PASH, 0600, nil)
	defer db.Close()
	if err != nil {
		fmt.Println("DB打开失败")
		return
	}
	err = db.Update(func(tx *bolt.Tx) error {
		//判断状态桶是否存在不存在则创建
		status := tx.Bucket([]byte(BK_STATUS))
		if status == nil {
			status, err = tx.CreateBucket([]byte(BK_STATUS))
			if err != nil {
				return err
			}
		}
		//对当前是否有用户登录进行判断
		get := status.Get([]byte(ST_NOWUSER))
		if !bytes.Equal(get, []byte("")) {
			return errors.New("当前为登录状态，不可重复登录")
		}

		//判断用户桶中是否有用户存在
		user := tx.Bucket([]byte(BK_USER))
		if user == nil {
			return errors.New("用户不存在")
		}
		//若有则通过 用户名获取密码
		password := user.Get([]byte(*name))
		//如果获取为空，则证明该用户不存在
		if password == nil {
			return errors.New("没用该用户信息")
		}
		//将获取到的密码与用户输入密码相比较
		if !bytes.Equal([]byte(*pswd), password) {
			return errors.New("密码错误")
		}

		//将登录的用户名放入状态桶中 使当前状态为登入状态
		err = status.Put([]byte(ST_NOWUSER), []byte(*name))
		if err != nil {
			return err
		}

		return err
	})

	if err != nil {
		fmt.Println("登录失败：" + err.Error())
		return
	}
	fmt.Println("登录成功")
}

/*
	退出当前用户
*/
func Exit() {
	e := flag.NewFlagSet("exit", flag.ExitOnError)
	e.Parse(os.Args[2:])
	db, err := bolt.Open(DB_PASH, 0600, nil)
	defer db.Close()
	if err != nil {
		fmt.Println("DB打开失败")
		return
	}
	err = db.Update(func(tx *bolt.Tx) error {
		//判断当前状态桶是否为登入状态
		status := tx.Bucket([]byte(BK_STATUS))
		if status == nil {
			return errors.New("无用户登录")
		}
		//如果为登入 则重置value值 使其为空状态
		err = status.Put([]byte(ST_NOWUSER), []byte(""))
		if err != nil {
			return err
		}
		return err
	})
	if err != nil {
		fmt.Println("退出失败：" + err.Error())
		return
	}
	fmt.Println("退出成功")
}

/**
取钱
*/
func GetMoney() {
	m := flag.NewFlagSet("money", flag.ExitOnError)
	money := m.Int("money", 0, "取钱")
	m.Parse(os.Args[2:])
	db, err := bolt.Open(DB_PASH, 0600, nil)
	defer db.Close()
	if err != nil {
		fmt.Println("数据打开失败")
		return
	}

	err = db.Update(func(tx *bolt.Tx) error {
		//判断当前登录用户是否登入
		status := tx.Bucket([]byte(BK_STATUS))
		if status == nil {
			return errors.New("无用户登录")
		}
		//若为登入状态 则获取当前登入用户的用户名
		user := status.Get([]byte(ST_NOWUSER))
		//使用余额桶，查询当前登录用户user的余额
		bk_money := tx.Bucket([]byte(BK_MONEY))
		mo := bk_money.Get(user)
		//将查询到的余额与用户输入的余额相减
		yu, _ := strconv.Atoi(string(mo))
		//对用户输入的金额与余额对比，判断是否有这么多钱
		if *money > yu {
			return errors.New("余额不足")
		}
		sumMoney := strconv.FormatInt(int64(yu-*money), 10)
		//将相加后的余额sum存入余额桶中
		err = bk_money.Put(user, []byte(sumMoney))
		if err != nil {
			return err
		}
		return err
	})

	if err != nil {
		fmt.Println("取款失败：" + err.Error())
		return
	}
	fmt.Println("取款成功")

}
/**
存钱
 */
func SaveMoney() {

	m := flag.NewFlagSet("money", flag.ExitOnError)
	money := m.Int("money", 0, "取钱")

	m.Parse(os.Args[2:])
	db, err := bolt.Open(DB_PASH, 0600, nil)
	defer db.Close()
	if err != nil {
		fmt.Println("DB打开失败")
		return
	}

	err = db.Update(func(tx *bolt.Tx) error {
		//判断当前是否为登入状态
		status := tx.Bucket([]byte(BK_STATUS))
		if status == nil {
			return errors.New("无用户登录")
		}
		//获取当前登入用户的用户名
		user := status.Get([]byte(ST_NOWUSER))

		//使用余额桶，查询当前登录用户user的余额
		bk_money := tx.Bucket([]byte(BK_MONEY))
		mo := bk_money.Get(user)
		//将查询到的余额与用户输入的余额相加
		yu, _ := strconv.Atoi(string(mo))
		sumMoney := strconv.FormatInt(int64(*money+yu), 10)

		//将相加后的余额sum存入余额桶中
		err = bk_money.Put(user, []byte(sumMoney))
		if err != nil {
			return err
		}
		return err
	})

	if err != nil {
		fmt.Println("存入失败：" + err.Error())
		return
	}
	fmt.Println("存入成功")

}
/**
查看余额
 */
func SelectMoney() {
	db, err := bolt.Open(DB_PASH, 0600, nil)
	defer db.Close()
	if err != nil {
		fmt.Println("DB打开失败")
		return
	}

	err = db.Update(func(tx *bolt.Tx) error {
		//判断当前状态
		status := tx.Bucket([]byte(BK_STATUS))
		if status == nil {
			return errors.New("无用户登录")
		}
		//获取登入用户的名称
		user := status.Get([]byte(ST_NOWUSER))

		//使用余额桶，查询当前登录用户user的余额
		bk_money := tx.Bucket([]byte(BK_MONEY))
		sumMoney := bk_money.Get(user)
		//将余额输入打印
		fmt.Printf("余额为：%s\n", sumMoney)
		return err
	})

	if err != nil {
		fmt.Println("查询余额失败：" + err.Error())
		return
	}
}
/*
	操作帮助指令
 */
func Help() {
	fmt.Println("register：参数2个 --name --pswd  用于用户注册")
	fmt.Println("login：参数2个 --name --pswd 用于用户登入")
	fmt.Println("exit：参数无 用于当前用户退出")
	fmt.Println("getMoney：参数1个 --money用于取钱")
	fmt.Println("setMoney：参数1个 --money用于存钱")
	fmt.Println("selectMoney：参数无 用于参看余额")
	fmt.Println("help：用于参看系统当前命令的用法")

}

