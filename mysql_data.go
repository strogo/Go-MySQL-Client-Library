package mysql

type MySQLField struct {
	Catalog		string;
	Db		string;
	Table		string;
	OrgTable	string;
	Name		string;
	OrgName		string;

	Charset		uint16;
	Length		uint32;
	Type		uint8;
	Flags		uint16;
	Decimals	uint8;
	Default		uint64;
}

func (f *MySQLField) String() string	{ return f.Name }

type MySQLData struct {
	Data	[]byte;
	Length	uint64;
	IsNull	bool;
	Type	uint8;
}

func (d *MySQLData) String() string {
	if d.IsNull {
		return "NULL"
	}
	return string(d.Data);
}

type MySQLRow struct {
	Data []*MySQLData;
}