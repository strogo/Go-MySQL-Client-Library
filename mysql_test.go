package mysql

import (
	"testing"
)

func SelectSingleRow(t *testing.T, q string) map[string]string {
	dbh, err := Connect("tcp", "", "127.0.0.1:3306", "test", "test", "test")
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if dbh == nil {
		t.Error("dbh is nil")
		t.FailNow()
	}
	dbh.Use("test")

	res, err := dbh.Query("SET NAMES utf8")
	res, err = dbh.Query(q)

	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	row := res.FetchRowMap()
	if row == nil {
		t.Error("row is nil")
		t.FailNow()
	}
	dbh.Quit()
	return row
}

func SelectSingleRowPrepared(t *testing.T, q string, p ...) map[string]string {
	dbh, err := Connect("tcp", "", "127.0.0.1:3306", "test", "test", "test")
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if dbh == nil {
		t.Error("dbh is nil")
		t.FailNow()
	}
	dbh.Use("test")

	res, err := dbh.Query("SET NAMES utf8")
	sth, err := dbh.Prepare(q)

	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	res, err = sth.Execute(p)
	row := res.FetchRowMap()
	if row == nil {
		t.Error("row is nil")
		t.FailNow()
	}
	dbh.Quit()
	return row
}

func TestSelectString(t *testing.T) {
	row := SelectSingleRow(t, "SELECT * FROM test WHERE name='test1'")
	test := "1234567890abcdef"
	if row["stuff"] != test {
		t.Error(row["stuff"], " != ", test)
	}
}

func TestSelectStringPrepared(t *testing.T) {
	row := SelectSingleRowPrepared(t, "SELECT * FROM test WHERE name=?", "test1")
	test := "1234567890abcdef"
	if row["stuff"] != test {
		t.Error(row["stuff"], " != ", test)
	}
}

func TestSelectUFT8(t *testing.T) {
	row := SelectSingleRow(t, "SELECT * FROM test WHERE name='unicodetest1'")
	test := "l̡̡̡ ̴̡ı̴̴̡ ̡̡͡|̲̲̲͡͡͡ ̲▫̲͡ ̲̲̲͡͡π̲̲͡͡ ̲̲͡▫̲̲͡͡ ̲|̡̡̡ ̡ ̴̡ı̴̡̡ ̡"
	if row["stuff"] != test {
		t.Error(row["stuff"], " != ", test)
	}
}

func TestSelectUFT8Prepared(t *testing.T) {
	row := SelectSingleRowPrepared(t, "SELECT * FROM test WHERE name=?", "unicodetest1")
	test := "l̡̡̡ ̴̡ı̴̴̡ ̡̡͡|̲̲̲͡͡͡ ̲▫̲͡ ̲̲̲͡͡π̲̲͡͡ ̲̲͡▫̲̲͡͡ ̲|̡̡̡ ̡ ̴̡ı̴̡̡ ̡"
	if row["stuff"] != test {
		t.Error(row["stuff"], " != ", test)
	}
}
