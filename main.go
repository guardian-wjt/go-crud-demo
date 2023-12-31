package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"strconv"
	"time"
)

func main() {

	dsn := "root:root123@tcp(127.0.0.1:3306)/crud-list?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			// 解决查表的时候会自动添加复数的问题，例如：user 变成 users
			SingularTable: true, // 单数表名
		},
	})

	fmt.Println(db)
	fmt.Println(err)
	//验证数据库创建成功

	// 1.配置连接池
	sqlDB, err := db.DB()
	// SetMaxIdleConns 设置空闲连接池中连接的最大数量
	sqlDB.SetMaxIdleConns(10)
	// SetMaxOpenConns 设置打开数据库连接的最大数量
	sqlDB.SetMaxOpenConns(100)
	// SetConnMaxLifetime 设置连接可复用的最大时间
	sqlDB.SetConnMaxLifetime(10 * time.Second) // 10秒

	// 2.结构体
	type List struct {
		gorm.Model
		Name    string `gorm:"type:varchar(20); not null" json:"name" binding:"required"`
		State   string `gorm:"type:varchar(20); not null" json:"state" binding:"required"`
		Phone   string `gorm:"type:varchar(20); not null" json:"phone" binding:"required"`
		Email   string `gorm:"type:varchar(40); not null" json:"email" binding:"required"`
		Address string `gorm:"type:varchar(200); not null" json:"address" binding:"required"`
	}
	/*
		注意：1.结构体里面的变量（Name）必须首字符大写
		gorm	指定类型
		json	表示json接收的时候的名称
		binding	required 表示必须传入
	*/

	// 3.迁移
	db.AutoMigrate(&List{})

	// 1.主键没有 （不符合规范）给结构体添加 gorm.Model
	// 2.表名复数问题

	//接口
	r := gin.Default()

	//测试
	//r.GET("/", func(c *gin.Context) {
	//	c.JSON(200, gin.H{
	//		"message": "请求成功了",
	//	})
	//})
	/* 业务代码
	正确：200
	错误：400
	*/

	// restful 编码规范	风格

	//增
	r.POST("/user/add", func(c *gin.Context) {
		var data List

		err := c.ShouldBindJSON(&data) //模型绑定

		// 判断绑定是否错误
		if err != nil {
			c.JSON(200, gin.H{
				"msg":  "添加失败",
				"data": gin.H{},
				"code": 400,
			})
		} else {
			// 进行数据库的操作 —— 创建一条数据
			db.Create(&data) // 通过结构体的指针来创建数据

			c.JSON(200, gin.H{
				"msg":  "添加成功了",
				"data": data,
				"code": 200,
			})

		}

	})

	//删
	// 1.找到id所对应的条目
	// 2.判断id是否存在
	// 3.存在从数据库中删除
	// 4.不存在，返回id没找到
	r.DELETE("/user/delete/:id", func(c *gin.Context) {
		var data []List //定义结构体，可能会查找到多个，用切片类型

		// 接收id
		id := c.Param("id")

		// 判断id是否存在
		db.Where("id = ?", id).Find(&data)

		// id存在从数据库中删除,不存在，返回id没找到
		if len(data) == 0 {
			c.JSON(200, gin.H{
				"msg":  "id没找到，删除失败",
				"code": 400,
			})
		} else {
			//操作数据库删除
			db.Where("id = ?", id).Delete(&data)

			c.JSON(200, gin.H{
				"msg":  "删除成功",
				"code": 200,
			})
		}

	})

	//改
	r.PUT("/user/update/:id", func(c *gin.Context) {
		// 1.找到id所对应的条目
		// 2.判断id是否存在
		// 3.存在修改对应条目
		// 4.不存在，返回id没找到

		var data List

		//接收id
		id := c.Param("id")

		//判断id是否存在
		db.Select("id").Where("id = ?", id).Find(&data)

		// 不存在，返回id没找到	|	存在修改对应条目
		if data.ID == 0 {
			c.JSON(200, gin.H{
				"msg":  "用户id没找到",
				"code": 400,
			})
		} else {
			//操作数据库修改
			err := c.ShouldBindJSON(&data) //模型绑定

			if err != nil {
				c.JSON(200, gin.H{
					"msg":  "修改失败",
					"code": 400,
				})
			} else {
				// db 修改数据库内容
				db.Where("id = ?", id).Updates(&data)
				c.JSON(200, gin.H{
					"msg":  "修改成功",
					"code": 200,
				})
			}
		}

	})

	//查 (条件查询，全部查询 / 分页查询)

	// 条件查询
	r.GET("/user/list/:name", func(c *gin.Context) {

		// 获取路径参数
		name := c.Param("name")
		//fmt.Println(name)
		var dataList []List // 定义结构体		可能会查询到一个或多个

		// 查询数据库
		db.Where("name = ?", name).Find(&dataList)

		// 判断是否查询到数据
		if len(dataList) == 0 {
			c.JSON(200, gin.H{
				"msg":  "没有查询到对应数据",
				"data": gin.H{},
				"code": 400,
			})
		} else {
			c.JSON(200, gin.H{
				"msg":  "查询成功了",
				"data": dataList,
				"code": 200,
			})
		}

	})

	// 全部查询
	r.GET("/user/list", func(c *gin.Context) {

		var dataList []List

		// 1.查询全部数据	或	查询分页数据
		pageNum, _ := strconv.Atoi(c.Query("pageNum"))   // 当前页数
		pageSize, _ := strconv.Atoi(c.Query("pageSize")) //当前页数显示的条数

		// 判断是否需要分页
		if pageNum == 0 || pageSize == 0 {
			pageNum = -1
			pageSize = -1
		}

		offsetVal := (pageNum - 1) * pageSize // 分页的固定写法
		if pageNum == -1 && pageSize == -1 {
			offsetVal = -1
		}

		//返回总数据
		var total int64
		//查询数据库
		db.Model(dataList).Count(&total).Limit(pageSize).Offset(offsetVal).Find(&dataList)

		if len(dataList) == 0 {
			c.JSON(200, gin.H{
				"msg":  "没有查询到对应数据",
				"data": gin.H{},
				"code": 400,
			})
		} else {
			c.JSON(200, gin.H{
				"msg":  "查询成功了",
				"code": 200,
				"data": gin.H{
					"list":     dataList,
					"pageNum":  pageNum,
					"pageSize": pageSize,
					"total":    total,
				},
			})
		}

	})

	//端口号
	PORT := "3000"
	r.Run(":" + PORT)

}
