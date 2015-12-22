package config

import (
	"fmt"
	. "types"
)

type CfgBuilding struct {
	Name  string `title:"STR_BuildingName"`
	Level int    `title:"INT_Lvl"`

	BuildTime    TimeInt64 `title:"INT_Time"`
	DestructTime TimeInt64 `title:"INT_DestructTime"`

	Upgradeable bool `title:"BOOL_Upgradable"`
	Destroyable bool `title:"BOOL_Destructable"`

	Wood    int64 `title:"INT_Wood"`
	Ivory   int64 `title:"INT_Ivory"`
	Leather int64 `title:"INT_Leather"`
	Meat    int64 `title:"INT_Meat"`

	Power int64 `title:"INT_Power"`
	Exp   int64 `title:"INT_Xp"`

	Item        map[string]string  `title:"TABLE_Item"`
	PreBuilding map[string]int     `title:"TABLE_Pre"`
	Buffs       map[string]float32 `title:"TABLE_Para"`

	Effects map[string]float32
	Rss     ResourceType
}

type cfgTableBuilding struct {
	Records map[string]*CfgBuilding
}

var cfgBuildingTable *cfgTableBuilding = nil

func (c CfgBuilding) uniqueId() string {
	return fmt.Sprintf("%s_lvl%d", c.Name, c.Level)
}

func (c *CfgBuilding) AfterParse() {
	id := c.uniqueId()

	c.Rss.Ivory = c.Ivory
	c.Rss.Leather = c.Leather
	c.Rss.Mana = 0
	c.Rss.Meat = c.Meat
	c.Rss.Wood = c.Wood
	c.Effects = Effector.RegisterBuff(c.Buffs)

	cfgBuildingTable.Records[id] = c
}

func (t *cfgTableBuilding) Init() {
	cfgBuildingTable = &cfgTableBuilding{
		Records: make(map[string]*CfgBuilding, 0),
	}
}

func (t *cfgTableBuilding) Name() string {
	return "all_buildings"
}

func (t *cfgTableBuilding) NewRecord() interface{} {
	return &CfgBuilding{}
}

func (t *cfgTableBuilding) AppendRecord(data interface{}) {
	record := data.(*CfgBuilding)
	cfgBuildingTable.Records[record.uniqueId()] = record
}

func (t *cfgTableBuilding) debugAllRecords() map[string]interface{} {
	records := make(map[string]interface{}, 0)
	for id, record := range cfgBuildingTable.Records {
		records[id] = record
	}
	return records
}
