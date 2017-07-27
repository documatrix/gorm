package gorm

import (
	"strings"
)

func (db *DB) TestExtension() *expr {
	return &expr{expr: "test"}
}

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

func (e *expr) Gt(value interface{}) *expr {
	e.expr = "(" + e.expr + " > ?)"
	e.args = append(e.args, value)

	return e
}

func (e *expr) Ge(value interface{}) *expr {
	e.expr = "(" + e.expr + " >= ?)"
	e.args = append(e.args, value)

	return e
}

func (e *expr) Lt(value interface{}) *expr {
	e.expr = "(" + e.expr + " < ?)"
	e.args = append(e.args, value)

	return e
}

func (e *expr) Le(value interface{}) *expr {
	e.expr = "(" + e.expr + " <= ?)"
	e.args = append(e.args, value)

	return e
}

func (e *expr) Like(value interface{}) *expr {
	e.expr = "(" + e.expr + " LIKE ?)"
	e.args = append(e.args, value)

	return e
}

func (e *expr) Eq(value interface{}) *expr {
	if value == nil {
		e.expr = "(" + e.expr + " IS NULL)"
	} else {
		e.expr = "(" + e.expr + " = ?)"
		e.args = append(e.args, value)
	}

	return e
}

func (e *expr) Neq(value interface{}) *expr {
	if value == nil {
		e.expr = "(" + e.expr + " IS NOT NULL)"
	} else {
		e.expr = "(" + e.expr + " != ?)"
		e.args = append(e.args, value)
	}

	return e
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
