package note

import (
	"testing"

	"github.com/huandu/go-sqlbuilder"
)

func TestSqlbuilder(t *testing.T) {
	// modExpr := sqlbuilder.Buildf("MOD(id, %v) = %v", 1, 2)
	// sql, args := modExpr.Build()
	// t.Logf("sql: %v, args: %v", sql, args)

	sb := sqlbuilder.NewSelectBuilder()
	modExpr := sqlbuilder.Buildf("MOD(id, %v) = %v", 3, 1)

	sb.Select("*").
		From("test_table").
		Where(
			sb.EQ("status", 0),
			sb.BuilderAs(modExpr, "mod_cond"),
		)

	sql, args := sb.Build()
	t.Logf("sql: %v", sql)
	t.Logf("args: %v", args)

}

func TestSqlbuilderWithSelect2(t *testing.T) {
	sb := sqlbuilder.NewSelectBuilder()

	sb.Args.Add(6)
	sb.Args.Add(7)
	sb.Select("*").
		From(procedureRecordTableName).
		Where(
			sb.GTE("next_check_time", 1),
			sb.LT("next_check_time", 2),
			sb.EQ("status", 3),
			sb.EQ("protype", 4),
			sb.GT("id", 5),
			"MOD(id, $1) = $2",
		)
	sb.OrderByAsc("next_check_time").
		OrderByAsc("id").
		Limit(110)
	sql, args := sb.Build()
	t.Log(sql, args)

	var record ProcedureRecordPO
	testDb.QueryRowCtx(t.Context(), &record, sql, args...)
	t.Logf("record: %+v", record)
}
