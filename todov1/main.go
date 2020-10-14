package main

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

/**
 * 项目说明：
 * 	基于Golang Gin 进行 Rest API 开发
 * 1、定义Todo model entity
 * 2、配置MySQL数据库链接
 * 3、在Gin handle function 定义中现进行数据交互（查询、更新、删除、新增等）
 * 4、在main函数中启动Gin 路由配置，然后启动Gin Web service
 *
 * 缺点：
 * 	没有项目结构的划分，数据库连接、模型定义、Gin handle 函数等都混在一起
 */

/**
 * 1、定义Todo model entity
 */

// Todo model entity
type Todo struct {
	ID        uint      `json:"id"`
	Title     string    `json:"title"`
	Status    uint8     `json:"status"`
	CreatedAt time.Time `json:"created_time"`
	UpdatedAt time.Time `json:"updated_time"`
}

// TableName ...
func (Todo) TableName() string {
	return "bb_todo_v1"
}

/**
 * 2、配置MySQL数据库链接
 */
var db *gorm.DB

func init() {
	var err error
	var constr string
	constr = fmt.Sprintf("%s:%s@(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local", "godemo", "godemo", "localhost", 3306, "godemo")
	fmt.Println(constr)
	db, err = gorm.Open("mysql", constr)
	if err != nil {
		panic("connect db failed!")
	}
	// 把Todo model 实体更新到数据库中，即创建对应的数据库表
	db.AutoMigrate(&Todo{})
}

/********************************************************************************/
/**
 * Gin web handle function
 */

// TodoHandle http to handle， 其实这里可以直接使用 Todo 模型也行
type TodoHandle struct {
	ID        uint      `json:"id"`
	Title     string    `json:"title"`
	Status    uint8     `json:"status"`
	CreatedAt time.Time `json:"created_time"`
	UpdatedAt time.Time `json:"updated_time"`
}

// HelloFunc define http handle function
func HelloFunc(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  200,
		"message": "hello Todo",
	})
}

// AddTodo 新增Todo 接口
func AddTodo(c *gin.Context) {
	var todo Todo
	// 通过 Json body 提交数据，使用 BindJson 把body数据和模型进行绑定
	if err := c.BindJSON(&todo); err != nil {
		fmt.Println("bind json error: ", err)
		panic(err)
	}

	if err := db.Debug().Create(&todo).Error; err != nil {
		fmt.Println("create error: ", err)
		panic(err)
	}
	c.JSON(http.StatusOK, gin.H{
		"status":  200,
		"message": "ok",
		"data":    todo,
	})

}

// GetTodo 获取单个记录接口
func GetTodo(c *gin.Context) {
	var (
		td  Todo
		err error
	)
	// Gin的 param 形式获取参数 /todo/:id
	id := c.Param("id")

	err = db.Debug().First(&td, id).Error
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  5005,
			"message": "record not found",
			"data":    "",
		})
		return
	}

	data := TodoHandle{
		ID:     td.ID,
		Title:  td.Title,
		Status: td.Status,
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  200,
		"message": "ok",
		"data":    data,
	})
}

// GetTodoList 获取单个记录接口
func GetTodoList(c *gin.Context) {
	var (
		tds    []Todo
		err    error
		status int
	)
	// Gin 的Query模式获取参数 /todo?status=1
	queryStatus := c.Query("status")
	if queryStatus == "" {
		status = 1
	} else {
		// c.Param 和  c.Query 获取到的是字符串格式，需要转化成 数字
		status, _ = strconv.Atoi(queryStatus)
	}

	err = db.Debug().Where("status = ? ", status).Find(&tds).Error
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  5005,
			"message": "record not found",
			"data":    "",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  200,
		"message": "ok",
		"data":    tds,
	})
}

// UpdateTodo 更新Todo Item
func UpdateTodo(c *gin.Context) {
	var (
		td  Todo
		err error
	)
	id := c.Param("id")
	err = db.Debug().First(&td, id).Error
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  5005,
			"message": "record not found",
			"data":    "",
		})
		return
	}

	if err := c.BindJSON(&td); err != nil {
		fmt.Println("bind json error: ", err)
		panic(err)
	}

	// 更新的时候有两种情况
	// 第一种是更新单个字段 Update("column", "newValue")
	// 第二种是更新多个字段，这个时候有个问题存在，
	// 采用 struct 方式更新，只会更新非零值字段，比如同时更新title和status, 这里如果更新 status=0 是不生效的
	// 可以采用 map[string]interface{} 的方式更新，不会有非零值不能更新的问题
	// 另外更新的时候需要明确更新的 Model，比如这里的 Model(&td)
	if err := db.Debug().Model(&td).Update(map[string]interface{}{"title": td.Title, "status": td.Status}).Error; err != nil {
		fmt.Println("update failed: ", err)
		panic(err)
	}
	c.JSON(http.StatusOK, gin.H{
		"status":  200,
		"message": "ok",
		"data":    td,
	})
}

// DeleteTodo 删除Todo Item
func DeleteTodo(c *gin.Context) {
	var (
		td  Todo
		err error
	)
	id := c.Param("id")
	err = db.Debug().First(&td, id).Error
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  5005,
			"message": "record not found",
			"data":    "",
		})
		return
	}

	if err := db.Debug().Delete(&td).Error; err != nil {
		fmt.Println("delete failed: ", err)
		panic(err)
	}
	c.JSON(http.StatusOK, gin.H{
		"status":  200,
		"message": "delete success",
		"data":    "",
	})
}
func main() {
	fmt.Println("Todo V1 version")

	r := gin.Default()
	v1 := r.Group("/api/v1")
	{
		v1.GET("/hello", HelloFunc)

		v1.POST("/todo", AddTodo)          // 添加新条目
		v1.GET("/todo", GetTodoList)       // 查询所有条目
		v1.GET("/todo/:id", GetTodo)       // 获取单个条目
		v1.PUT("/todo/:id", UpdateTodo)    // 更新单个条目
		v1.DELETE("/todo/:id", DeleteTodo) // 删除单个条目
	}
	r.Run(":9999")

}
