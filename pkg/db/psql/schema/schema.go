package schema

import (
	"context"
	"database/sql"
	"github.com/sjhitchner/toolbox/pkg/db/psql"
)

type Table struct {
	SchemaName     sql.NullString `db:"schemaname"`
	TableName      sql.NullString `db:"tablename"`
	TableOwner     sql.NullString `db:"tableowner"`
	TableSpace     sql.NullString `db:"tablespace"`
	HasIndexes     bool           `db:"hasindexes"`
	HasRules       bool           `db:"hasrules"`
	HasTriggers    bool           `db:"hastriggers"`
	HasRowSecurity bool           `db:"rowsecurity"`
}

type Column struct {
	TableCatalog           sql.NullString `db:"table_catalog"`
	TableSchema            sql.NullString `db:"table_schema"`
	TableName              sql.NullString `db:"table_name"`
	ColumnName             sql.NullString `db:"column_name"`
	OridinalPosition       sql.NullInt64  `db:"ordinal_position"`
	ColumnDefault          sql.NullString `db:"column_default"`
	IsNullable             sql.NullString `db:"is_nullable"`
	DataType               sql.NullString `db:"data_type"`
	CharactorMaximumLength sql.NullInt64  `db:"character_maximum_length"`
	CharactorOctetLength   sql.NullInt64  `db:"character_octet_length"`
	NumericPrecision       sql.NullInt64  `db:"numeric_precision"`
	UDTCatalog             sql.NullString `db:"udt_catalog"`
	UDTSchema              sql.NullString `db:"udt_schema"`
	UDTName                sql.NullString `db:"udt_name"`
	IsUpdatable            sql.NullString `db:"is_updatable"`
	/*
	   numeric_precision_radix  |
	   numeric_scale            |
	   datetime_precision       |
	   sql.NullInt64erval_type            |
	   sql.NullInt64erval_precision       |
	   character_set_catalog    |
	   character_set_schema     |
	   character_set_name       |
	   collation_catalog        |
	   collation_schema         |
	   collation_name           |
	   domain_catalog           |
	   domain_schema            |
	   domain_name              |
	   scope_catalog            |
	   scope_schema             |
	   scope_name               |
	   maximum_cardinality      |
	   dtd_identifier           | 1
	   is_self_referencing      | NO
	   is_identity              | NO
	   identity_generation      |
	   identity_start           |
	   identity_increment       |
	   identity_maximum         |
	   identity_minimum         |
	   identity_cycle           | NO
	   is_generated             | NEVER
	   generation_expression    |
	*/
}

const (
	TablesQuery = `
SELECT * FROM pg_tables WHERE schemaname='public'
`

	ColumnsQuery = `
SELECT 
  table_catalog
  , table_schema
  , table_name
  , column_name
  , ordinal_position
  , column_default
  , is_nullable
  , data_type
  , character_maximum_length
  , character_octet_length
  , numeric_precision
  , udt_catalog
  , udt_schema
  , udt_name
  , is_updatable
 FROM information_schema.COLUMNS WHERE table_name = $1 
`

	ColumnsFullQuery = `
SELECT 
  table_catalog
  , table_schema
  , table_name
  , column_name
  , ordinal_position
  , column_default
  , is_nullable
  , data_type
  , character_maximum_length
  , character_octet_length
  , numeric_precision
  , numeric_precision_radix
  , numeric_scale
  , datetime_precision
  , interval_type
  , interval_precision
  , character_set_catalog
  , character_set_schema
  , character_set_name
  , collation_catalog
  , collation_schema
  , collation_name
  , domain_catalog
  , domain_schema
  , domain_name
  , udt_catalog
  , udt_schema
  , udt_name
  , scope_catalog
  , scope_schema
  , scope_name
  , maximum_cardinality
  , dtd_identifier
  , is_self_referencing
  , is_identity
  , identity_generation
  , identity_start
  , identity_increment
  , identity_maximum
  , identity_minimum
  , identity_cycle
  , is_generated
  , generation_expression
  , is_updatable
 FROM information_schema.COLUMNS WHERE table_name = $1 
`
)

func GetTables(ctx context.Context, db *psql.PSQLHandler) ([]Table, error) {
	var tables []Table

	if err := db.Select(ctx, &tables, TablesQuery); err != nil {
		return nil, err
	}

	return tables, nil
}

func GetColumns(ctx context.Context, db *psql.PSQLHandler, table string) ([]Column, error) {
	var columns []Column

	if err := db.Select(ctx, &columns, ColumnsQuery, table); err != nil {
		return nil, err
	}

	return columns, nil
}
