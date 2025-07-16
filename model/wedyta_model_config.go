package model

import (
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
	RelatedData         map[string]string                 `json:"relatedData"`
	AddableFields       []string                          `json:"addableFields"`
	RequiredFields      []string                          `json:"requiredFields"`
	EditableFields      []string                          `json:"editableFields"`
	FieldEditor         map[string]string                 `json:"fieldsEditor"`
	NoZeroValueFields   []string                          `json:"noZeroValueFields"`
	ColumnDataFunc      map[string]string                 `json:"columnDataFunc"`
	CountRelatedData    map[string]CountRelatedDataConfig `json:"countRelatedData"`
	Links               map[string]LinkConfig             `json:"links"`
	Parent              map[string]string                 `json:"parent"`
	Breadcrumb          BreadcrumbConfig
	HasParent           bool
	DbTablePrimaryKey   string
	ParentConfig        *ConfigOfModel
	FieldConfig         map[string]FieldParams
	HeaderTags          string
	AdditionalScripts   string
	SqlWhere            string
	AdditionalUrlParams string
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
	RelatedData               *RelatedDataConfig
}

type RenderTableCache struct {
	RelatedData map[string]string
}

type RelatedDataConfig struct {
	TableAndField       string
	TableName           string
	FieldName           string
	PrimaryKeyFieldName string
}

type BreadcrumbConfig struct {
	LabelField string
}
