package config

import (
	"fmt"
	"regexp"
	"strings"
	. "types"
)

var configTsvPath string = "../../conf/"

func init() {
	Config.RegisterTable(cfgBuildingTable)

	Config.RegisterParser(asTimeInt64)
	Config.RegisterParser(asMapStringString)
	Config.RegisterParser(asMapStringInt)
	Config.RegisterParser(asMapStringFloat32)

	LoadAllTable(configTsvPath)
	//DebugDumpAllTable()
}

type config struct {
	parser *Parser
	tables []IConfigTable
}

type IConfigTable interface {
	Init()
	Name() string
	NewRecord() interface{}
	AppendRecord(interface{})
	debugAllRecords() map[string]interface{}
}

var Config = &config{
	parser: NewParser("\t", "\n", "title", true),
	tables: make([]IConfigTable, 0),
}

func (c *config) RegisterTable(t interface{}) {
	it := t.(IConfigTable) //panic on error
	it.Init()
	c.tables = append(c.tables, it)
}

func (c *config) RegisterParser(fun interface{}) {
	c.parser.RegisterParser(fun)

}

func LoadAllTable(path string) {
	for _, table := range Config.tables {
		name := table.Name()
		d, err := ReadFile(path + name + ".tsv")
		if err != nil {
			panic("Read file<" + name + "> error" + err.Error())
			return
		}
		fieldIndex := make(map[string]int, 0)
		for idx, line := range strings.Split(d, Config.parser.LinesSplitter) {
			if idx == 0 {
				fieldIndex = Config.parser.ParseFieldIndex(line)
			} else {
				record := table.NewRecord()
				Config.parser.ParseRecord(line, fieldIndex, record)
				table.AppendRecord(record)
			}
		}
	}
}

func DebugDumpAllTable() {
	for _, table := range Config.tables {
		DebugDumpOneTable(table)
	}
}

func DebugDumpOneTable(t IConfigTable) {
	for _, record := range t.debugAllRecords() {
		fmt.Println(Config.parser.FormatRecord(record))
	}
}

func asTimeInt64(s string) (TimeInt64, error) {
	v, err := asInt64(s)
	if err != nil {
		return TimeInt64(0), err
	}
	return TimeInt64(v), nil
}

func asMapStringString(s string) (map[string]string, error) {
	keyValue := make(map[string]string, 0)
	find := regexp.MustCompile(`([a-zA-Z_][a-zA-Z0-9_]+)[ \t]*=[ \t"]*([a-zA-Z0-9_.]+)`).FindAllSubmatch([]byte(s), -1)

	for _, v := range find {
		keyValue[string(v[1])] = string(v[2])
	}

	return keyValue, nil
}

func asMapStringInt(s string) (map[string]int, error) {
	keyValue := make(map[string]int, 0)
	ms, err := asMapStringString(s)
	if err != nil {
		return keyValue, nil
	}
	for key, v := range ms {
		value, err := asInt(v)
		if err != nil {
			return keyValue, err
		}
		keyValue[key] = value
	}

	return keyValue, nil
}

func asMapStringFloat32(s string) (map[string]float32, error) {
	keyValue := make(map[string]float32, 0)
	ms, err := asMapStringString(s)
	if err != nil {
		return keyValue, nil
	}
	for ks, v := range ms {
		value, err := asFloat32(v)
		if err != nil {
			return keyValue, err
		}
		keyValue[ks] = value
	}

	return keyValue, nil
}
