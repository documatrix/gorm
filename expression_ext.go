package gorm

import (
	"reflect"
	"strings"
)

type LExpr struct {
	expr string
}

type jexpr struct {
	expr string
	args []interface{}
}

func join(joinType string, db *DB, model interface{}, alias ...string) *jexpr {
	var al string
	if len(alias) > 0 {
		al = alias[0]
	}

	if val, ok := model.(*expr); ok {
		return &jexpr{expr: " " + joinType + " JOIN (" + val.expr + ") " + al, args: val.args}
	}
	return &jexpr{expr: " " + joinType + " JOIN " + db.T(model) + " " + al}
}

func (db *DB) InnerJoin(model interface{}, alias ...string) *jexpr {
	return join("INNER", db, model, alias...)
}

func (db *DB) LeftJoin(model interface{}, alias ...string) *jexpr {
	return join("LEFT", db, model, alias...)
}

func (db *DB) RightJoin(model interface{}, alias ...string) *jexpr {
	return join("RIGHT", db, model, alias...)
}

func (db *DB) OuterJoin(model interface{}, alias ...string) *jexpr {
	return join("OUTER", db, model, alias...)
}

func (je *jexpr) On(col1 *LExpr, col2 *LExpr) *expr {
	return &expr{expr: je.expr + " ON " + col1.expr + " = " + col2.expr, args: je.args}
}

func (db *DB) L(model interface{}, name string) *LExpr {
	scope := db.NewScope(model)
	field, _ := scope.FieldByName(name)
	return &LExpr{expr: scope.Quote(scope.TableName()) + "." + scope.Quote(field.DBName)}
}

func (db *DB) LA(model interface{}, alias string, name string) *LExpr {
	scope := db.NewScope(model)
	field, _ := scope.FieldByName(name)
	return &LExpr{expr: scope.Quote(alias) + "." + scope.Quote(field.DBName)}
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

func (e *LExpr) operator(operator string, value interface{}) *expr {
	if value == nil {
		return &expr{expr: "(" + e.expr + " " + operator + " )"}
	}

	if val, ok := value.(*LExpr); ok {
		return &expr{expr: "(" + e.expr + " " + operator + " " + val.expr + ")"}
	}

	if _, ok := value.(*expr); ok {
		e.expr = "(" + e.expr + " " + operator + " (?))"
	} else {
		e.expr = "(" + e.expr + " " + operator + " ?)"
	}

	return &expr{expr: e.expr, args: append(make([]interface{}, 0), value)}
}

func (e *LExpr) Gt(value interface{}) *expr {
	return e.operator(">", value)
}

func (e *LExpr) Ge(value interface{}) *expr {
	return e.operator(">=", value)
}

func (e *LExpr) Lt(value interface{}) *expr {
	return e.operator("<", value)
}

func (e *LExpr) Le(value interface{}) *expr {
	return e.operator("<=", value)
}

func (e *LExpr) Like(value interface{}) *expr {
	return e.operator("LIKE", value)
}

func (e *LExpr) Eq(value interface{}) *expr {
	if value == nil {
		return e.operator("IS NULL", value)
	}

	return e.operator("=", value)
}

func (e *LExpr) Neq(value interface{}) *expr {
	if value == nil {
		return e.operator("IS NOT NULL", value)
	}

	return e.operator("!=", value)
}

func (e *LExpr) In(values ...interface{}) *expr {
	// NOTE: Maybe there is a better way to do this? :)
	if len(values) == 1 {
		if s := reflect.ValueOf(values[0]); s.Kind() == reflect.Slice {
			vals := make([]interface{}, s.Len())
			qm := make([]string, s.Len())

			for i := 0; i < s.Len(); i++ {
				vals[i] = s.Index(i).Interface()
				qm[i] = "?"
			}

			return &expr{expr: "(" + e.expr + " IN (" + strings.Join(qm, ",") + "))", args: vals}
		}
	}

	qm := make([]string, len(values))
	for i := 0; i < len(values); i++ {
		qm[i] = "?"
	}

	return &expr{expr: "(" + e.expr + " IN (" + strings.Join(qm, ",") + "))", args: append(make([]interface{}, 0), values...)}
}

func (e *LExpr) OrderAsc() string {
	return e.expr + " ASC "
}

func (e *LExpr) OrderDesc() string {
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

func (db *DB) SelectFields(fields ...string) *DB {
	selects := strings.Join(fields, ", ")

	return db.Select(selects)
}

func (e *expr) Intersect(e2 *expr) *expr {
	e.expr = "((" + e.expr + ") INTERSECT (" + e2.expr + "))"
	e.args = append(e.args, e2.args...)

	return e
}

func (e *LExpr) Alias(alias string) *LExpr {
	e.expr = e.expr + " " + alias + " "

	return e
}

func (e *expr) Alias(alias string) *expr {
	e.expr = e.expr + " " + alias + " "

	return e
}
