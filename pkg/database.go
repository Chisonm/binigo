package binigo

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"
	"time"

	_ "github.com/lib/pq" // PostgreSQL driver
)

// DB represents a database connection
type DB struct {
	conn   *sql.DB
	config DatabaseConfig
}

// NewDB creates a new database connection
func NewDB(config DatabaseConfig) (*DB, error) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		config.Host, config.Port, config.Username, config.Password, config.Database)

	conn, err := sql.Open(config.Driver, dsn)
	if err != nil {
		return nil, err
	}

	// Test connection
	if err := conn.Ping(); err != nil {
		return nil, err
	}

	return &DB{
		conn:   conn,
		config: config,
	}, nil
}

// Close closes the database connection
func (db *DB) Close() error {
	return db.conn.Close()
}

// Model represents a base model with common fields
type Model struct {
	ID        int64     `db:"id" json:"id"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

// QueryBuilder provides a fluent interface for building queries
type QueryBuilder struct {
	db          *DB
	table       string
	selectCols  []string
	whereClause []string
	whereArgs   []interface{}
	orderBy     string
	limitVal    int
	offsetVal   int
	joins       []string
}

// Table starts a new query builder for a table
func (db *DB) Table(table string) *QueryBuilder {
	return &QueryBuilder{
		db:         db,
		table:      table,
		selectCols: []string{"*"},
	}
}

// Select specifies columns to select
func (qb *QueryBuilder) Select(columns ...string) *QueryBuilder {
	qb.selectCols = columns
	return qb
}

// Where adds a WHERE clause
func (qb *QueryBuilder) Where(column string, operator string, value interface{}) *QueryBuilder {
	qb.whereClause = append(qb.whereClause, fmt.Sprintf("%s %s ?", column, operator))
	qb.whereArgs = append(qb.whereArgs, value)
	return qb
}

// WhereIn adds a WHERE IN clause
func (qb *QueryBuilder) WhereIn(column string, values []interface{}) *QueryBuilder {
	placeholders := make([]string, len(values))
	for i := range values {
		placeholders[i] = "?"
		qb.whereArgs = append(qb.whereArgs, values[i])
	}

	qb.whereClause = append(qb.whereClause,
		fmt.Sprintf("%s IN (%s)", column, strings.Join(placeholders, ",")))
	return qb
}

// OrWhere adds an OR WHERE clause
func (qb *QueryBuilder) OrWhere(column string, operator string, value interface{}) *QueryBuilder {
	if len(qb.whereClause) > 0 {
		qb.whereClause = append(qb.whereClause, fmt.Sprintf("OR %s %s ?", column, operator))
	} else {
		qb.whereClause = append(qb.whereClause, fmt.Sprintf("%s %s ?", column, operator))
	}
	qb.whereArgs = append(qb.whereArgs, value)
	return qb
}

// Join adds a JOIN clause
func (qb *QueryBuilder) Join(table, condition string) *QueryBuilder {
	qb.joins = append(qb.joins, fmt.Sprintf("INNER JOIN %s ON %s", table, condition))
	return qb
}

// LeftJoin adds a LEFT JOIN clause
func (qb *QueryBuilder) LeftJoin(table, condition string) *QueryBuilder {
	qb.joins = append(qb.joins, fmt.Sprintf("LEFT JOIN %s ON %s", table, condition))
	return qb
}

// OrderBy adds ORDER BY clause
func (qb *QueryBuilder) OrderBy(column string, direction ...string) *QueryBuilder {
	dir := "ASC"
	if len(direction) > 0 {
		dir = strings.ToUpper(direction[0])
	}
	qb.orderBy = fmt.Sprintf("%s %s", column, dir)
	return qb
}

// Limit sets the LIMIT clause
func (qb *QueryBuilder) Limit(limit int) *QueryBuilder {
	qb.limitVal = limit
	return qb
}

// Offset sets the OFFSET clause
func (qb *QueryBuilder) Offset(offset int) *QueryBuilder {
	qb.offsetVal = offset
	return qb
}

// buildQuery constructs the SQL query
func (qb *QueryBuilder) buildQuery() string {
	query := fmt.Sprintf("SELECT %s FROM %s",
		strings.Join(qb.selectCols, ", "), qb.table)

	// Add joins
	if len(qb.joins) > 0 {
		query += " " + strings.Join(qb.joins, " ")
	}

	// Add WHERE
	if len(qb.whereClause) > 0 {
		query += " WHERE " + strings.Join(qb.whereClause, " AND ")
	}

	// Add ORDER BY
	if qb.orderBy != "" {
		query += " ORDER BY " + qb.orderBy
	}

	// Add LIMIT
	if qb.limitVal > 0 {
		query += fmt.Sprintf(" LIMIT %d", qb.limitVal)
	}

	// Add OFFSET
	if qb.offsetVal > 0 {
		query += fmt.Sprintf(" OFFSET %d", qb.offsetVal)
	}

	// Replace ? with $1, $2, etc. for PostgreSQL
	for i := 1; i <= len(qb.whereArgs); i++ {
		query = strings.Replace(query, "?", fmt.Sprintf("$%d", i), 1)
	}

	return query
}

// Get executes the query and returns all rows
func (qb *QueryBuilder) Get(dest interface{}) error {
	query := qb.buildQuery()

	rows, err := qb.db.conn.Query(query, qb.whereArgs...)
	if err != nil {
		return err
	}
	defer rows.Close()

	return scanRows(rows, dest)
}

// First executes the query and returns the first row
func (qb *QueryBuilder) First(dest interface{}) error {
	qb.Limit(1)
	query := qb.buildQuery()

	row := qb.db.conn.QueryRow(query, qb.whereArgs...)
	return scanRow(row, dest)
}

// Count returns the count of rows
func (qb *QueryBuilder) Count() (int64, error) {
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s", qb.table)

	if len(qb.whereClause) > 0 {
		query += " WHERE " + strings.Join(qb.whereClause, " AND ")
	}

	// Replace ? with $1, $2, etc. for PostgreSQL
	for i := 1; i <= len(qb.whereArgs); i++ {
		query = strings.Replace(query, "?", fmt.Sprintf("$%d", i), 1)
	}

	var count int64
	err := qb.db.conn.QueryRow(query, qb.whereArgs...).Scan(&count)
	return count, err
}

// Insert inserts a new record
func (qb *QueryBuilder) Insert(data Map) (int64, error) {
	columns := make([]string, 0, len(data))
	values := make([]interface{}, 0, len(data))
	placeholders := make([]string, 0, len(data))

	i := 1
	for col, val := range data {
		columns = append(columns, col)
		values = append(values, val)
		placeholders = append(placeholders, fmt.Sprintf("$%d", i))
		i++
	}

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s) RETURNING id",
		qb.table,
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "))

	var id int64
	err := qb.db.conn.QueryRow(query, values...).Scan(&id)
	return id, err
}

// Update updates records
func (qb *QueryBuilder) Update(data Map) (int64, error) {
	setClauses := make([]string, 0, len(data))
	values := make([]interface{}, 0, len(data)+len(qb.whereArgs))

	i := 1
	for col, val := range data {
		setClauses = append(setClauses, fmt.Sprintf("%s = $%d", col, i))
		values = append(values, val)
		i++
	}

	query := fmt.Sprintf("UPDATE %s SET %s", qb.table, strings.Join(setClauses, ", "))

	if len(qb.whereClause) > 0 {
		// Adjust placeholders for WHERE clause
		whereStr := strings.Join(qb.whereClause, " AND ")
		for j := 0; j < len(qb.whereArgs); j++ {
			whereStr = strings.Replace(whereStr, "?", fmt.Sprintf("$%d", i), 1)
			i++
		}
		query += " WHERE " + whereStr
		values = append(values, qb.whereArgs...)
	}

	result, err := qb.db.conn.Exec(query, values...)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}

// Delete deletes records
func (qb *QueryBuilder) Delete() (int64, error) {
	query := fmt.Sprintf("DELETE FROM %s", qb.table)

	if len(qb.whereClause) > 0 {
		whereStr := strings.Join(qb.whereClause, " AND ")
		for i := 1; i <= len(qb.whereArgs); i++ {
			whereStr = strings.Replace(whereStr, "?", fmt.Sprintf("$%d", i), 1)
		}
		query += " WHERE " + whereStr
	}

	result, err := qb.db.conn.Exec(query, qb.whereArgs...)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}

// Raw executes a raw SQL query
func (db *DB) Raw(query string, args ...interface{}) (*sql.Rows, error) {
	return db.conn.Query(query, args...)
}

// Exec executes a query without returning rows
func (db *DB) Exec(query string, args ...interface{}) (sql.Result, error) {
	return db.conn.Exec(query, args...)
}

// Transaction begins a transaction
func (db *DB) Transaction(fn func(*sql.Tx) error) error {
	tx, err := db.conn.Begin()
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()

	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}

// Helper functions for scanning

func scanRows(rows *sql.Rows, dest interface{}) error {
	destValue := reflect.ValueOf(dest)
	if destValue.Kind() != reflect.Ptr {
		return fmt.Errorf("dest must be a pointer")
	}

	sliceValue := destValue.Elem()
	if sliceValue.Kind() != reflect.Slice {
		return fmt.Errorf("dest must be a pointer to a slice")
	}

	elemType := sliceValue.Type().Elem()

	for rows.Next() {
		elem := reflect.New(elemType).Elem()

		if err := scanRow(rows, elem.Addr().Interface()); err != nil {
			return err
		}

		sliceValue.Set(reflect.Append(sliceValue, elem))
	}

	return rows.Err()
}

func scanRow(scanner interface{}, dest interface{}) error {
	// Simplified scan - in production, use a proper struct scanner
	// like sqlx or implement full reflection-based scanning
	return fmt.Errorf("implement full struct scanning based on db tags")
}
