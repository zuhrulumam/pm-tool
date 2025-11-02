package query_builder

import (
	"fmt"
	"strings"
	"time"
)

// QueryBuilder helps build dynamic SQL WHERE clauses
type QueryBuilder struct {
	whereClauses []string
	args         []interface{}
	argCounter   int
}

// NewQueryBuilder creates a new query builder
func NewQueryBuilder() *QueryBuilder {
	return &QueryBuilder{
		whereClauses: make([]string, 0),
		args:         make([]interface{}, 0),
		argCounter:   0,
	}
}

// AddFilter adds a filter condition to the query
func (qb *QueryBuilder) AddFilter(key string, value interface{}) {
	if value == nil {
		return
	}

	// Handle special filter operators
	if strings.HasSuffix(key, "_like") {
		column := strings.TrimSuffix(key, "_like")
		qb.whereClauses = append(qb.whereClauses, fmt.Sprintf("%s LIKE ?", column))
		qb.args = append(qb.args, fmt.Sprintf("%%%v%%", value))
		return
	}

	if strings.HasSuffix(key, "_from") {
		column := strings.TrimSuffix(key, "_from")
		qb.whereClauses = append(qb.whereClauses, fmt.Sprintf("%s >= ?", column))
		qb.args = append(qb.args, value)
		return
	}

	if strings.HasSuffix(key, "_to") {
		column := strings.TrimSuffix(key, "_to")
		qb.whereClauses = append(qb.whereClauses, fmt.Sprintf("%s <= ?", column))
		qb.args = append(qb.args, value)
		return
	}

	if strings.HasSuffix(key, "_in") {
		column := strings.TrimSuffix(key, "_in")
		// value should be a slice
		switch v := value.(type) {
		case []string:
			if len(v) > 0 {
				placeholders := strings.Repeat("?,", len(v))
				placeholders = placeholders[:len(placeholders)-1] // remove trailing comma
				qb.whereClauses = append(qb.whereClauses, fmt.Sprintf("%s IN (%s)", column, placeholders))
				for _, item := range v {
					qb.args = append(qb.args, item)
				}
			}
		case []int, []int64, []interface{}:
			// Handle other slice types if needed
			qb.whereClauses = append(qb.whereClauses, fmt.Sprintf("%s IN (?)", column))
			qb.args = append(qb.args, value)
		}
		return
	}

	if strings.HasSuffix(key, "_gt") {
		column := strings.TrimSuffix(key, "_gt")
		qb.whereClauses = append(qb.whereClauses, fmt.Sprintf("%s > ?", column))
		qb.args = append(qb.args, value)
		return
	}

	if strings.HasSuffix(key, "_lt") {
		column := strings.TrimSuffix(key, "_lt")
		qb.whereClauses = append(qb.whereClauses, fmt.Sprintf("%s < ?", column))
		qb.args = append(qb.args, value)
		return
	}

	if strings.HasSuffix(key, "_gte") {
		column := strings.TrimSuffix(key, "_gte")
		qb.whereClauses = append(qb.whereClauses, fmt.Sprintf("%s >= ?", column))
		qb.args = append(qb.args, value)
		return
	}

	if strings.HasSuffix(key, "_lte") {
		column := strings.TrimSuffix(key, "_lte")
		qb.whereClauses = append(qb.whereClauses, fmt.Sprintf("%s <= ?", column))
		qb.args = append(qb.args, value)
		return
	}

	if strings.HasSuffix(key, "_ne") {
		column := strings.TrimSuffix(key, "_ne")
		qb.whereClauses = append(qb.whereClauses, fmt.Sprintf("%s != ?", column))
		qb.args = append(qb.args, value)
		return
	}

	// Default: exact match
	qb.whereClauses = append(qb.whereClauses, fmt.Sprintf("%s = ?", key))
	qb.args = append(qb.args, value)
}

// AddFilters adds multiple filters from a map
func (qb *QueryBuilder) AddFilters(filters map[string]interface{}) {
	for key, value := range filters {
		qb.AddFilter(key, value)
	}
}

// Build returns the WHERE clause and arguments
func (qb *QueryBuilder) Build() (string, []interface{}) {
	if len(qb.whereClauses) == 0 {
		return "", nil
	}
	
	whereSQL := "WHERE " + strings.Join(qb.whereClauses, " AND ")
	return whereSQL, qb.args
}

// BuildWithoutWhere returns just the conditions without WHERE keyword
func (qb *QueryBuilder) BuildWithoutWhere() (string, []interface{}) {
	if len(qb.whereClauses) == 0 {
		return "", nil
	}
	
	return strings.Join(qb.whereClauses, " AND "), qb.args
}

// HasConditions returns true if there are any conditions
func (qb *QueryBuilder) HasConditions() bool {
	return len(qb.whereClauses) > 0
}

// Count returns the number of conditions
func (qb *QueryBuilder) Count() int {
	return len(qb.whereClauses)
}

// BuildWhereClause is a helper function to build WHERE clause from filters
func BuildWhereClause(filters map[string]interface{}) (string, []interface{}) {
	qb := NewQueryBuilder()
	qb.AddFilters(filters)
	return qb.Build()
}

// Example helper for common date range filtering
func AddDateRangeFilter(qb *QueryBuilder, column string, from, to *time.Time) {
	if from != nil {
		qb.AddFilter(column+"_from", *from)
	}
	if to != nil {
		qb.AddFilter(column+"_to", *to)
	}
}
