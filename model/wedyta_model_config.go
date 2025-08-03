package model

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type ConfigOfModel struct {
	ModelName           string
	PageTitle           string                            `json:"pageTitle"`
	DbTable             string                            `json:"dbTable"`
	SqlWhereOriginal    string                            `json:"sqlWhere"`
	Fields              []string                          `json:"fields"`
	OrderBy             string                            `json:"orderBy"`
	Headers             map[string]string                 `json:"headers"`
	Titles              map[string]string                 `json:"titles"`
	Classes             map[string]string                 `json:"classes"`
	DisplayMode         map[string]string                 `json:"displayMode"`
	DateTimeFields      map[string]string                 `json:"dateTimeFields"`
	RelatedData         map[string]RelatedDataEntry       `json:"relatedData"`
	AddableFields       []string                          `json:"addableFields"`
	RequiredFields      []string                          `json:"requiredFields"`
	EditableFields      []string                          `json:"editableFields"`
	FieldEditor         map[string]map[string]interface{} `json:"fieldsEditor"`
	NoZeroValueFields   []string                          `json:"noZeroValueFields"`
	ColumnDataFunc      map[string]string                 `json:"columnDataFunc"`
	CountRelatedData    map[string]CountRelatedDataConfig `json:"countRelatedData"`
	Links               map[string]LinkConfig             `json:"links"`
	Parent              ParentConfig                      `json:"parent"`
	Breadcrumb          BreadcrumbConfig
	HasParent           bool
	DbTablePrimaryKey   string
	ParentConfig        *ConfigOfModel
	FieldConfig         map[string]FieldParams
	HeaderTags          string
	AdditionalScripts   string
	SqlWhere            string
	AdditionalUrlParams string
	//InsertModeHiddenFields []string
}

type CachedModelConfig struct {
	Config  *ConfigOfModel
	ModTime time.Time
}

type CountRelatedDataConfig struct {
	LocalFieldID  string `json:"localFieldID"`
	Table         string `json:"table"`
	TargetFieldID string `json:"targetFieldID"`
}

type LinkConfig struct {
	Preset   string `json:"preset"`
	Template string `json:"template"`
}

type FieldParams struct {
	Field                     string
	Header                    string
	Title                     string
	IsAddable                 bool
	IsEditable                bool
	IsRequired                bool
	FieldEditor               string
	Classes                   string
	DisplayMode               string
	PermitDisplayInTableMode  bool
	PermitDisplayInRecordMode bool
	PermitDisplayInUpdateMode bool
	PermitDisplayInInsertMode bool
	//InsertHiddenMode          bool
	RelatedData *RelatedDataEntry
}

type RenderTableCache struct {
	RelatedData map[string]string
}

type BreadcrumbConfig struct {
	LabelField string
}

type ParentConfig struct {
	ModelName            string
	LocalConnectionField string
	QueryVariableName    string
	QueryVariableValue   string
}

type RelatedDataEntry struct {
	Table      string `json:"table"`
	ValueField string `json:"valueField"`
	KeyField   string `json:"keyField,omitempty"`
	OrderBy    string `json:"orderBy,omitempty"`
}

func (r *RelatedDataEntry) UnmarshalJSON(data []byte) error {
	// simple format: "web_users.username"
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		parts := strings.Split(s, ".")
		if len(parts) != 2 {
			return fmt.Errorf("invalid relatedData string format: %s", s)
		}
		r.Table = parts[0]
		r.ValueField = parts[1]
		return nil
	}

	// extended format: object
	type alias RelatedDataEntry
	var tmp alias
	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}
	*r = RelatedDataEntry(tmp)
	return nil
}
