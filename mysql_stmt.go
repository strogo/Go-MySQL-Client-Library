package mysql

import (
	"bufio"
	"encoding/binary"
	"os"
	"fmt"
	"strconv"
)

type MySQLStatement struct {
	StatementId uint32
	Columns     uint16
	Parameters  uint16
	Warnings    uint16
	FieldCount  uint8

	ResultSet *MySQLResultSet
	mysql     *MySQLInstance
}

// Encode values fro each field
func encodeParamValues(a []interface{}) ([]byte, int) {
	var b []byte
	for i := 0; i < len(a); i++ {
		f := a[i];
		switch t := f.(type) {
		case string:
			b = append(b, packString(string(t))...)
		case int:
			b = append(b, packString(strconv.Itoa(int(t)))...)
		case float:
			b = append(b, packString(strconv.Ftoa(float(t), 'f', -1))...)
		}
	}
	return b, len(b)
}

func putuint16(b []byte, v uint16) {
	b[0] = byte(v)
	b[1] = byte(v >> 8)
}

// For each field encode 2 byte type code. First bit is signed/unsigned
// Cheats and only bind parameters as strings
func encodeParamTypes(a []interface{}) ([]byte, int) {
	buf := make([]byte, len(a)*2)
	off := 0
	for i := 0; i < len(a); i++ {
		f := a[i];
		if f == nil {
			putuint16(buf[off:off+2], uint16(MYSQL_TYPE_NULL))
			continue
		}
		putuint16(buf[off:off+2], uint16(MYSQL_TYPE_STRING))
		off += 2
	}
	return buf, off
}


func readPrepareInit(br *bufio.Reader) (s *MySQLStatement, err os.Error) {
	var ph *PacketHeader
	ph, err = readHeader(br)
	if err != nil { return }
	s = new(MySQLStatement)
	err = binary.Read(br, binary.LittleEndian, &s.FieldCount)
	if err != nil { return }
	if s.FieldCount == uint8(0xff) {
		return nil, readErrorPacket(br, int(ph.Len))
	}
	err = binary.Read(br, binary.LittleEndian, &s.StatementId)
	if err != nil { return }
	err = binary.Read(br, binary.LittleEndian, &s.Columns)
	if err != nil { return }
	err = binary.Read(br, binary.LittleEndian, &s.Parameters)
	if err != nil { return }
	if ph.Len >= 12 {
		err = ignoreBytes(br, 1)
		if err != nil { return }
		err = binary.Read(br, binary.LittleEndian, &s.Warnings)
		if err != nil { return }
	}
	return
}

//Currently just skips the pakets as I'm not sure if they are useful.
func readPrepareParameters(br *bufio.Reader, s *MySQLStatement) os.Error {
	for i := uint16(0); i < s.Parameters; i++ {
		ph, err := readHeader(br)
		if err != nil { return err }
		err = ignoreBytes(br, ph.Len)
		if err != nil { return err }
	}
	return nil
}

func (sth *MySQLStatement) execute(va []interface{}) (res *MySQLResponse, err os.Error) {
	if va == nil || int(sth.Parameters) != len(va) {
		return nil, os.ErrorString(fmt.Sprintf("Parameter count mismatch. %d != %d", sth.Parameters, len(va)))
	}
	type_parm, tn := encodeParamTypes(va)
	value_parm, vn := encodeParamValues(va)
	bitmap_len := (len(va) + 7) / 8
	mysql := sth.mysql
	err = packUint24(mysql.writer, uint32(11+bitmap_len+tn+vn))
	if err != nil { return }
	err = packUint8(mysql.writer, uint8(0))
	if err != nil { return }
	err = packUint8(mysql.writer, uint8(COM_STMT_EXECUTE))
	if err != nil { return }
	err = packUint32(mysql.writer, uint32(sth.StatementId))
	if err != nil { return }
	err = packUint8(mysql.writer, uint8(0))
	if err != nil { return }
	err = packUint32(mysql.writer, uint32(1))
	if err != nil { return }
	b := make([]byte, bitmap_len)
	_, err = mysql.writer.Write(b) //TODO: Support null params.
	if err != nil { return }
	err = packUint8(mysql.writer, uint8(1))
	if err != nil { return }
	_, err = mysql.writer.Write(type_parm)
	if err != nil { return }
	_, err = mysql.writer.Write(value_parm)
	if err != nil { return }
	err = mysql.writer.Flush()
	if err != nil { return }
	res, err = mysql.readResult()
	if err != nil { return }
	res.Prepared = true
	return
}

func (mysql *MySQLInstance) prepare(arg string) (sth *MySQLStatement, err os.Error) {
	plen := len(arg) + 1
	head := make([]byte,5)
	head[0] = byte(plen)
	head[1] = byte(plen >> 8)
	head[2] = byte(plen >> 16)
	head[3] = 0
	head[4] = uint8(COM_STMT_PREPARE)
	_, err = mysql.writer.Write(head)
	if err != nil { return }
	_, err = mysql.writer.WriteString(arg)
	if err != nil { return }
	err = mysql.writer.Flush()
	if err != nil { return }
	sth, err = readPrepareInit(mysql.reader)
	if err != nil { return }
	if sth.Parameters > 0 {
		err = readPrepareParameters(mysql.reader, sth)
		if err != nil { return }
		err = readEOFPacket(mysql.reader)
		if err != nil { return }
	}
	if sth.Columns > 0 {
		rs, err := mysql.readResultSet(uint64(sth.Columns))
		if err != nil { return nil, err }
		sth.ResultSet = rs
	}
	sth.mysql = mysql
	return
}
