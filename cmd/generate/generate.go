package main

import (
	"devops-console-backend/pkg/configs"
	"fmt"
	"log"
	"strings"

	"gorm.io/driver/mysql"
	"gorm.io/gen"
	"gorm.io/gorm"
)

func toCamelCase(str string) string {
	if str == "" {
		return ""
	}

	str = strings.Trim(str, "_")
	parts := strings.Split(str, "_")

	for i := 0; i < len(parts); i++ {
		if parts[i] == "" {
			continue
		}
		if i == 0 {
			parts[i] = strings.ToLower(parts[i][:1]) + parts[i][1:]
		} else {
			parts[i] = strings.ToUpper(parts[i][:1]) + parts[i][1:]
		}
	}

	return strings.Join(parts, "")
}

func main() {
	// 加载配置文件
	err := configs.LoadConfig()
	if err != nil {
		log.Printf("config load faild: %v", err)
		return
	}
	databaseConfig := configs.Config.Database.MySQL
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		databaseConfig.Username, databaseConfig.Password, databaseConfig.Host, databaseConfig.Port, databaseConfig.Database)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	// 创建代码生成器
	g := gen.NewGenerator(gen.Config{
		// 查询代码输出目录
		OutPath:           "./internal/dal/query", // 代码生成目录
		ModelPkgPath:      "./model",              // 模型包名相对于上面的路径名称
		Mode:              gen.WithDefaultQuery | gen.WithQueryInterface | gen.WithoutContext,
		FieldNullable:     true, //当字段为可为空时生成指针
		FieldSignable:     true, // 当字段为数字类型时，生成带符号的指针
		FieldWithIndexTag: true, // 生成索引标签
		FieldWithTypeTag:  true, // 生成类型标签

	})
	// 设置数据库
	g.UseDB(db)
	g.WithJSONTagNameStrategy(toCamelCase)

	// 生成所有的表
	g.ApplyBasic(g.GenerateAllTable()...)
	// 执行代码生成
	g.Execute()

}

// 为User模型生成自定义查询方法

type Querier interface {
	// SELECT * FROM @@table WHERE name = @name AND age = @age
	FindByNameAndAge(name string, age int) ([]gen.T, error)

	// UPDATE @@table SET deleted_at = NOW() WHERE id = @id
	DeleteByID(id uint) error
}
