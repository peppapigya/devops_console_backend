package executor

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
)

type SQLExecutor struct{}

func NewSQLExecutor() *SQLExecutor {
	return &SQLExecutor{}
}

func (e *SQLExecutor) GetType() string {
	return "sql"
}

func (e *SQLExecutor) Validate(config map[string]interface{}) error {
	sqlStr := getString(config, "sql", "")
	if sqlStr == "" {
		return fmt.Errorf("SQL语句不能为空")
	}
	return nil
}

func (e *SQLExecutor) Execute(ctx context.Context, execCtx *TaskExecutionContext) *ExecutionResult {
	startTime := time.Now()

	dbInstanceID := getUint64(execCtx.Config, "db_instance_id", 0)
	database := getString(execCtx.Config, "database", "")
	sqlStr := getString(execCtx.Config, "sql", "")

	execCtx.Logger.Log("info", fmt.Sprintf("开始执行SQL，数据库实例ID: %d", dbInstanceID))
	execCtx.Logger.Log("info", fmt.Sprintf("SQL语句: %s", truncateString(sqlStr, 200)))

	dbConfig, err := getDBInstanceConfig(dbInstanceID)
	if err != nil {
		return &ExecutionResult{
			Success:  false,
			ErrorMsg: fmt.Sprintf("获取数据库配置失败: %v", err),
			Duration: time.Since(startTime).Milliseconds(),
		}
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True",
		dbConfig.Username, dbConfig.Password, dbConfig.Host, dbConfig.Port, database)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return &ExecutionResult{
			Success:  false,
			ErrorMsg: fmt.Sprintf("连接数据库失败: %v", err),
			Duration: time.Since(startTime).Milliseconds(),
		}
	}
	defer db.Close()

	result, err := db.ExecContext(ctx, sqlStr)
	duration := time.Since(startTime).Milliseconds()

	if err != nil {
		execCtx.Logger.Log("error", fmt.Sprintf("SQL执行失败: %v", err))
		return &ExecutionResult{
			Success:  false,
			ErrorMsg: err.Error(),
			Duration: duration,
		}
	}

	rowsAffected, _ := result.RowsAffected()
	lastID, _ := result.LastInsertId()

	execCtx.Logger.Log("info", fmt.Sprintf("SQL执行成功，影响行数: %d, 耗时: %dms", rowsAffected, duration))

	return &ExecutionResult{
		Success:  true,
		Output:   fmt.Sprintf("影响行数: %d, 最后插入ID: %d", rowsAffected, lastID),
		Duration: duration,
		OutputVars: map[string]interface{}{
			"rows_affected":  rowsAffected,
			"last_insert_id": lastID,
		},
	}
}

type DBInstanceConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	Type     string
}

func getDBInstanceConfig(id uint64) (*DBInstanceConfig, error) {
	return &DBInstanceConfig{
		Host:     "localhost",
		Port:     3306,
		Username: "root",
		Password: "password",
		Type:     "mysql",
	}, nil
}
