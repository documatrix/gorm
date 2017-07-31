package gorm

import (
	"reflect"
	"strings"
)

type lexpr struct {
	expr string
}

type jexpr struct {
	expr string
	args []interface{}
}

func (db *DB) InnerJoin(model interface{}, alias ...string) *jexpr {
	var al string
	if len(alias) > 0 {
		al = alias[0]
	}

	if val, ok := model.(*expr); ok {
		return &jexpr{expr: " INNER JOIN (" + val.expr + ") " + al, args: val.args}
	}
	return &jexpr{expr: " INNER JOIN " + db.T(model) + " " + al}
}

func (je *jexpr) On(col1 *lexpr, col2 *lexpr) *expr {
	return &expr{expr: je.expr + " ON " + col1.expr + " = " + col2.expr, args: je.args}
}

func (db *DB) L(model interface{}, name string) *lexpr {
	scope := db.NewScope(model)
	field, _ := scope.FieldByName(name)
	return &lexpr{expr: scope.Quote(scope.TableName()) + "." + scope.Quote(field.DBName)}
}

func (db *DB) LA(model interface{}, alias string, name string) *lexpr {
	scope := db.NewScope(model)
	field, _ := scope.FieldByName(name)
	return &lexpr{expr: scope.Quote(alias) + "." + scope.Quote(field.DBName)}
}

func (db *DB) C(model interface{}, names ...string) string {
	columns := make([]string, 0)

	scope := db.NewScope(model)
	for _, name := range names {
		field, _ := scope.FieldByName(name)
		columns = append(columns, field.DBName)
	}

	return strings.Join(columns, ", ")
}

func (db *DB) CA(model interface{}, alias string, names ...string) string {
	columns := make([]string, 0)

	for _, name := range names {
		columns = append(columns, db.LA(model, alias, name).expr)
	}

	return strings.Join(columns, ", ")
}

func (db *DB) CQ(model interface{}, names ...string) string {
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

func (e *lexpr) operator(operator string, value interface{}) *expr {
	if value == nil {
		return &expr{expr: "(" + e.expr + " " + operator + " )"}
	}

	if val, ok := value.(*lexpr); ok {
		return &expr{expr: "(" + e.expr + " " + operator + " " + val.expr + ")"}
	}

	if _, ok := value.(*expr); ok {
		e.expr = "(" + e.expr + " " + operator + " (?))"
	} else {
		e.expr = "(" + e.expr + " " + operator + " ?)"
	}

	return &expr{expr: e.expr, args: append(make([]interface{}, 0), value)}
}

func (e *lexpr) Gt(value interface{}) *expr {
	return e.operator(">", value)
}

func (e *lexpr) Ge(value interface{}) *expr {
	return e.operator(">=", value)
}

func (e *lexpr) Lt(value interface{}) *expr {
	return e.operator("<", value)
}

func (e *lexpr) Le(value interface{}) *expr {
	return e.operator("<=", value)
}

func (e *lexpr) Like(value interface{}) *expr {
	return e.operator("LIKE", value)
}

func (e *lexpr) Eq(value interface{}) *expr {
	if value == nil {
		return e.operator("IS NULL", value)
	}

	return e.operator("=", value)
}

func (e *lexpr) Neq(value interface{}) *expr {
	if value == nil {
		return e.operator("IS NOT NULL", value)
	}

	return e.operator("!=", value)
}

func (e *lexpr) In(values ...interface{}) *expr {
	// NOTE: Maybe there is a better way to do this? :)
	qm := make([]string, len(values))
	for i := 0; i < len(values); i++ {
		qm[i] = "?"
	}

	return &expr{expr: "(" + e.expr + " IN (" + strings.Join(qm, ",") + "))", args: append(make([]interface{}, 0), values...)}
}

func (e *lexpr) OrderAsc() string {
	return e.expr + " ASC "
}

func (e *lexpr) OrderDesc() string {
	return e.expr + " DESC "
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

func (db *DB) UpdateFields(fields ...string) *DB {
	sets := make(map[string]interface{})
	m := reflect.ValueOf(db.Value).Elem()
	for _, field := range fields {
		sets[db.C(db.Value, field)] = m.FieldByName(field).Interface()
	}

	return db.Update(sets)
}

func (e *expr) Intersect(e2 *expr) *expr {
	e.expr = "((" + e.expr + ") INTERSECT (" + e2.expr + "))"
	e.args = append(e.args, e2.args...)

	return e
}

func (e *lexpr) Alias(alias string) *lexpr {
	e.expr = e.expr + " " + alias + " "

	return e
}

func (e *expr) Alias(alias string) *expr {
	e.expr = e.expr + " " + alias + " "

	return e
}

