package gorpmapping

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/go-gorp/gorp"
	"github.com/lib/pq"

	"github.com/ovh/cds/sdk"
)

const (
	// ViolateForeignKeyPGCode is the pg code when violating foreign key
	ViolateForeignKeyPGCode = "23503"

	// ViolateUniqueKeyPGCode is the pg code when duplicating unique key
	ViolateUniqueKeyPGCode = "23505"

	// StringDataRightTruncation is raisedalue is too long for varchar.
	StringDataRightTruncation = "22001"
)

// NewQuery returns a new query from given string request.
func NewQuery(q string) Query { return Query{query: q} }

// Query to get gorp entities in database.
type Query struct {
	query     string
	arguments []interface{}
}

// Args store query arguments.
func (q Query) Args(as ...interface{}) Query {
	q.arguments = as
	return q
}

// IDsToQueryString returns a comma separated list of given ids.
func IDsToQueryString(ids []int64) string {
	res := make([]string, len(ids))
	for i := range ids {
		res[i] = fmt.Sprintf("%d", ids[i])
	}
	return strings.Join(res, ",")
}

// Insert value in given db.
func Insert(db gorp.SqlExecutor, i interface{}) error {
	err := db.Insert(i)
	if e, ok := err.(*pq.Error); ok {
		switch e.Code {
		case ViolateUniqueKeyPGCode:
			err = sdk.NewError(sdk.ErrInvalidData, e)
		case StringDataRightTruncation:
			err = sdk.NewError(sdk.ErrConflict, e)
		}
	}
	return sdk.WithStack(err)
}

// Update value in given db.
func Update(db gorp.SqlExecutor, i interface{}) error {
	_, err := db.Update(i)
	if e, ok := err.(*pq.Error); ok {
		switch e.Code {
		case ViolateUniqueKeyPGCode:
			err = sdk.NewError(sdk.ErrInvalidData, e)
		case StringDataRightTruncation:
			err = sdk.NewError(sdk.ErrInvalidData, e)
		}
	}
	return sdk.WithStack(err)
}

// Delete value in given db.
func Delete(db gorp.SqlExecutor, i interface{}) error {
	_, err := db.Delete(i)
	return sdk.WithStack(err)
}

// GetAll values from database.
func GetAll(db gorp.SqlExecutor, q Query, i interface{}) error {
	_, err := db.Select(i, q.query, q.arguments...)
	return sdk.WithStack(err)
}

// Get a value from database.
func Get(db gorp.SqlExecutor, q Query, i interface{}) (bool, error) {
	if err := db.SelectOne(i, q.query, q.arguments...); err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, sdk.WithStack(err)
	}
	return true, nil
}
