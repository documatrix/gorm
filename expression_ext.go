package gorm

import (
	"strings"
)

func (db *DB) L(model interface{}, name string) *expr {
	scope := db.NewScope(model)
	field, _ := scope.FieldByName(name)
	return &expr{expr: scope.Quote(scope.TableName()) + "." + scope.Quote(field.DBName)}
}

func (db *DB) C(model interface{}, names ...string) string {
	columns := make([]string, 0)

	for _, name := range names {
		columns = append(columns, db.L(model, name).expr)
	}

	return strings.Join(columns, ", ")
}

func (db *DB) T(model interface{}) string {
	scope := db.NewScope(model)
	return scope.TableName()
}

func (db *DB) QT(model interface{}) string {
	scope := db.NewScope(model)
	return scope.QuotedTableName()
}

func (e *expr) operator(operator string, value interface{}) *expr {
	if _, ok := value.(*expr); ok {
		e.expr = "(" + e.expr + " " + operator + " (?))"
	} else {
		e.expr = "(" + e.expr + " " + operator + " ?)"
	}
	if value != nil {
		e.args = append(e.args, value)
	}

	return e
}

func (e *expr) Gt(value interface{}) *expr {
	return e.operator(">", value)
}

func (e *expr) Ge(value interface{}) *expr {
	return e.operator(">=", value)
}

func (e *expr) Lt(value interface{}) *expr {
	return e.operator("<", value)
}

func (e *expr) Le(value interface{}) *expr {
	return e.operator("<=", value)
}

func (e *expr) Like(value interface{}) *expr {
	return e.operator("LIKE", value)
}

func (e *expr) Eq(value interface{}) *expr {
	if value == nil {
		return e.operator("IS NULL", value)
	}

	return e.operator("=", value)
}

func (e *expr) Neq(value interface{}) *expr {
	if value == nil {
		return e.operator("IS NOT NULL", value)
	}

	return e.operator("!=", value)
}

func (e *expr) In(values ...interface{}) *expr {
	// NOTE: Maybe there is a better way to do this? :)
	qm := make([]string, len(values))
	for i := 0; i < len(values); i++ {
		qm[i] = "?"
	}

	e.expr = "(" + e.expr + " IN (" + strings.Join(qm, ",") + "))"
	e.args = append(e.args, values...)

	return e
}

func (e *expr) Or(e2 *expr) *expr {
	e.expr = "(" + e.expr + " OR " + e2.expr + ")"
	e.args = append(e.args, e2.args...)

	return e
}

func (e *expr) And(e2 *expr) *expr {
	e.expr = "(" + e.expr + " AND " + e2.expr + ")"
	e.args = append(e.args, e2.args...)

	return e
}

func (e *expr) OrderAsc() string {
	return e.expr + " ASC "
}

func (e *expr) OrderDesc() string {
	return e.expr + " DESC "
}
